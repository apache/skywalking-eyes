// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package deps

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/apache/skywalking-eyes/pkg/logger"
)

// GemfileLockResolver resolves Ruby dependencies from Gemfile.lock
// It determines project type by the presence of a *.gemspec file in the same directory as Gemfile.lock.
// - Library projects (with gemspec): ignore development dependencies; include only runtime deps and their transitive closure.
// - App projects (no gemspec): include all dependencies in Gemfile.lock.
// Licenses are fetched from RubyGems API unless overridden by user config.
// See issue description for detailed rules.

type GemfileLockResolver struct {
	Resolver
}

func (r *GemfileLockResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	return base == "Gemfile.lock"
}

func (r *GemfileLockResolver) Resolve(lockfile string, config *ConfigDeps, report *Report) error {
	dir := filepath.Dir(lockfile)

	content, err := os.ReadFile(lockfile)
	if err != nil {
		return err
	}

	// Parse lockfile into specs graph and top-level dependencies
	specs, deps, err := parseGemfileLock(string(content))
	if err != nil {
		return err
	}

	isLibrary := hasGemspec(dir)

	var roots []string
	if isLibrary {
		// Extract runtime dependencies from gemspec(s)
		runtimeRoots, err := runtimeDepsFromGemspecs(dir)
		if err != nil {
			return err
		}
		// Use only runtime dependencies from gemspec(s);
		// Do not fallback to DEPENDENCIES
		// A gem library inherits runtime dependencies ONLY from the gemspec.
		// All other dependencies are development only.
		roots = runtimeRoots
	} else {
		// App: all declared dependencies are relevant
		roots = deps
	}

	// Compute the set of included gems
	include := reachable(specs, roots)
	// For app without explicit deps (rare), include all specs
	if !isLibrary && len(roots) == 0 {
		for name := range specs {
			include[name] = struct{}{}
		}
	}

	// Resolve licenses for included gems
	for name := range include {
		// Some roots may not exist in the specs graph (e.g., git-sourced gems)
		var version string
		var localPath string
		if spec, ok := specs[name]; ok && spec != nil {
			version = spec.Version
			localPath = spec.LocalPath
		}
		if exclude, _ := config.IsExcluded(name, version); exclude {
			continue
		}
		if l, ok := config.GetUserConfiguredLicense(name, version); ok {
			report.Resolve(&Result{Dependency: name, LicenseSpdxID: l, Version: version})
			continue
		}

		if localPath != "" {
			baseDir, err := filepath.Abs(dir)
			if err != nil {
				logger.Log.WithError(err).Warn("failed to resolve base directory for local gem path")
			} else {
				candidatePath := filepath.Clean(filepath.Join(baseDir, localPath))
				if candidatePath == baseDir || strings.HasPrefix(candidatePath, baseDir+string(os.PathSeparator)) {
					fullPath := candidatePath
					license, err := fetchLocalLicense(fullPath, name)
					if err == nil && license != "" {
						report.Resolve(&Result{Dependency: name, LicenseSpdxID: license, Version: version})
						continue
					}
				} else {
					logger.Log.WithField("path", localPath).Warn("ignoring potentially unsafe local gem path outside project directory")
				}
			}
		}

		licenseID := fetchInstalledLicense(name, version)
		var err error
		if licenseID == "" {
			licenseID, err = fetchRubyGemsLicense(name, version)
		}
		if err != nil || licenseID == "" {
			// Gracefully treat as unresolved license and record in report
			report.Skip(&Result{Dependency: name, LicenseSpdxID: Unknown, Version: version})
			continue
		}
		report.Resolve(&Result{Dependency: name, LicenseSpdxID: licenseID, Version: version})
	}

	return nil
}

// GemspecResolver resolves dependencies from a .gemspec file.
// It extracts runtime dependencies defined in the gemspec and recursively resolves
// their transitive dependencies by looking up installed gems in the local environment.
type GemspecResolver struct {
	Resolver
}

// CanResolve checks if the given file is a .gemspec file.
func (r *GemspecResolver) CanResolve(file string) bool {
	return strings.HasSuffix(file, ".gemspec")
}

// Resolve parses the gemspec file, identifies runtime dependencies, and resolves
// them along with their transitive dependencies. It reports the found dependencies
// and their licenses.
func (r *GemspecResolver) Resolve(file string, config *ConfigDeps, report *Report) error {
	deps, err := parseInitialDependencies(file)
	if err != nil {
		return err
	}

	if errResolve := resolveTransitiveDependencies(deps); errResolve != nil {
		return errResolve
	}

	for name, version := range deps {
		if exclude, _ := config.IsExcluded(name, version); exclude {
			continue
		}
		if l, ok := config.GetUserConfiguredLicense(name, version); ok {
			report.Resolve(&Result{Dependency: name, LicenseSpdxID: l, Version: version})
			continue
		}

		// Check installed gems first, then fallback to RubyGems API
		licenseID := fetchInstalledLicense(name, version)
		if licenseID == "" {
			licenseID, err = fetchRubyGemsLicense(name, version)
		}
		if err != nil || licenseID == "" {
			report.Skip(&Result{Dependency: name, LicenseSpdxID: Unknown, Version: version})
			continue
		}
		report.Resolve(&Result{Dependency: name, LicenseSpdxID: licenseID, Version: version})
	}
	return nil
}

func parseInitialDependencies(file string) (map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	deps := make(map[string]string) // name -> version constraint
	for scanner.Scan() {
		line := scanner.Text()
		trimLeft := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimLeft, "#") {
			continue
		}
		if m := gemspecRuntimeRe.FindStringSubmatch(line); len(m) == 2 {
			// NOTE: Version constraints are currently ignored. We resolve to the first found installed version of the gem.
			// This may lead to incorrect resolution if multiple versions are installed and the first one doesn't satisfy the constraint.
			deps[m[1]] = ""
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return deps, nil
}

func resolveTransitiveDependencies(deps map[string]string) error {
	queue := make([]string, 0, len(deps))
	visited := make(map[string]struct{}, len(deps))
	for name := range deps {
		queue = append(queue, name)
		visited[name] = struct{}{}
	}

	for i := 0; i < len(queue); i++ {
		name := queue[i]
		// Find installed gemspec for 'name'
		path, err := findInstalledGemspec(name, "")
		if err != nil {
			logger.Log.Debugf("failed to find installed gemspec for %s: %v", name, err)
			continue
		}
		if path == "" {
			continue
		}

		// Parse dependencies of this gemspec
		newDeps, err := parseGemspecDependencies(path)
		if err != nil {
			logger.Log.Debugf("failed to parse gemspec dependencies for %s at %s: %v", name, path, err)
			continue
		}

		for _, dep := range newDeps {
			if _, ok := visited[dep]; !ok {
				if len(queue) >= 10000 {
					return fmt.Errorf("dependency graph exceeded maximum size of 10000 nodes (current: %d). "+
						"This may indicate a circular dependency or an unusually large dependency tree", len(queue))
				}
				visited[dep] = struct{}{}
				queue = append(queue, dep)
				if _, ok := deps[dep]; !ok {
					deps[dep] = ""
				}
			}
		}
	}
	return nil
}

// -------- Parsing Gemfile.lock --------

type gemSpec struct {
	Name      string
	Version   string
	Deps      []string
	LocalPath string
}

type gemGraph map[string]*gemSpec

var (
	lockSpecHeader = regexp.MustCompile(`^\s{4}([a-zA-Z0-9_\-]+) \(([^)]+)\)`) //     rake (13.0.6)
	lockDepLine    = regexp.MustCompile(`^\s{6}([a-zA-Z0-9_\-]+)(?:\s|$)`)     //       activesupport (~> 6.1)
)

func parseGemfileLock(s string) (graph gemGraph, roots []string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanLines)
	graph = make(gemGraph)

	state := &lockParserState{
		graph: graph,
	}

	for scanner.Scan() {
		state.processLine(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return graph, state.roots, nil
}

type lockParserState struct {
	inSpecs          bool
	inDeps           bool
	inPath           bool
	current          *gemSpec
	currentLocalPath string
	graph            gemGraph
	roots            []string
}

func (s *lockParserState) processLine(line string) {
	if strings.HasPrefix(line, "GEM") {
		s.inSpecs = true
		s.inDeps = false
		s.inPath = false
		s.currentLocalPath = ""
		s.current = nil
		return
	}
	if strings.HasPrefix(line, "PATH") {
		s.inSpecs = true
		s.inDeps = false
		s.inPath = true
		s.currentLocalPath = ""
		s.current = nil
		return
	}
	if strings.HasPrefix(line, "DEPENDENCIES") {
		s.inSpecs = false
		s.inDeps = true
		s.inPath = false
		s.current = nil
		return
	}
	if strings.TrimSpace(line) == "specs:" && s.inSpecs {
		// just a marker
		return
	}

	if s.inSpecs {
		s.processSpecs(line)
		return
	}

	if s.inDeps {
		s.processDeps(line)
		return
	}
}

func (s *lockParserState) processSpecs(line string) {
	trim := strings.TrimSpace(line)
	if strings.HasPrefix(trim, "remote:") {
		// The inPath check ensures that only PATH block remote paths are captured,
		// not GEM block remote URLs (like gem.coop).
		// This distinction is important for proper local dependency resolution.
		if s.inPath {
			s.currentLocalPath = strings.TrimSpace(strings.TrimPrefix(trim, "remote:"))
		}
		return
	}

	if m := lockSpecHeader.FindStringSubmatch(line); len(m) == 3 {
		name := m[1]
		version := m[2]
		s.current = &gemSpec{Name: name, Version: version}
		if s.inPath {
			s.current.LocalPath = s.currentLocalPath
		}
		s.graph[name] = s.current
		return
	}
	if s.current != nil {
		if m := lockDepLine.FindStringSubmatch(line); len(m) == 2 {
			depName := m[1]
			s.current.Deps = append(s.current.Deps, depName)
		}
	}
}

func (s *lockParserState) processDeps(line string) {
	trim := strings.TrimSpace(line)
	if trim == "" || strings.HasPrefix(trim, "BUNDLED WITH") {
		s.inDeps = false
		return
	}
	// dependency line: byebug (~> 11.1)
	root := trim
	if i := strings.Index(root, " "); i >= 0 {
		root = root[:i]
	}
	root = strings.TrimSuffix(root, "!")
	// ignore comments and platforms
	if root != "" && !strings.HasPrefix(root, "#") {
		s.roots = append(s.roots, root)
	}
}

func hasGemspec(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".gemspec") {
			return true
		}
	}
	return false
}

var gemspecRuntimeRe = regexp.MustCompile(`\badd_(?:runtime_)?dependency\s*\(?\s*["']([^"']+)["']`)
var gemspecLicenseRe = regexp.MustCompile(`\.licenses?\s*=\s*(\[[^\]]*\]|['"][^'"]*['"])`)
var gemspecStringRe = regexp.MustCompile(`['"]([^'"]+)['"]`)
var gemspecNameRe = regexp.MustCompile(`\.name\s*=\s*['"]([^'"]+)['"]`)
var rubyVersionRe = regexp.MustCompile(`^\d+(\.[0-9a-zA-Z]+)*(-[0-9a-zA-Z]+)?$`)

func runtimeDepsFromGemspecs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	runtime := make(map[string]struct{})
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".gemspec") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		deps, err := parseGemspecDependencies(path)
		if err != nil {
			return nil, err
		}
		for _, d := range deps {
			runtime[d] = struct{}{}
		}
	}
	res := make([]string, 0, len(runtime))
	for k := range runtime {
		res = append(res, k)
	}
	return res, nil
}

func parseGemspecDependencies(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var deps []string
	for scanner.Scan() {
		line := scanner.Text()
		trimLeft := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimLeft, "#") {
			continue
		}
		if m := gemspecRuntimeRe.FindStringSubmatch(line); len(m) == 2 {
			deps = append(deps, m[1])
		}
	}
	return deps, scanner.Err()
}

var (
	gemspecsCache     map[string][]string
	gemspecsCacheLock sync.Mutex
)

func getAllGemspecs() []string {
	env := os.Getenv("GEM_PATH")
	if env == "" {
		env = os.Getenv("GEM_HOME")
	}

	gemspecsCacheLock.Lock()
	defer gemspecsCacheLock.Unlock()

	if gemspecsCache == nil {
		gemspecsCache = make(map[string][]string)
	}

	if cached, ok := gemspecsCache[env]; ok {
		return cached
	}

	var allGemspecs []string
	gemPaths := getGemPaths()
	for _, dir := range gemPaths {
		specsDir := filepath.Join(dir, "specifications")
		entries, err := os.ReadDir(specsDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".gemspec") {
				allGemspecs = append(allGemspecs, filepath.Join(specsDir, e.Name()))
			}
		}
	}
	gemspecsCache[env] = allGemspecs
	return allGemspecs
}

func findInstalledGemspec(name, version string) (string, error) {
	gems := getAllGemspecs()
	for _, path := range gems {
		filename := filepath.Base(path)
		if version != "" && rubyVersionRe.MatchString(version) {
			if filename == name+"-"+version+".gemspec" {
				return path, nil
			}
		} else {
			if !strings.HasPrefix(filename, name+"-") {
				continue
			}
			stem := strings.TrimSuffix(filename, ".gemspec")
			// Ensure that the character immediately after the "name-" prefix
			// is a digit, so we only consider filenames where the suffix is
			// a version component (e.g., "foo-1.0.0.gemspec") and avoid
			// similar names like "foo-bar-1.0.0.gemspec" when searching for "foo".
			if len(stem) <= len(name)+1 {
				continue
			}
			versionStart := stem[len(name)+1]
			if versionStart < '0' || versionStart > '9' {
				continue
			}

			if specName, _, err := parseGemspecInfo(path); err == nil && specName == name {
				return path, nil
			}
		}
	}
	return "", os.ErrNotExist
}

func fetchLocalLicense(dir, targetName string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".gemspec") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		specName, license, err := parseGemspecInfo(path)
		if err == nil && specName == targetName && license != "" {
			return license, nil
		}
	}
	return "", nil
}

func fetchInstalledLicense(name, version string) string {
	if version != "" && !rubyVersionRe.MatchString(version) {
		return ""
	}
	gems := getAllGemspecs()
	for _, path := range gems {
		filename := filepath.Base(path)
		// If version is specific
		if version != "" && rubyVersionRe.MatchString(version) {
			if filename == name+"-"+version+".gemspec" {
				if _, license, err := parseGemspecInfo(path); err == nil && license != "" {
					return license
				}
			}
		} else {
			// Scan for any version
			if !strings.HasPrefix(filename, name+"-") {
				continue
			}
			stem := strings.TrimSuffix(filename, ".gemspec")
			ver := strings.TrimPrefix(stem, name+"-")
			// Ensure the character after the gem name corresponds to the start of a version
			if ver == "" || ver[0] < '0' || ver[0] > '9' {
				continue
			}
			if specName, license, err := parseGemspecInfo(path); err == nil && specName == name && license != "" {
				return license
			}
		}
	}
	return ""
}

func getGemPaths() []string {
	env := os.Getenv("GEM_PATH")
	if env == "" {
		env = os.Getenv("GEM_HOME")
	}
	if env == "" {
		return nil
	}
	return strings.Split(env, string(os.PathListSeparator))
}

func parseGemspecInfo(path string) (gemName, gemLicense string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var name, license string
	for scanner.Scan() {
		line := scanner.Text()
		trimLeft := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimLeft, "#") {
			continue
		}
		if name == "" {
			if m := gemspecNameRe.FindStringSubmatch(line); len(m) == 2 {
				name = m[1]
			}
		}
		if license == "" {
			if m := gemspecLicenseRe.FindStringSubmatch(line); len(m) == 2 {
				matches := gemspecStringRe.FindAllStringSubmatch(m[1], -1)
				var licenses []string
				for _, match := range matches {
					if len(match) == 2 {
						licenses = append(licenses, match[1])
					}
				}
				if len(licenses) > 0 {
					// NOTE: When multiple licenses are declared in the gemspec, we assume they are
					// all required ("AND") to be conservative. If the author intended "OR",
					// the user can override this in the configuration.
					if len(licenses) > 1 {
						gemRef := name
						if gemRef == "" {
							gemRef = path
						}
						logger.Log.Warnf("Multiple licenses found for gem %s: %v. Assuming 'AND' relationship for safety.", gemRef, licenses)
					}
					license = strings.Join(licenses, " AND ")
				}
			}
		}
		if name != "" && license != "" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}
	return name, license, nil
}

func reachable(graph gemGraph, roots []string) map[string]struct{} {
	vis := make(map[string]struct{})
	var dfs func(string)
	dfs = func(n string) {
		if _, ok := vis[n]; ok {
			return
		}
		if _, ok := graph[n]; !ok {
			// unknown in specs, still include the root
			vis[n] = struct{}{}
			return
		}
		vis[n] = struct{}{}
		for _, c := range graph[n].Deps {
			dfs(c)
		}
	}
	for _, r := range roots {
		dfs(r)
	}
	return vis
}

// -------- License resolution via RubyGems API --------

type rubyGemsVersionInfo struct {
	Licenses []string `json:"licenses"`
	License  string   `json:"license"`
}

func fetchRubyGemsLicense(name, version string) (string, error) {
	// If version is unknown (e.g., git-sourced), query latest gem info endpoint
	if strings.TrimSpace(version) == "" {
		url := fmt.Sprintf("https://gem.coop/api/v1/gems/%s.json", name)
		return fetchRubyGemsLicenseFrom(url)
	}
	// Prefer version-specific API
	url := fmt.Sprintf("https://gem.coop/api/v2/rubygems/%s/versions/%s.json", name, version)
	licenseID, err := fetchRubyGemsLicenseFrom(url)
	if err == nil && licenseID != "" {
		return licenseID, nil
	}
	// Fallback to latest info
	url = fmt.Sprintf("https://gem.coop/api/v1/gems/%s.json", name)
	return fetchRubyGemsLicenseFrom(url)
}

var httpClientRuby = &http.Client{Timeout: 10 * time.Second}

func fetchRubyGemsLicenseFrom(url string) (string, error) {
	const maxAttempts = 3
	backoff := 1 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "skywalking-eyes/License-Eye (+https://github.com/apache/skywalking-eyes)")
		req.Header.Set("Accept", "application/json")

		resp, err := httpClientRuby.Do(req) // #nosec G107
		if err != nil {
			if attempt == maxAttempts {
				return "", err
			}
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		license, wait, retry, hErr := handleRubyGemsResponse(resp)
		_ = resp.Body.Close()

		if hErr != nil {
			if retry && attempt < maxAttempts {
				if wait > 0 {
					time.Sleep(wait)
				} else {
					time.Sleep(backoff)
				}
				backoff *= 2
				continue
			}
			return "", hErr
		}

		if retry { // safety branch, normally handled above when hErr != nil
			if attempt == maxAttempts {
				return "", fmt.Errorf("max attempts reached")
			}
			if wait > 0 {
				time.Sleep(wait)
			} else {
				time.Sleep(backoff)
			}
			backoff *= 2
			continue
		}

		return license, nil
	}
	return "", nil
}

func handleRubyGemsResponse(resp *http.Response) (license string, wait time.Duration, retry bool, err error) {
	switch {
	case resp.StatusCode == http.StatusOK:
		license, err := parseRubyGemsLicenseJSON(resp.Body)
		return license, 0, false, err
	case resp.StatusCode == http.StatusNotFound:
		// Treat as no license info available
		return "", 0, false, nil
	case resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode <= 599):
		wait := retryAfterDuration(resp.Header.Get("Retry-After"))
		return "", wait, true, fmt.Errorf("retryable status: %s", resp.Status)
	default:
		return "", 0, false, fmt.Errorf("unexpected status: %s", resp.Status)
	}
}

func parseRubyGemsLicenseJSON(r io.Reader) (string, error) {
	var info rubyGemsVersionInfo
	dec := json.NewDecoder(r)
	if err := dec.Decode(&info); err != nil {
		return "", err
	}
	var items []string
	if len(info.Licenses) > 0 {
		items = info.Licenses
	} else if info.License != "" {
		items = []string{info.License}
	}
	for i := range items {
		items[i] = strings.TrimSpace(items[i])
	}
	m := make(map[string]struct{})
	for _, it := range items {
		if it == "" {
			continue
		}
		m[it] = struct{}{}
	}
	if len(m) == 0 {
		return "", nil
	}
	var out []string
	for k := range m {
		out = append(out, k)
	}
	slicesSort(out)
	return strings.Join(out, " OR "), nil
}

func retryAfterDuration(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		wait := time.Duration(secs) * time.Second
		if wait > 10*time.Second {
			wait = 10 * time.Second
		}
		return wait
	}
	return 0
}

// small helper to sort string slice without importing sort here to keep imports aligned with style used in this package
func slicesSort(ss []string) {
	// simple insertion sort for small slices
	for i := 1; i < len(ss); i++ {
		j := i
		for j > 0 && ss[j-1] > ss[j] {
			ss[j-1], ss[j] = ss[j], ss[j-1]
			j--
		}
	}
}

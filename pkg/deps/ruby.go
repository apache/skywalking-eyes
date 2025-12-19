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
	"time"
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
			fullPath := filepath.Join(dir, localPath)
			license, err := fetchLocalLicense(fullPath, name)
			if err == nil && license != "" {
				report.Resolve(&Result{Dependency: name, LicenseSpdxID: license, Version: version})
				continue
			}
		}

		licenseID, err := fetchRubyGemsLicense(name, version)
		if err != nil || licenseID == "" {
			// Gracefully treat as unresolved license and record in report
			report.Skip(&Result{Dependency: name, LicenseSpdxID: Unknown, Version: version})
			continue
		}
		report.Resolve(&Result{Dependency: name, LicenseSpdxID: licenseID, Version: version})
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

	inSpecs := false
	inDeps := false
	inPath := false
	var current *gemSpec
	var currentRemotePath string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "GEM") {
			inSpecs = true
			inDeps = false
			inPath = false
			current = nil
			continue
		}
		if strings.HasPrefix(line, "PATH") {
			inSpecs = true
			inDeps = false
			inPath = true
			current = nil
			continue
		}
		if strings.HasPrefix(line, "DEPENDENCIES") {
			inSpecs = false
			inDeps = true
			inPath = false
			current = nil
			continue
		}
		if strings.TrimSpace(line) == "specs:" && inSpecs {
			// just a marker
			continue
		}

		if inSpecs {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "remote:") {
				if inPath {
					currentRemotePath = strings.TrimSpace(strings.TrimPrefix(trim, "remote:"))
				}
				continue
			}

			if m := lockSpecHeader.FindStringSubmatch(line); len(m) == 3 {
				name := m[1]
				version := m[2]
				current = &gemSpec{Name: name, Version: version}
				if inPath {
					current.LocalPath = currentRemotePath
				}
				graph[name] = current
				continue
			}
			if current != nil {
				if m := lockDepLine.FindStringSubmatch(line); len(m) == 2 {
					depName := m[1]
					current.Deps = append(current.Deps, depName)
				}
			}
			continue
		}

		if inDeps {
			trim := strings.TrimSpace(line)
			if trim == "" || strings.HasPrefix(trim, "BUNDLED WITH") {
				inDeps = false
				continue
			}
			// dependency line: byebug (~> 11.1)
			root := trim
			if i := strings.Index(root, " "); i >= 0 {
				root = root[:i]
			}
			root = strings.TrimSuffix(root, "!")
			// ignore comments and platforms
			if root != "" && !strings.HasPrefix(root, "#") {
				roots = append(roots, root)
			}
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return graph, roots, nil
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
var gemspecLicenseRe = regexp.MustCompile(`\.licenses?\s*=\s*(?:\[\s*)?['"]([^'"]+)['"]`)

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
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			trimLeft := strings.TrimLeft(line, " \t")
			if strings.HasPrefix(trimLeft, "#") {
				continue // ignore commented-out lines
			}
			if m := gemspecRuntimeRe.FindStringSubmatch(line); len(m) == 2 {
				runtime[m[1]] = struct{}{}
			}
		}
		_ = f.Close()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	res := make([]string, 0, len(runtime))
	for k := range runtime {
		res = append(res, k)
	}
	return res, nil
}

func fetchLocalLicense(dir, name string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".gemspec") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		scanner := bufio.NewScanner(f)
		var license string
		for scanner.Scan() {
			line := scanner.Text()
			trimLeft := strings.TrimLeft(line, " \t")
			if strings.HasPrefix(trimLeft, "#") {
				continue
			}
			if m := gemspecLicenseRe.FindStringSubmatch(line); len(m) == 2 {
				license = m[1]
				break
			}
		}
		_ = f.Close()
		if license != "" {
			return license, nil
		}
	}
	return "", nil
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
		url := fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", name)
		return fetchRubyGemsLicenseFrom(url)
	}
	// Prefer version-specific API
	url := fmt.Sprintf("https://rubygems.org/api/v2/rubygems/%s/versions/%s.json", name, version)
	licenseID, err := fetchRubyGemsLicenseFrom(url)
	if err == nil && licenseID != "" {
		return licenseID, nil
	}
	// Fallback to latest info
	url = fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", name)
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

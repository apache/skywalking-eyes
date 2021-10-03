//
// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
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
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/internal/logger"
)

type MavenPomResolver struct {
	JarResolver
	mvn *mavenTool
}

// CanResolve determine whether the file can be resolve by name of the file
func (resolver *MavenPomResolver) CanResolve(mavenPomFile string) bool {
	base := filepath.Base(mavenPomFile)
	logger.Log.Debugln("Base name:", base)
	return base == "pom.xml"
}

// Resolve resolves licenses of all dependencies declared in the pom.xml file.
func (resolver *MavenPomResolver) Resolve(mavenPomFile string, report *Report) error {
	if err := resolver.prepare(mavenPomFile); err != nil {
		return err
	}

	deps, err := resolver.mvn.LoadDependencies()
	if err != nil {
		return err
	}

	return resolver.ResolveDependencies(deps, report)
}

func (resolver *MavenPomResolver) prepare(mavenPomFile string) error {
	if err := os.Chdir(filepath.Dir(mavenPomFile)); err != nil {
		return err
	}

	mvn := findMaven()
	if mvn == nil {
		return errors.New("not found maven tool")
	}

	resolver.mvn = mvn
	return nil
}

// ResolveDependencies resolves the licenses of the given dependencies
func (resolver *MavenPomResolver) ResolveDependencies(deps []*DependencyWrapper, report *Report) error {
	for _, dep := range deps {
		state := NotFound
		err := resolver.ResolveLicense(&state, dep, report)
		if err != nil {
			logger.Log.Warnf("Failed to resolve the license of <%s>: %v\n", dep.Jar(), state.String())
			report.Skip(&Result{
				Dependency:    dep.DepName(),
				LicenseSpdxID: Unknown,
			})
		}
	}
	return nil
}

var (
	mavenType = [...]string{"./mvnw", "mvn"}
)

// findMaven find available maven tools and local repositories
func findMaven() *mavenTool {
	mvn := new(mavenTool)
	for _, maven := range mavenType {
		if _, err := exec.Command(maven, "--version").Output(); err == nil {
			mvn.exe = maven
			logger.Log.Debugf("use %s as default maven tool", filepath.Base(maven))
			break
		}
	}

	if mvn.exe == "" {
		return nil
	}

	output, err := exec.Command(mvn.exe, "help:evaluate", "-Dexpression=settings.localRepository", "-q", "-DforceStdout").Output() // #nosec G204
	if err != nil {
		logger.Log.Debugln(err)
		return nil
	}

	mvn.repo = string(output)
	return mvn
}

type mavenTool struct {
	exe  string
	repo string
}

func (mvn *mavenTool) LoadDependencies() (deps []*DependencyWrapper, err error) {
	deps, err = mvn.LoadDependenciesFromProject()
	if err == nil {
		return deps, nil
	}

	deps, err = mvn.LoadDependenciesAfterDownload()
	if err == nil {
		return deps, nil
	}

	deps, err = mvn.LoadDependenciesAfterInstall()
	if err == nil {
		return deps, nil
	}

	return nil, errors.New("failed to resolve dependencies")
}

func (mvn *mavenTool) downloadDeps() error {
	cmd := exec.Command(mvn.exe, "dependency:resolve") // #nosec G204
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (mvn *mavenTool) install() error {
	cmd := exec.Command(mvn.exe, "clean", "install", "-Dcheckstyle.skip=true", "-Drat.skip=true", "-Dmaven.test.skip=true") // #nosec G204
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (mvn *mavenTool) loadDependencies(params ...string) ([]*DependencyWrapper, error) {
	buf := bytes.NewBuffer(nil)

	var cmd *exec.Cmd
	if len(params) == 1 {
		cmd = exec.Command(mvn.exe, "dependency:tree", "-f", params[0]) // #nosec G204
	} else {
		cmd = exec.Command(mvn.exe, "dependency:tree") // #nosec G204
	}
	cmd.Stdout = bufio.NewWriter(buf)
	cmd.Stderr = os.Stderr

	logger.Log.Debugf("Run command: 「%v」, please wait", cmd.String())
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	deps := LoadDependencies(buf.Bytes())
	return deps, nil
}

func (mvn *mavenTool) LoadDependenciesFromProject() ([]*DependencyWrapper, error) {
	project, err := mvn.NewProject()
	if err != nil {
		return nil, err
	}

	return project.LoadDependencies()
}

// LoadDependenciesAfterDownload download and then load dependencies
func (mvn *mavenTool) LoadDependenciesAfterDownload() ([]*DependencyWrapper, error) {
	if err := mvn.downloadDeps(); err != nil {
		return nil, err
	}

	return mvn.loadDependencies()
}

// LoadDependenciesAfterInstall install and then load dependencies
func (mvn *mavenTool) LoadDependenciesAfterInstall() ([]*DependencyWrapper, error) {
	if err := mvn.install(); err != nil {
		return nil, err
	}

	return mvn.loadDependencies()
}

// ResolveLicense search all possible locations of the license, such as pom file, jar package
func (resolver *MavenPomResolver) ResolveLicense(state *State, dep *DependencyWrapper, report *Report) error {
	err := resolver.ResolveJar(state, filepath.Join(resolver.mvn.repo, dep.Path(), dep.Jar()), report)
	if err == nil {
		return nil
	}

	return resolver.ResolveLicenseFromPom(state, dep, report)
}

// ResolveLicenseFromPom search for license in the pom file, which may appear in the header comments or in license element of xml
func (resolver *MavenPomResolver) ResolveLicenseFromPom(state *State, dep *DependencyWrapper, report *Report) (err error) {
	pomFile := filepath.Join(resolver.mvn.repo, dep.Path(), dep.Pom())

	pom, err := resolver.ReadLicensesFromPom(pomFile)
	if err != nil {
		return err
	} else if pom != nil && pom.HaveLicenses() {
		report.Resolve(&Result{
			Dependency:      dep.DepName(),
			LicenseFilePath: pomFile,
			LicenseContent:  pom.Raw(),
			LicenseSpdxID:   pom.AllLicenses(),
		})

		return nil
	}

	headerComments := pom.HeaderComment
	if headerComments != "" {
		*state |= FoundLicenseInPomHeader
		return resolver.IdentifyLicense(pomFile, dep.Jar(), headerComments, report)
	}

	return fmt.Errorf("not found in pom file")
}

func (resolver *MavenPomResolver) ReadLicensesFromPom(pomFile string) (*pomFileWrapper, error) {
	return newPomfile(filepath.Split(pomFile))
}

var (
	reMaybeLicense                = regexp.MustCompile(`(?i)licen[sc]e|copyright|copying`)
	reHaveManifestFile            = regexp.MustCompile(`(?i)^(\S*/)?manifest\.MF$`)
	reSearchLicenseInManifestFile = regexp.MustCompile(`(?im)^.*?licen[cs]e.*?(http.+)`)
	skipScppe                     = []string{"test", "provided", "system"}
)

// SeemLicense determine whether the content of the file may be a license file
func SeemLicense(content string) bool {
	return reMaybeLicense.MatchString(content)
}

func LoadDependencies(data []byte) []*DependencyWrapper {
	allDeps := []*DependencyWrapper{}

	depTrees := LoadDependenciesTree(data)

	visited := make(map[string]bool)
	allDepTrees := make(map[string]bool)
	for _, depTree := range depTrees {
		visited[depTree.Path()] = true
		allDepTrees[depTree.Path()] = true
	}

	for _, depTree := range depTrees {
		deps := depTree.Flatten(allDepTrees)
		for _, dep := range deps {
			if visited[dep.Path()] {
				continue
			}
			visited[dep.Path()] = true
			allDeps = append(allDeps, dep)
		}
	}

	return allDeps
}

func LoadDependenciesTree(data []byte) []*DependencyWrapper {
	type Elem struct {
		*DependencyWrapper
		level int
	}

	stack := []Elem{}

	reFind := regexp.MustCompile(`(?im)^\[info\] (([| ]*)(\+-|\\-) )?(.+?):(.+?):(.+?):(.+?)(:(.+?))?( \(.+)?$`)
	rawDeps := reFind.FindAllSubmatch(data, -1)

	if bytes.Contains(rawDeps[len(rawDeps)-1][0], []byte("Finished at")) {
		rawDeps = rawDeps[:len(rawDeps)-1]
	}

	deps := []*DependencyWrapper{}
	for _, rawDep := range rawDeps {
		dep := &DependencyWrapper{
			Dependency: Dependency{
				GroupID:    string(rawDep[4]),
				ArtifactID: string(rawDep[5]),
				Type:       string(rawDep[6]),
				Version:    string(rawDep[7]),
			},
		}

		if len(rawDep[1]) == 0 {
			deps = append(deps, dep)
			if len(stack) != 0 {
				stack = stack[:0]
			}
			stack = append(stack, Elem{dep, 0})
			continue
		}

		dep.Scope = string(rawDep[9])

		level := len(rawDep[2])/3 + 1
		dependence := string(rawDep[3])

		tail := stack[len(stack)-1]

		if level == tail.level {
			stack[len(stack)-1] = Elem{dep, level}
			stack[len(stack)-2].TransitiveDeps = append(stack[len(stack)-2].TransitiveDeps, dep)
		} else {
			stack = append(stack, Elem{dep, level})
			tail.TransitiveDeps = append(tail.TransitiveDeps, dep)
		}

		if dependence == `\-` {
			stack = stack[:len(stack)-1]
		}
	}
	return deps
}

const (
	FoundLicenseInPomHeader State = 1 << iota
	FoundLicenseInJarLicenseFile
	FoundLicenseInJarManifestFile
	NotFound State = 0
)

type State int

func (s *State) String() string {
	if *s == 0 {
		return "no possible license found"
	}

	var m []string

	if *s&FoundLicenseInPomHeader != 0 {
		m = append(m, "failed to resolve license found in pom header")
	}
	if *s&FoundLicenseInJarLicenseFile != 0 {
		m = append(m, "failed to resolve license file found in jar")
	}
	if *s&FoundLicenseInJarManifestFile != 0 {
		m = append(m, "failed to resolve license content from manifest file found in jar")
	}

	return strings.Join(m, "｜")
}

type DependencyWrapper struct {
	Dependency
	TransitiveDeps []*DependencyWrapper
}

func (dep *DependencyWrapper) Clone() *DependencyWrapper {
	return &DependencyWrapper{
		Dependency: dep.Dependency,
	}
}

func (dep *DependencyWrapper) Count() int {
	cnt := 1
	for _, tDep := range dep.TransitiveDeps {
		cnt += tDep.Count()
	}
	return cnt
}

func (dep *DependencyWrapper) DepName() string {
	return fmt.Sprintf("%v/%v/%v", dep.GroupID, dep.ArtifactID, dep.Version)
}

func (dep *DependencyWrapper) Path() string {
	gid := strings.Split(dep.GroupID, ".")
	return fmt.Sprintf("%v/%v/%v", strings.Join(gid, "/"), dep.ArtifactID, dep.Version)
}

func (dep *DependencyWrapper) Pom() string {
	return fmt.Sprintf("%v-%v.pom", dep.ArtifactID, dep.Version)
}

func (dep *DependencyWrapper) Jar() string {
	return fmt.Sprintf("%v-%v.jar", dep.ArtifactID, dep.Version)
}

func (dep *DependencyWrapper) Flatten(skip map[string]bool) []*DependencyWrapper {
	deps := make([]*DependencyWrapper, 0, dep.Count())

	visited := make(map[string]bool)
	queue := []*DependencyWrapper{}
	queue = append(queue, dep.TransitiveDeps...)
loop:
	for len(queue) > 0 {
		d := queue[0]
		queue = queue[1:]

		if skip[d.Path()] {
			continue
		}

		if visited[d.Path()] {
			continue
		}

		for _, skip := range skipScppe {
			if d.Scope == skip {
				continue loop
			}
		}

		visited[d.Path()] = true
		deps = append(deps, d.Clone())
		queue = append(queue, d.TransitiveDeps...)
	}

	return deps
}

func (dep *DependencyWrapper) String() string {
	buf := bytes.NewBuffer(nil)
	w := bufio.NewWriter(buf)

	_, err := w.WriteString(fmt.Sprintf("%v -> %v : %v %v", dep.GroupID, dep.ArtifactID, dep.Version, dep.Scope))
	if err != nil {
		logger.Log.Error(err)
	}

	for _, tDep := range dep.TransitiveDeps {
		_, err = w.WriteString(fmt.Sprintf("\n\t%v", tDep))
		if err != nil {
			logger.Log.Error(err)
		}
	}

	_ = w.Flush()
	return buf.String()
}

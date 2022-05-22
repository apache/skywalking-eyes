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
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html/charset"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/license"
)

type MavenPomResolver struct {
	JarResolver
	maven string
	repo  string
}

// CanResolve determine whether the file can be resolve by name of the file
func (resolver *MavenPomResolver) CanResolve(mavenPomFile string) bool {
	base := filepath.Base(mavenPomFile)
	logger.Log.Debugln("Base name:", base)
	return base == "pom.xml"
}

// Resolve resolves licenses of all dependencies declared in the pom.xml file.
func (resolver *MavenPomResolver) Resolve(mavenPomFile string, config *ConfigDeps, report *Report) error {
	if err := os.Chdir(filepath.Dir(mavenPomFile)); err != nil {
		return err
	}

	if err := resolver.CheckMVN(); err != nil {
		return err
	}

	deps, err := resolver.LoadDependencies()
	if err != nil {
		// attempt to download dependencies
		if err = resolver.DownloadDeps(); err != nil {
			return fmt.Errorf("dependencies download error")
		}
		// load again
		deps, err = resolver.LoadDependencies()
		if err != nil {
			return err
		}
	}

	return resolver.ResolveDependencies(deps, config, report)
}

// CheckMVN check available maven tools, find local repositories and download all dependencies
func (resolver *MavenPomResolver) CheckMVN() error {
	if err := resolver.FindMaven("./mvnw"); err == nil {
		logger.Log.Debugln("mvnw is found, will use mvnw by default")
	} else if err := resolver.FindMaven("mvn"); err != nil {
		return fmt.Errorf("neither found mvnw nor mvn")
	}

	if err := resolver.FindLocalRepository(); err != nil {
		return fmt.Errorf("can not find the local repository: %v", err)
	}

	return nil
}

func (resolver *MavenPomResolver) FindMaven(execName string) error {
	if _, err := exec.Command(execName, "--version").Output(); err != nil {
		return err
	}

	resolver.maven = execName
	return nil
}

func (resolver *MavenPomResolver) FindLocalRepository() error {
	output, err := exec.Command(resolver.maven, "help:evaluate", "-Dexpression=settings.localRepository", "-q", "-DforceStdout").Output() // #nosec G204
	if err != nil {
		return err
	}

	resolver.repo = string(output)
	return nil
}

func (resolver *MavenPomResolver) DownloadDeps() error {
	cmd := exec.Command(resolver.maven, "dependency:resolve") // #nosec G204
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}
	// the failure may be caused by the lack of sub modules, try to install it
	install := exec.Command(resolver.maven, "clean", "install", "-Dcheckstyle.skip=true", "-Drat.skip=true", "-Dmaven.test.skip=true") // #nosec G204
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr

	return install.Run()
}

func (resolver *MavenPomResolver) LoadDependencies() ([]*Dependency, error) {
	buf := bytes.NewBuffer(nil)

	cmd := exec.Command(resolver.maven, "dependency:tree") // #nosec G204
	cmd.Stdout = bufio.NewWriter(buf)
	cmd.Stderr = os.Stderr

	logger.Log.Debugf("Running command: [%v], please wait", cmd.String())
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	deps := LoadDependencies(buf.Bytes())
	return deps, nil
}

// ResolveDependencies resolves the licenses of the given dependencies
func (resolver *MavenPomResolver) ResolveDependencies(deps []*Dependency, config *ConfigDeps, report *Report) error {
	for _, dep := range deps {
		func() {
			for _, l := range config.Licenses {
				for _, version := range strings.Split(l.Version, ",") {
					if l.Name == fmt.Sprintf("%s:%s", strings.Join(dep.GroupID, "."), dep.ArtifactID) && version == dep.Version {
						report.Resolve(&Result{
							Dependency:    dep.Jar(),
							LicenseSpdxID: l.License,
							Version:       dep.Version,
						})
						return
					}
				}
			}
			state := NotFound
			err := resolver.ResolveLicense(config, &state, dep, report)
			if err != nil {
				logger.Log.Warnf("Failed to resolve the license of <%s>: %v\n", dep.Jar(), state.String())
				report.Skip(&Result{
					Dependency:    dep.Jar(),
					LicenseSpdxID: Unknown,
					Version:       dep.Version,
				})
			}
		}()
	}
	return nil
}

// ResolveLicense search all possible locations of the license, such as pom file, jar package
func (resolver *MavenPomResolver) ResolveLicense(config *ConfigDeps, state *State, dep *Dependency, report *Report) error {
	err := resolver.ResolveJar(config, state, filepath.Join(resolver.repo, dep.Path(), dep.Jar()), dep.Version, report)
	if err == nil {
		return nil
	}

	return resolver.ResolveLicenseFromPom(config, state, dep, report)
}

// ResolveLicenseFromPom search for license in the pom file, which may appear in the header comments or in license element of xml
func (resolver *MavenPomResolver) ResolveLicenseFromPom(config *ConfigDeps, state *State, dep *Dependency, report *Report) (err error) {
	pomFile := filepath.Join(resolver.repo, dep.Path(), dep.Pom())

	pom, err := resolver.ReadLicensesFromPom(pomFile)
	if err != nil {
		return err
	}

	if pom != nil && len(pom.Licenses) != 0 {
		report.Resolve(&Result{
			Dependency:      dep.Jar(),
			LicenseFilePath: pomFile,
			LicenseContent:  pom.Raw(),
			LicenseSpdxID:   pom.AllLicenses(config),
			Version:         dep.Version,
		})

		return nil
	}

	headerComments, err := resolver.ReadHeaderCommentsFromPom(pomFile)
	if err != nil {
		return err
	} else if headerComments != "" {
		*state |= FoundLicenseInPomHeader
		return resolver.IdentifyLicense(config, pomFile, dep.Jar(), headerComments, dep.Version, report)
	}

	return fmt.Errorf("not found in pom file")
}

func (resolver *MavenPomResolver) ReadLicensesFromPom(pomFile string) (*PomFile, error) {
	file, err := os.Open(pomFile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	dec := xml.NewDecoder(file)
	dec.CharsetReader = charset.NewReaderLabel

	pom := new(PomFile)
	err = dec.Decode(pom)
	if err != nil {
		return nil, err
	}

	return pom, nil
}

func (resolver *MavenPomResolver) ReadHeaderCommentsFromPom(pomFile string) (string, error) {
	file, err := os.Open(pomFile)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	var comments string

	dec := xml.NewDecoder(file)
	dec.CharsetReader = charset.NewReaderLabel
loop:
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		switch tok := tok.(type) {
		// search header only
		case xml.StartElement:
			break loop
		case xml.Comment:
			comments += string(tok.Copy())
		}
	}

	if SeemLicense(comments) {
		return comments, nil
	}

	return "", nil
}

var (
	reMaybeLicense                = regexp.MustCompile(`(?i)licen[sc]e|copyright|copying$`)
	reHaveManifestFile            = regexp.MustCompile(`(?i)^(\S*/)?manifest\.MF$`)
	reSearchLicenseInManifestFile = regexp.MustCompile(`(?im)^.*?licen[cs]e.*?(http.+)`)
)

// SeemLicense determine whether the content of the file may be a license file
func SeemLicense(content string) bool {
	return reMaybeLicense.MatchString(content)
}

func LoadDependencies(data []byte) []*Dependency {
	depsTree := LoadDependenciesTree(data)

	cnt := 0
	for _, dep := range depsTree {
		cnt += dep.Count()
	}

	deps := make([]*Dependency, 0, cnt)

	queue := []*Dependency{}
	for _, depTree := range depsTree {
		queue = append(queue, depTree)
		for len(queue) > 0 {
			dep := queue[0]

			queue = queue[1:]
			queue = append(queue, dep.TransitiveDeps...)

			deps = append(deps, dep.Clone())
		}
	}
	return deps
}

func LoadDependenciesTree(data []byte) []*Dependency {
	type Elem struct {
		*Dependency
		level int
	}

	stack := []Elem{}
	unique := make(map[string]struct{})

	reFind := regexp.MustCompile(`(?im)^.*? ([| ]*)(\+-|\\-) (?P<gid>\b.+?):(?P<aid>\b.+?):(?P<packaging>\b.+)(:\b.+)?:(?P<version>\b.+):(?P<scope>\b.+?)$`) //nolint:lll // can't break down regex
	rawDeps := reFind.FindAllSubmatch(data, -1)

	deps := make([]*Dependency, 0, len(rawDeps))
	for _, rawDep := range rawDeps {
		gid := strings.Split(string(rawDep[reFind.SubexpIndex("gid")]), ".")
		dep := &Dependency{
			GroupID:    gid,
			ArtifactID: string(rawDep[reFind.SubexpIndex("aid")]),
			Packaging:  string(rawDep[reFind.SubexpIndex("packaging")]),
			Version:    string(rawDep[reFind.SubexpIndex("version")]),
			Scope:      string(rawDep[reFind.SubexpIndex("scope")]),
		}

		if _, have := unique[dep.Path()]; have {
			continue
		}

		if dep.Scope == "test" || dep.Scope == "provided" || dep.Scope == "system" {
			continue
		}

		unique[dep.Path()] = struct{}{}

		level := len(rawDep[1]) / 3
		dependence := string(rawDep[2])

		if level == 0 {
			deps = append(deps, dep)

			if len(stack) != 0 {
				stack = stack[:0]
			}

			stack = append(stack, Elem{dep, level})
			continue
		}

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

	return strings.Join(m, " | ")
}

type Dependency struct {
	GroupID                               []string
	ArtifactID, Version, Packaging, Scope string
	TransitiveDeps                        []*Dependency
}

func (dep *Dependency) Clone() *Dependency {
	return &Dependency{
		GroupID:    dep.GroupID,
		ArtifactID: dep.ArtifactID,
		Version:    dep.Version,
		Packaging:  dep.Packaging,
		Scope:      dep.Scope,
	}
}

func (dep *Dependency) Count() int {
	cnt := 1
	for _, tDep := range dep.TransitiveDeps {
		cnt += tDep.Count()
	}
	return cnt
}

func (dep *Dependency) Path() string {
	return fmt.Sprintf("%v/%v/%v", strings.Join(dep.GroupID, "/"), dep.ArtifactID, dep.Version)
}

func (dep *Dependency) Pom() string {
	return fmt.Sprintf("%v-%v.pom", dep.ArtifactID, dep.Version)
}

func (dep *Dependency) Jar() string {
	return fmt.Sprintf("%v-%v.jar", dep.ArtifactID, dep.Version)
}

func (dep *Dependency) String() string {
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

// PomFile is used to extract license from the pom.xml file
type PomFile struct {
	XMLName  xml.Name      `xml:"project"`
	Licenses []*XMLLicense `xml:"licenses>license,omitempty"`
}

// AllLicenses return all licenses found in pom.xml file
func (pom *PomFile) AllLicenses(config *ConfigDeps) string {
	licenses := []string{}
	for _, l := range pom.Licenses {
		licenses = append(licenses, l.Item(config))
	}
	return strings.Join(licenses, " and ")
}

// Raw return raw data
func (pom *PomFile) Raw() string {
	contents := []string{}
	for _, l := range pom.Licenses {
		contents = append(contents, l.Raw())
	}
	return strings.Join(contents, "\n")
}

type XMLLicense struct {
	Name         string `xml:"name,omitempty"`
	URL          string `xml:"url,omitempty"`
	Distribution string `xml:"distribution,omitempty"`
	Comments     string `xml:"comments,omitempty"`
}

func (l *XMLLicense) Item(config *ConfigDeps) string {
	if l.URL != "" {
		return GetLicenseFromURL(l.URL, config)
	}
	if l.Name != "" {
		return l.Name
	}
	return l.URL
}

func (l *XMLLicense) Raw() string {
	return fmt.Sprintf(`License: {Name: %s, URL: %s, Distribution: %s, Comments: %s, }`, l.Name, l.URL, l.Distribution, l.Comments)
}

func GetLicenseFromURL(url string, config *ConfigDeps) string {
	if l, err := license.Identify(url, config.Threshold); err == nil {
		return l
	}
	return url
}

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
package deps_test

import (
	"bufio"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apache/skywalking-eyes/license-eye/pkg/deps"
)

func TestCanResolve(t *testing.T) {
	resolver := new(deps.MavenPomResolver)
	for _, test := range []struct {
		fileName string
		exp      bool
	}{
		{"pom.xml", true},
		{"POM.XML", true},
		{"log4j-1.2.12.pom", true},
		{".pom", false},
	} {
		b := resolver.CanResolve(test.fileName)
		if b != test.exp {
			t.Errorf("MavenPomResolver.CanResolve(\"%v\") = %v, want %v", test.fileName, b, test.exp)
		}
	}
}

type PomFile struct {
	XMLName        xml.Name     `xml:"project"`
	NameSpace      string       `xml:"xmlns,attr"`
	SchemeInstance string       `xml:"xmlns:xsi,attr"`
	SchemaLocation string       `xml:"xsi:schemaLocation,attr"`
	ModelVersion   string       `xml:"modelVersion"`
	GroupID        string       `xml:"groupId"`
	ArtifactID     string       `xml:"artifactId"`
	Version        string       `xml:"version"`
	Dependencies   []Dependency `xml:"dependencies>dependency"`
	Description    string       `xml:",innerxml"`
}

type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope,omitempty"`
}

func newPomFile() *PomFile {
	return &PomFile{
		NameSpace:      "http://maven.apache.org/POM/4.0.0",
		SchemeInstance: "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd",
		ModelVersion:   "4.0.0",
		GroupID:        "apache",
		ArtifactID:     "skywalking-eyes",
		Version:        "1.0",
	}
}

func (pom *PomFile) AddDependency(dep Dependency) {
	pom.Dependencies = append(pom.Dependencies, dep)
}

func (pom *PomFile) SetDependency(deps []Dependency) {
	pom.Dependencies = deps
}

func (pom *PomFile) Dump(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := xml.Marshal(pom)
	if err != nil {
		return err
	}

	write := bufio.NewWriter(file)
	write.WriteString(xml.Header)
	_, err = write.Write(data)
	if err != nil {
		return err
	}
	write.Flush()
	return nil
}

func TestResolve(t *testing.T) {
	resolver := new(deps.MavenPomResolver)
	pom := newPomFile()

	tempDir := deps.NewTempDirGenerator()
	path, err := tempDir.Create()
	if err != nil {
		t.Error(err)
		return
	}
	defer tempDir.Destroy()
	pomFile := filepath.Join(path, "pom.xml")

	for _, test := range []struct {
		dep  Dependency
		skip bool
	}{
		{Dependency{"junit", "junit", "4.13.2", ""}, false},
		{Dependency{"junit", "junit", "4.13.2", "test"}, true},
		{Dependency{"org.apache.commons", "commons-math3", "3.6.1", ""}, false},
		{Dependency{"org.apache.commons", "commons-math3", "3.6.1", "test"}, true},
		{Dependency{"commons-logging", "commons-logging", "1.2", "compile"}, false},
		{Dependency{"commons-logging", "commons-logging", "1.2", "test"}, true},
		{Dependency{"org.apache.skywalking", "skywalking-sharing-server-plugin", "8.6.0", "runtime"}, false},
		{Dependency{"org.apache.skywalking", "skywalking-sharing-server-plugin", "8.6.0", "test"}, true},
	} {
		pom.SetDependency([]Dependency{test.dep})
		pom.Dump(pomFile)

		if resolver.CanResolve(pomFile) {
			report := deps.Report{}
			if err := resolver.Resolve(pomFile, &report); err != nil {
				t.Error(err)
			}

			if test.skip && len(report.Resolved) != 0 {
				t.Errorf("these files from dependency %v should be skip but they are not:\n%v", test.dep, report.String())
			}

			if skipped := len(report.Skipped); skipped > 0 {
				pkgs := make([]string, skipped)
				for i, s := range report.Skipped {
					pkgs[i] = s.Dependency
				}
				t.Errorf(
					"failed to identify the licenses of following packages (%d):\n%s",
					len(pkgs), strings.Join(pkgs, "\n"),
				)
			}
		}
	}
}

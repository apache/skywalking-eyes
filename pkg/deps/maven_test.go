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

package deps_test

import (
	"bufio"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apache/skywalking-eyes/pkg/config"
	"github.com/apache/skywalking-eyes/pkg/deps"
)

func TestCanResolvePomFile(t *testing.T) {
	resolver := new(deps.MavenPomResolver)
	for _, test := range []struct {
		fileName string
		exp      bool
	}{
		{"pom.xml", true},
		{"POM.XML", false},
		{"log4j-1.2.12.pom", false},
		{".pom", false},
	} {
		b := resolver.CanResolve(test.fileName)
		if b != test.exp {
			t.Errorf("MavenPomResolver.CanResolve(\"%v\") = %v, want %v", test.fileName, b, test.exp)
		}
	}
}

func writeFile(fileName, content string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	write := bufio.NewWriter(file)
	_, err = write.WriteString(content)
	if err != nil {
		return err
	}

	_ = write.Flush()
	return nil
}

func ensureDir(dirName string) error {
	return os.MkdirAll(dirName, 0777)
}

//go:embed testdata/maven/**/*
var testAssets embed.FS

func copy(assetDir, destination string) error {
	return fs.WalkDir(testAssets, assetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		filename := filepath.Join(destination, strings.Replace(path, assetDir, "", 1))
		if err := ensureDir(filepath.Dir(filename)); err != nil {
			return err
		}

		content, err := testAssets.ReadFile(path)
		if err != nil {
			return err
		}
		writeFile(filename, string(content))

		return nil
	})
}

func TestResolveMaven(t *testing.T) {
	resolver := new(deps.MavenPomResolver)

	for _, test := range []struct {
		workingDir string
		testCase   string
		cnt        int
	}{
		{t.TempDir(), "normal", 110},
		{t.TempDir(), "exclude", 109},
		{t.TempDir(), "exclude-recursive", 7},
	} {
		if err := copy("testdata/maven/base", test.workingDir); err != nil {
			t.Error(err)
		}
		if err := copy(filepath.Join("testdata/maven/cases", test.testCase), test.workingDir); err != nil {
			t.Error(err)
		}

		config, err := config.NewConfigFromFile(filepath.Join(test.workingDir, "licenserc.yaml"))
		if err != nil {
			t.Error(err)
		}

		pomFile := filepath.Join(test.workingDir, "pom.xml")
		if resolver.CanResolve(pomFile) {
			report := deps.Report{}
			if err := resolver.Resolve(pomFile, config.Dependencies(), &report); err != nil {
				t.Error(err)
				return
			}

			if len(report.Resolved)+len(report.Skipped) != test.cnt {
				t.Errorf("the expected number of jar packages is: %d, but actually: %d. result:\n%v", test.cnt, len(report.Resolved)+len(report.Skipped), report.String())
			}
		}
	}
}

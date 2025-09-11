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

	"github.com/apache/skywalking-eyes/pkg/deps"
)

func writeFileRuby(fileName, content string) error {
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

func ensureDirRuby(dirName string) error {
	return os.MkdirAll(dirName, 0777)
}

//go:embed testdata/ruby/**/*
var rubyTestAssets embed.FS

func copyRuby(assetDir, destination string) error {
	return fs.WalkDir(rubyTestAssets, assetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		filename := filepath.Join(destination, strings.Replace(path, assetDir, "", 1))
		if err := ensureDirRuby(filepath.Dir(filename)); err != nil {
			return err
		}
		content, err := rubyTestAssets.ReadFile(path)
		if err != nil {
			return err
		}
		return writeFileRuby(filename, string(content))
	})
}

func TestRubyGemfileLockResolver(t *testing.T) {
	resolver := new(deps.GemfileLockResolver)

	// App case: include all specs (3)
	{
		tmp := t.TempDir()
		if err := copyRuby("testdata/ruby/app", tmp); err != nil {
			t.Fatal(err)
		}
		lock := filepath.Join(tmp, "Gemfile.lock")
		if !resolver.CanResolve(lock) {
			t.Fatalf("GemfileLockResolver cannot resolve %s", lock)
		}
		cfg := &deps.ConfigDeps{Files: []string{lock}, Licenses: []*deps.ConfigDepLicense{
			{Name: "rake", Version: "13.0.6", License: "MIT"},
			{Name: "rspec", Version: "3.10.0", License: "MIT"},
			{Name: "rspec-core", Version: "3.10.1", License: "MIT"},
		}}
		report := deps.Report{}
		if err := resolver.Resolve(lock, cfg, &report); err != nil {
			t.Fatal(err)
		}
		if len(report.Resolved)+len(report.Skipped) != 3 {
			t.Fatalf("expected 3 dependencies, got %d", len(report.Resolved)+len(report.Skipped))
		}
	}

	// Library case: only runtime deps reachable from gemspec (1: rake)
	{
		tmp := t.TempDir()
		if err := copyRuby("testdata/ruby/library", tmp); err != nil {
			t.Fatal(err)
		}
		lock := filepath.Join(tmp, "Gemfile.lock")
		cfg := &deps.ConfigDeps{Files: []string{lock}, Licenses: []*deps.ConfigDepLicense{
			{Name: "rake", Version: "13.0.6", License: "MIT"},
		}}
		report := deps.Report{}
		if err := resolver.Resolve(lock, cfg, &report); err != nil {
			t.Fatal(err)
		}
		if len(report.Resolved)+len(report.Skipped) != 1 {
			t.Fatalf("expected 1 dependency for library, got %d", len(report.Resolved)+len(report.Skipped))
		}
	}
}

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
	"embed"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	resolver := new(GemfileLockResolver)

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
		cfg := &ConfigDeps{Files: []string{lock}, Licenses: []*ConfigDepLicense{
			{Name: "rake", Version: "13.0.6", License: "MIT"},
			{Name: "rspec", Version: "3.10.0", License: "MIT"},
			{Name: "rspec-core", Version: "3.10.1", License: "MIT"},
		}}
		report := Report{}
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
		cfg := &ConfigDeps{Files: []string{lock}, Licenses: []*ConfigDepLicense{
			{Name: "rake", Version: "13.0.6", License: "MIT"},
		}}
		report := Report{}
		if err := resolver.Resolve(lock, cfg, &report); err != nil {
			t.Fatal(err)
		}
		if len(report.Resolved)+len(report.Skipped) != 1 {
			t.Fatalf("expected 1 dependency for library, got %d", len(report.Resolved)+len(report.Skipped))
		}
	}
}

// mock RoundTripper to control HTTP responses
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestRubyMissingSpecIsSkippedGracefully(t *testing.T) {
	// Mock HTTP client to avoid real network: always return 404 Not Found
	saved := httpClientRuby
	httpClientRuby = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Status:     "404 Not Found",
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     make(http.Header),
		}, nil
	})}
	defer func() { httpClientRuby = saved }()

	// Create a Gemfile.lock where a dependency is not present in specs
	content := "" +
		"GEM\n" +
		"  remote: https://rubygems.org/\n" +
		"  specs:\n" +
		"    rake (13.0.6)\n" +
		"\n" +
		"PLATFORMS\n" +
		"  ruby\n" +
		"\n" +
		"DEPENDENCIES\n" +
		"  rake\n" +
		"  missing_gem\n" +
		"\n" +
		"BUNDLED WITH\n" +
		"   2.4.10\n"

	dir := t.TempDir()
	lock := filepath.Join(dir, "Gemfile.lock")
	if err := writeFileRuby(lock, content); err != nil {
		t.Fatal(err)
	}

	resolver := new(GemfileLockResolver)
	cfg := &ConfigDeps{Files: []string{lock}, Licenses: []*ConfigDepLicense{
		{Name: "rake", Version: "13.0.6", License: "MIT"}, // only rake is configured; missing_gem should be skipped
	}}
	report := Report{}
	if err := resolver.Resolve(lock, cfg, &report); err != nil {
		t.Fatal(err)
	}

	if got := len(report.Resolved) + len(report.Skipped); got != 2 {
		t.Fatalf("expected 2 dependencies total, got %d", got)
	}

	// Ensure 'missing_gem' is in skipped with empty version
	found := false
	for _, s := range report.Skipped {
		if s.Dependency == "missing_gem" {
			found = true
			if s.Version != "" {
				t.Fatalf("expected empty version for missing_gem, got %q", s.Version)
			}
		}
	}
	if !found {
		t.Fatalf("expected missing_gem to be marked as skipped")
	}
}

func TestRubyLibraryWithNoRuntimeDependenciesIncludesNone(t *testing.T) {
	resolver := new(GemfileLockResolver)

	// Prepare a library project with a gemspec that has NO runtime dependencies
	dir := t.TempDir()
	lockContent := "" +
		"GEM\n" +
		"  remote: https://rubygems.org/\n" +
		"  specs:\n" +
		"    rake (13.0.6)\n" +
		"    rspec (3.10.0)\n" +
		"      rspec-core (~> 3.10)\n" +
		"    rspec-core (3.10.1)\n" +
		"\n" +
		"PLATFORMS\n" +
		"  ruby\n" +
		"\n" +
		"DEPENDENCIES\n" +
		"  rake\n" +
		"  rspec\n" +
		"\n" +
		"BUNDLED WITH\n" +
		"   2.4.10\n"
	if err := writeFileRuby(filepath.Join(dir, "Gemfile.lock"), lockContent); err != nil {
		t.Fatal(err)
	}

	gemspec := "" +
		"# minimal gemspec without runtime dependencies\n" +
		"Gem::Specification.new do |spec|\n" +
		"  spec.name          = \"sample\"\n" +
		"  spec.version       = \"0.1.0\"\n" +
		"  spec.summary       = \"Sample gem\"\n" +
		"  spec.description   = \"Sample\"\n" +
		"  spec.authors       = [\"Test\"]\n" +
		"  spec.files         = []\n" +
		"  # only development dependency present\n" +
		"  spec.add_development_dependency 'rspec', '~> 3.10'\n" +
		"end\n"
	if err := writeFileRuby(filepath.Join(dir, "sample.gemspec"), gemspec); err != nil {
		t.Fatal(err)
	}

	lock := filepath.Join(dir, "Gemfile.lock")
	cfg := &ConfigDeps{Files: []string{lock}}
	report := Report{}
	if err := resolver.Resolve(lock, cfg, &report); err != nil {
		t.Fatal(err)
	}

	if got := len(report.Resolved) + len(report.Skipped); got != 0 {
		t.Fatalf("expected 0 dependencies for library with no runtime deps, got %d", got)
	}
}

func TestGemspecIgnoresCommentedRuntimeDependencies(t *testing.T) {
	resolver := new(GemfileLockResolver)

	// Prepare a library project whose gemspec contains a commented runtime dependency line
	dir := t.TempDir()
	lockContent := "" +
		"GEM\n" +
		"  remote: https://rubygems.org/\n" +
		"  specs:\n" +
		"    rake (13.0.6)\n" +
		"\n" +
		"PLATFORMS\n" +
		"  ruby\n" +
		"\n" +
		"DEPENDENCIES\n" +
		"  rake\n" +
		"\n" +
		"BUNDLED WITH\n" +
		"   2.4.10\n"
	if err := writeFileRuby(filepath.Join(dir, "Gemfile.lock"), lockContent); err != nil {
		t.Fatal(err)
	}

	gemspec := "" +
		"Gem::Specification.new do |spec|\n" +
		"  spec.name          = \"sample\"\n" +
		"  spec.version       = \"0.1.0\"\n" +
		"  spec.summary       = \"Sample gem\"\n" +
		"  spec.description   = \"Sample\"\n" +
		"  spec.authors       = [\"Test\"]\n" +
		"  spec.files         = []\n" +
		"  # spec.add_dependency(\"git\", \">= 1.19.1\")\n" +
		"end\n"
	if err := writeFileRuby(filepath.Join(dir, "sample.gemspec"), gemspec); err != nil {
		t.Fatal(err)
	}

	lock := filepath.Join(dir, "Gemfile.lock")
	cfg := &ConfigDeps{Files: []string{lock}}
	report := Report{}
	if err := resolver.Resolve(lock, cfg, &report); err != nil {
		t.Fatal(err)
	}

	// There are no runtime dependencies in gemspec; the commented one must be ignored
	if got := len(report.Resolved) + len(report.Skipped); got != 0 {
		t.Fatalf("expected 0 dependencies when runtime deps are only commented, got %d", got)
	}
}

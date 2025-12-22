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
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	tomlRb = "toml-rb"
	citrus = "citrus"
	mit    = "MIT"
)

func TestRubyGemspecResolver(t *testing.T) {
	resolver := new(GemspecResolver)

	t.Run("toml-merge case", func(t *testing.T) {
		tmp := t.TempDir()
		if err := copyRuby("testdata/ruby/toml-merge", tmp); err != nil {
			t.Fatal(err)
		}
		gemspec := filepath.Join(tmp, "toml-merge.gemspec")
		if !resolver.CanResolve(gemspec) {
			t.Fatalf("GemspecResolver cannot resolve %s", gemspec)
		}
		cfg := &ConfigDeps{Files: []string{gemspec}}
		report := Report{}
		if err := resolver.Resolve(gemspec, cfg, &report); err != nil {
			t.Fatal(err)
		}

		// Expect toml-rb dependency.
		found := false
		for _, r := range report.Resolved {
			if r.Dependency == tomlRb {
				found = true
				break
			}
		}
		for _, r := range report.Skipped {
			if r.Dependency == tomlRb {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected toml-rb dependency, got %v", report.Resolved)
		}
	})

	t.Run("citrus case", func(t *testing.T) {
		tmp := t.TempDir()
		gemHome := filepath.Join(tmp, "gemhome")
		specsDir := filepath.Join(gemHome, "specifications")
		if err := os.MkdirAll(specsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("GEM_HOME", gemHome)

		// Create toml-rb gemspec (dependency of toml-merge)
		tomlRbContent := `
Gem::Specification.new do |s|
  s.name = 'toml-rb'
  s.version = '1.0.0'
  s.add_dependency 'citrus', '~> 3.0'
end
`
		if err := writeFileRuby(filepath.Join(specsDir, "toml-rb-1.0.0.gemspec"), tomlRbContent); err != nil {
			t.Fatal(err)
		}

		// Create citrus gemspec (dependency of toml-rb)
		citrusContent := `
Gem::Specification.new do |s|
  s.name = 'citrus'
  s.version = '3.0.2'
  s.licenses = ['MIT']
end
`
		if err := writeFileRuby(filepath.Join(specsDir, "citrus-3.0.2.gemspec"), citrusContent); err != nil {
			t.Fatal(err)
		}

		// Create toml-merge gemspec (the project file)
		tomlMergeContent := `
Gem::Specification.new do |s|
  s.name = 'toml-merge'
  s.version = '0.0.1'
  s.add_dependency 'toml-rb', '~> 1.0'
end
`
		gemspec := filepath.Join(tmp, "toml-merge.gemspec")
		if err := writeFileRuby(gemspec, tomlMergeContent); err != nil {
			t.Fatal(err)
		}

		cfg := &ConfigDeps{Files: []string{gemspec}}
		report := Report{}
		if err := resolver.Resolve(gemspec, cfg, &report); err != nil {
			t.Fatal(err)
		}

		// Check for citrus
		found := false
		var license string
		for _, r := range report.Resolved {
			if r.Dependency == citrus {
				found = true
				license = r.LicenseSpdxID
				break
			}
		}
		if !found {
			// Check skipped
			for _, r := range report.Skipped {
				if r.Dependency == citrus {
					found = true
					license = r.LicenseSpdxID
					break
				}
			}
		}

		if !found {
			t.Error("expected citrus dependency (transitive)")
		} else {
			t.Logf("citrus license: %s", license)
			if license != mit {
				t.Errorf("expected citrus license MIT, got %s", license)
			}
		}
	})

	t.Run("multiple licenses case (non-conflicting)", func(t *testing.T) {
		testMultiLicense(t, resolver, "multi-license", "['MIT', 'Apache-2.0']", "MIT AND Apache-2.0")
	})

	t.Run("multiple licenses case (conflicting/incompatible)", func(t *testing.T) {
		testMultiLicense(t, resolver, "conflicting-license", "['GPL-2.0', 'Apache-2.0']", "GPL-2.0 AND Apache-2.0")
	})
}

func testMultiLicense(t *testing.T, resolver *GemspecResolver, gemName, licensesStr, expectedLicense string) {
	tmp := t.TempDir()
	gemHome := filepath.Join(tmp, "gemhome")
	specsDir := filepath.Join(gemHome, "specifications")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GEM_HOME", gemHome)

	// Create multi-license gem
	gemContent := fmt.Sprintf(`
Gem::Specification.new do |s|
  s.name = '%s'
  s.version = '1.0.0'
  s.licenses = %s
end
`, gemName, licensesStr)
	if err := writeFileRuby(filepath.Join(specsDir, fmt.Sprintf("%s-1.0.0.gemspec", gemName)), gemContent); err != nil {
		t.Fatal(err)
	}

	// Create project gemspec
	projectContent := fmt.Sprintf(`
Gem::Specification.new do |s|
  s.name = 'project'
  s.version = '0.0.1'
  s.add_dependency '%s', '~> 1.0'
end
`, gemName)
	gemspec := filepath.Join(tmp, "project.gemspec")
	if err := writeFileRuby(gemspec, projectContent); err != nil {
		t.Fatal(err)
	}

	cfg := &ConfigDeps{Files: []string{gemspec}}
	report := Report{}
	if err := resolver.Resolve(gemspec, cfg, &report); err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range report.Resolved {
		if r.Dependency == gemName {
			found = true
			if r.LicenseSpdxID != expectedLicense {
				t.Errorf("expected %s license '%s', got '%s'", gemName, expectedLicense, r.LicenseSpdxID)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected %s dependency", gemName)
	}
}

package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRubyGemspecResolver(t *testing.T) {
	resolver := new(GemspecResolver)

	// toml-merge case: parse gemspec, detect license and dependencies
	{
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
			if r.Dependency == "toml-rb" {
				found = true
				break
			}
		}
		for _, r := range report.Skipped {
			if r.Dependency == "toml-rb" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected toml-rb dependency, got %v", report.Resolved)
		}
	}

	// citrus case: transitive dependency resolution via installed gems
	{
		tmp := t.TempDir()
		gemHome := filepath.Join(tmp, "gemhome")
		specsDir := filepath.Join(gemHome, "specifications")
		if err := os.MkdirAll(specsDir, 0755); err != nil {
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
			if r.Dependency == "citrus" {
				found = true
				license = r.LicenseSpdxID
				break
			}
		}
		if !found {
			// Check skipped
			for _, r := range report.Skipped {
				if r.Dependency == "citrus" {
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
			if license != "MIT" {
				t.Errorf("expected citrus license MIT, got %s", license)
			}
		}
	}
}

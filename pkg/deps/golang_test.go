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
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/apache/skywalking-eyes/pkg/deps"
	"github.com/apache/skywalking-eyes/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

const (
	validGoMod = `module example.com/foo

go 1.21
`
	noModuleDirective = `go 1.21
`
	noGoDirective = `module example.com/foo
`
)

func TestCanResolveGoMod(t *testing.T) {
	resolver := new(deps.GoModResolver)
	dir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"go.mod", "go.mod", true},
		{"go.tool.mod", "go.tool.mod", true},
		{"custom.mod", "custom.mod", true},
		{"non-.mod extension", "Cargo.toml", false},
		{"no extension", "go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, dir, tt.filename, validGoMod)
			if got := resolver.CanResolve(path); got != tt.want {
				t.Errorf("CanResolve(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestResolveGoModInvalidFile(t *testing.T) {
	resolver := new(deps.GoModResolver)
	config := &deps.ConfigDeps{Threshold: 75}

	tests := []struct {
		name    string
		content string
	}{
		{"missing module directive", noModuleDirective},
		{"missing go directive", noGoDirective},
		{"non-Go content", "worker_processes auto;\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeTempFile(t, dir, "go.mod", tt.content)
			var report deps.Report
			if err := resolver.Resolve(path, config, &report); err == nil {
				t.Errorf("Resolve should return an error for: %v", tt.name)
			}
		})
	}
}

func TestResolvePackageLicense(t *testing.T) {
	resolver := new(deps.GoModResolver)
	config := &deps.ConfigDeps{Threshold: 75}

	apacheLicense, err := os.ReadFile("../../LICENSE")
	if err != nil {
		t.Fatalf("failed to read LICENSE fixture: %v", err)
	}

	t.Run("license found in module dir", func(t *testing.T) {
		dir := t.TempDir()
		writeTempFile(t, dir, "LICENSE", string(apacheLicense))

		module := &packages.Module{Path: "example.com/foo", Version: "v1.0.0", Dir: dir}
		var report deps.Report
		if err := resolver.ResolvePackageLicense(config, module, &report); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report.Resolved) != 1 {
			t.Fatalf("expected 1 resolved, got %d", len(report.Resolved))
		}
		if report.Resolved[0].LicenseSpdxID != npmLicenseApache20 {
			t.Errorf("expected %v, got %v", npmLicenseApache20, report.Resolved[0].LicenseSpdxID)
		}
	})

	t.Run("no license found", func(t *testing.T) {
		dir := t.TempDir()
		module := &packages.Module{Path: "example.com/foo", Version: "v1.0.0", Dir: dir}
		var report deps.Report
		if err := resolver.ResolvePackageLicense(config, module, &report); err == nil {
			t.Error("expected error when no license file present")
		}
	})
}

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write %v: %v", name, err)
	}
	return path
}

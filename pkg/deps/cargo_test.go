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
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/deps"
)

func TestCanResolveCargo(t *testing.T) {
	resolver := new(deps.CargoTomlResolver)
	if !resolver.CanResolve("Cargo.toml") {
		t.Error("CargoTomlResolver should resolve Cargo.toml")
		return
	}
	if resolver.CanResolve("go.mod") {
		t.Error("CargoTomlResolver shouldn't resolve go.mod")
	}
}

func TestResolveCargos(t *testing.T) {
	if _, err := exec.Command("cargo", "--version").Output(); err != nil {
		logger.Log.Warnf("Failed to find cargo, the test `TestResolveCargo` was skipped")
		return
	}

	{
		cargoToml := `
[package]
name = "foo"
version = "0.0.0"
publish = false
edition = "2021"
license = "Apache-2.0"
`

		config := deps.ConfigDeps{
			Threshold: 0,
			Files:     []string{"Cargo.toml"},
			Licenses:  []*deps.ConfigDepLicense{},
			Excludes:  []deps.Exclude{},
		}

		report := resolveTmpCargo(t, cargoToml, &config)
		if len(report.Resolved) != 1 {
			t.Error("len(report.Resolved) != 1")
		}
		if report.Resolved[0].LicenseSpdxID != "Apache-2.0" {
			t.Error("Package foo license isn't Apache-2.0")
		}
	}

	{
		cargoToml := `
[package]
name = "foo"
version = "0.0.0"
publish = false
edition = "2021"
license = "Apache-2.0"
`

		config := deps.ConfigDeps{
			Threshold: 0,
			Files:     []string{"Cargo.toml"},
			Licenses:  []*deps.ConfigDepLicense{},
			Excludes:  []deps.Exclude{{Name: "foo", Version: "0.0.0"}},
		}

		report := resolveTmpCargo(t, cargoToml, &config)
		if len(report.Resolved) != 0 {
			t.Error("len(report.Resolved) != 0")
		}
	}

	{
		cargoToml := `
[package]
name = "foo"
version = "0.0.0"
publish = false
edition = "2021"
license = "Apache-2.0"
`

		config := deps.ConfigDeps{
			Threshold: 0,
			Files:     []string{},
			Licenses: []*deps.ConfigDepLicense{
				{
					Name:    "foo",
					Version: "0.0.0",
					License: "MIT",
				},
			},
			Excludes: []deps.Exclude{},
		}

		report := resolveTmpCargo(t, cargoToml, &config)
		if len(report.Resolved) != 1 {
			t.Error("len(report.Resolved) != 1")
		}
		if report.Resolved[0].LicenseSpdxID != "MIT" {
			t.Error("Package foo license isn't modified to  MIT")
		}
	}

	{
		cargoToml := `
[package]
name = "foo"
version = "0.0.0"
publish = false
edition = "2021"
license = "Apache-2.0"

[dependencies]
libc = "0.2.126"    # actual license: MIT OR Apache-2.0
bitflags = "1.3.2"  # actual license: MIT/Apache-2.0
`

		config := deps.ConfigDeps{
			Threshold: 0,
			Files:     []string{"Cargo.toml"},
			Licenses:  []*deps.ConfigDepLicense{},
			Excludes:  []deps.Exclude{},
		}

		report := resolveTmpCargo(t, cargoToml, &config)
		if len(report.Resolved) != 3 {
			t.Error("len(report.Resolved) != 3")
		}
		for _, result := range report.Resolved {
			if result.Dependency == "libc" {
				if result.LicenseSpdxID != "Apache-2.0 OR MIT" || result.LicenseContent == "" {
					t.Error("Resolve dependency libc failed")
				}
			}
			if result.Dependency == "bitflags" {
				if result.LicenseSpdxID != "Apache-2.0 OR MIT" || result.LicenseContent == "" {
					t.Error("Resolve dependency libc failed")
				}
			}
		}
	}
}

func resolveTmpCargo(t *testing.T, cargoTomlContent string, config *deps.ConfigDeps) *deps.Report {
	dir, err := os.MkdirTemp("", "skywalking-eyes-test-cargo-")
	if err != nil {
		t.Error("Make temp dir failed", err)
		return nil
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			logger.Log.Warn(err)
		}
	}(dir) // clean up

	if err := os.Chdir(dir); err != nil {
		t.Error("Chdir failed", err)
		return nil
	}

	if _, err := exec.Command("cargo", "init", "--lib").Output(); err != nil {
		t.Error("Cargo init failed", err)
		return nil
	}

	cargoFile := filepath.Join(dir, "Cargo.toml")
	if err := os.WriteFile(cargoFile, []byte(cargoTomlContent), 0644); err != nil {
		t.Error("Write Cargo.toml failed", err)
		return nil
	}

	resolver := new(deps.CargoTomlResolver)

	var report deps.Report
	if err := resolver.Resolve(cargoFile, config, &report); err != nil {
		t.Error("CargoTomlResolver resolve failed", err)
		return nil
	}
	return &report
}

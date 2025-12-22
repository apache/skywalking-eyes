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
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/apache/skywalking-eyes/pkg/deps"
)

const (
	licenseMIT       = "MIT"
	licenseApache20  = "Apache-2.0"
)

// TC-NEW-001
// Regression test: cross-platform npm binary packages must be skipped safely.
func TestResolvePackageLicense_SkipCrossPlatformPackages(t *testing.T) {
	resolver := &deps.NpmResolver{}
	cfg := &deps.ConfigDeps{}

	var crossPlatformPkgs []string
	switch runtime.GOOS {
	case "linux":
		crossPlatformPkgs = []string{
			"@parcel/watcher-darwin-arm64",
			"@parcel/watcher-win32-x64",
		}
	case "darwin":
		crossPlatformPkgs = []string{
			"@parcel/watcher-linux-x64",
			"@parcel/watcher-win32-x64",
		}
	default:
		crossPlatformPkgs = []string{
			"@parcel/watcher-linux-x64",
		}
	}

	for _, pkg := range crossPlatformPkgs {
		pkg := pkg

		t.Run(pkg+"/path-not-exist", func(t *testing.T) {
			// 001-A: cross-platform + path not exist
			result := resolver.ResolvePackageLicense(pkg, "/non/existent/path", cfg)
			if result.LicenseSpdxID != "" {
				t.Fatalf(
					"expected empty license for cross-platform package %q, got %q",
					pkg,
					result.LicenseSpdxID,
				)
			}
		})

		t.Run(pkg+"/package-json-exists", func(t *testing.T) {
			// 001-B: cross-platform + package.json exists
			tmp := t.TempDir()
			err := os.WriteFile(
				filepath.Join(tmp, "package.json"),
				[]byte(`{"n

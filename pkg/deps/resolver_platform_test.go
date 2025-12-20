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
	"testing"

	"github.com/apache/skywalking-eyes/pkg/deps"
)

// -----------------------------
// ResolvePackageLicense
// -----------------------------

func TestResolvePackageLicense_SkipCrossPlatform(t *testing.T) {
	resolver := &deps.NpmResolver{}
	cfg := &deps.ConfigDeps{}

	var pkg string
	switch runtime.GOOS {
	case "linux":
		pkg = "@parcel/watcher-darwin-arm64"
	case "darwin":
		pkg = "@parcel/watcher-linux-x64"
	default:
		pkg = "@parcel/watcher-linux-x64"
	}

	result := resolver.ResolvePackageLicense(
		pkg,
		"/non/existent/path",
		cfg,
	)

	if result.LicenseSpdxID != "" {
		t.Fatalf(
			"expected empty license for cross-platform package %q, got %q",
			pkg, result.LicenseSpdxID,
		)
	}
}

func TestResolvePackageLicense_CurrentPlatform(t *testing.T) {
	resolver := &deps.NpmResolver{}
	cfg := &deps.ConfigDeps{}

	tmp := t.TempDir()
	pkgFile := filepath.Join(tmp, deps.PkgFileName)

	err := os.WriteFile(pkgFile, []byte(`{
		"name": "normal-pkg",
		"license": "Apache-2.0"
	}`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	result := resolver.ResolvePackageLicense(
		"normal-pkg",
		tmp,
		cfg,
	)

	if result.LicenseSpdxID != "Apache-2.0" {
		t.Fatalf(
			"expected license Apache-2.0, got %q",
			result.LicenseSpdxID,
		)
	}
}

func TestResolvePackageLicense_InvalidPath(t *testing.T) {
	resolver := &deps.NpmResolver{}
	cfg := &deps.ConfigDeps{}

	_ = resolver.ResolvePackageLicense(
		"some-random-package",
		"/definitely/not/exist",
		cfg,
	)
}

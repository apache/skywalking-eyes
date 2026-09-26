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

// Internal, because what needs testing here — which projects the resolver
// claims, and how a name is recovered from an install path — is not exported.

package deps

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile creates dir/name containing content, making the directory first.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPnpmResolverCanResolve(t *testing.T) {
	resolver := new(PnpmResolver)

	t.Run("declines a file that is not package.json", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, PnpmLockFileName, "lockfileVersion: '9.0'\n")
		if resolver.CanResolve(filepath.Join(dir, "go.mod")) {
			t.Error("claimed go.mod")
		}
	})

	t.Run("declines an npm project, leaving it to NpmResolver", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, PkgFileName, `{"name":"plain"}`)
		writeFile(t, dir, "package-lock.json", `{}`)
		if resolver.CanResolve(filepath.Join(dir, PkgFileName)) {
			t.Error("claimed an npm project")
		}
	})

	t.Run("claims a project with a pnpm lock file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, PkgFileName, `{"name":"root"}`)
		writeFile(t, dir, PnpmLockFileName, "lockfileVersion: '9.0'\n")
		if !resolver.CanResolve(filepath.Join(dir, PkgFileName)) {
			t.Error("did not claim a pnpm project")
		}
	})

	t.Run("claims a project declaring pnpm without a lock file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, PkgFileName, `{"name":"root","packageManager":"pnpm@9.12.0"}`)
		if !resolver.CanResolve(filepath.Join(dir, PkgFileName)) {
			t.Error("did not honor the packageManager field")
		}
	})

	// A declared manager is decisive in both directions: a yarn project with a
	// leftover pnpm lock file in the tree is not pnpm's to claim.
	t.Run("declines a project declaring another manager", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, PkgFileName, `{"name":"root","packageManager":"yarn@4.1.0"}`)
		writeFile(t, dir, PnpmLockFileName, "lockfileVersion: '9.0'\n")
		if resolver.CanResolve(filepath.Join(dir, PkgFileName)) {
			t.Error("claimed a project that declares yarn")
		}
	})

	// The case that matters in practice: resolution runs from each workspace
	// member in turn, and a member declares neither the field nor the lock file.
	t.Run("claims a workspace member via its root", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, PkgFileName, `{"name":"root"}`)
		writeFile(t, root, PnpmLockFileName, "lockfileVersion: '9.0'\n")
		member := filepath.Join(root, "apps", "bff")
		writeFile(t, member, PkgFileName, `{"name":"@scope/bff"}`)

		if !resolver.CanResolve(filepath.Join(member, PkgFileName)) {
			t.Error("did not claim a workspace member")
		}
	})
}

// The two Node resolvers both answer to package.json and Resolve() takes the
// first match, so this ordering is load-bearing, not cosmetic.
func TestPnpmResolverIsOfferedPackageJSONFirst(t *testing.T) {
	pnpmAt, npmAt := -1, -1
	for i, r := range Resolvers {
		switch r.(type) {
		case *PnpmResolver:
			pnpmAt = i
		case *NpmResolver:
			npmAt = i
		}
	}
	if pnpmAt < 0 || npmAt < 0 {
		t.Fatalf("both Node resolvers must be registered: pnpm=%d npm=%d", pnpmAt, npmAt)
	}
	if pnpmAt > npmAt {
		t.Errorf("PnpmResolver (%d) must precede NpmResolver (%d), or NpmResolver claims every package.json", pnpmAt, npmAt)
	}
}

func TestPnpmWorkspaceRoot(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, PnpmLockFileName, "lockfileVersion: '9.0'\n")
	member := filepath.Join(root, "packages", "api-client")
	if err := os.MkdirAll(member, 0o755); err != nil {
		t.Fatal(err)
	}

	// Every member of one workspace must answer with the same root — that is
	// what stops the same package set being listed once per member.
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{root, member} {
		got, err := filepath.EvalSymlinks(pnpmWorkspaceRoot(dir))
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("from %q: got %q, want %q", dir, got, want)
		}
	}

	if got := pnpmWorkspaceRoot(t.TempDir()); got != "" {
		t.Errorf("outside any workspace: got %q, want empty", got)
	}
}

func TestPnpmPkgName(t *testing.T) {
	for _, c := range []struct {
		path string
		want string
	}{
		// pnpm links out of its store, so the name follows the LAST
		// node_modules segment, not the first.
		{"/w/node_modules/.pnpm/vue@3.5.13/node_modules/vue", "vue"},
		{"/w/node_modules/.pnpm/@vue+shared@3.5.13/node_modules/@vue/shared", "@vue/shared"},
		// A direct link, as a non-hoisted install also produces.
		{"/w/node_modules/vue", "vue"},
		// No node_modules in the path at all: not a dependency, so not listed.
		{"/w/apps/bff", ""},
	} {
		if got := pnpmPkgName(c.path); got != c.want {
			t.Errorf("pnpmPkgName(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

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

package header

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCheckFile(t *testing.T) {
	type args struct {
		name   string
		file   string
		result *Result
	}

	var c struct {
		Header ConfigHeader `yaml:"header"`
	}

	require.NoError(t, os.Chdir("../.."))
	content, err := os.ReadFile("test/testdata/.licenserc_for_test_check.yaml")
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(content, &c))
	require.NoError(t, c.Header.Finalize())

	t.Run("WithLicense", func(t *testing.T) {
		tests := func() []args {
			files, err := filepath.Glob("test/testdata/include_test/with_license/*")
			require.NoError(t, err)
			var cases []args
			for _, file := range files {
				cases = append(cases, args{
					name:   file,
					file:   file,
					result: &Result{},
				})
			}
			return cases
		}()
		require.NotEmpty(t, tests)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.NotEmpty(t, strings.TrimSpace(c.Header.GetLicenseContent()))
				require.NoError(t, CheckFile(tt.file, &c.Header, tt.result))
				require.Len(t, tt.result.Ignored, 0)
				require.False(t, tt.result.HasFailure())
			})
		}
	})

	t.Run("WithoutLicense", func(t *testing.T) {
		tests := func() []args {
			files, err := filepath.Glob("test/testdata/include_test/without_license/*")
			require.NoError(t, err)
			var cases []args
			for _, file := range files {
				cases = append(cases, args{
					name:   file,
					file:   file,
					result: &Result{},
				})
			}
			return cases
		}()
		require.NotEmpty(t, tests)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.NotEmpty(t, strings.TrimSpace(c.Header.GetLicenseContent()))
				require.NoError(t, CheckFile(tt.file, &c.Header, tt.result))
				require.Len(t, tt.result.Ignored, 0)
				require.True(t, tt.result.HasFailure())
			})
		}
	})
}

func TestListFilesWithEmptyRepo(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "skywalking-eyes-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if chErr := os.Chdir(tempDir); chErr != nil {
		t.Fatal(chErr)
	}

	// Initialize an empty git repository
	_, err = git.PlainInit(".", false)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	// Create a basic config
	config := &ConfigHeader{
		Paths: []string{"**/*.go"},
	}

	// This should not panic even with empty repository
	fileList, err := listFiles(config.Paths)
	if err != nil {
		t.Fatal(err)
	}

	// Should still find files using glob fallback
	if len(fileList) == 0 {
		t.Error("Expected to find at least one file")
	}
}

func TestListFilesWithWorktreeDetachedHEAD(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "skywalking-eyes-worktree-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if chErr := os.Chdir(tempDir); chErr != nil {
		t.Fatal(chErr)
	}

	// Initialize a git repository with a commit
	repo, err := git.PlainInit(".", false)
	if err != nil {
		t.Fatal(err)
	}

	// Create and commit a file
	testFile := "test.go"
	err = os.WriteFile(testFile, []byte("package main"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	_, err = worktree.Add(testFile)
	if err != nil {
		t.Fatal(err)
	}

	commit, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// First, verify normal case works
	config := &ConfigHeader{
		Paths: []string{"**/*.go"},
	}

	fileList, err := listFiles(config.Paths)
	if err != nil {
		t.Fatal(err)
	}

	if len(fileList) == 0 {
		t.Error("Expected to find files with valid commit")
	}

	// Now simulate detached HEAD by checking out to a non-existent commit hash
	// This will create an invalid HEAD state that our fix should handle
	invalidHash := plumbing.NewHash("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: invalidHash,
	})
	// We expect this to fail, creating an invalid state
	if err == nil {
		t.Fatal("Expected checkout to invalid hash to fail")
	}

	// This should not panic even with problematic git state
	fileList2, err := listFiles(config.Paths)
	if err != nil {
		// It's okay if there's an error, we just don't want a panic
		t.Logf("Got expected error: %v", err)
	}

	// Should still find files using glob fallback
	if len(fileList2) == 0 {
		t.Error("Expected to find at least one file via fallback")
	}

	t.Logf("Found %d files: %v", len(fileList2), fileList2)

	// Verify we can find our test file
	found := false
	for _, file := range fileList2 {
		if filepath.Base(file) == testFile {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find test.go in file list")
	}

	// Test with valid commit to ensure normal case still works
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: commit,
	})
	if err != nil {
		t.Fatal(err)
	}

	fileList3, err := listFiles(config.Paths)
	if err != nil {
		t.Fatal(err)
	}

	if len(fileList3) == 0 {
		t.Error("Expected to find files with valid commit")
	}
}

// initRepoWithFiles creates a Git repository in the current directory and
// commits the given files to it.
func initRepoWithFiles(t *testing.T, files map[string]string) {
	t.Helper()

	repo, err := git.PlainInit(".", false)
	require.NoError(t, err)
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	for name, content := range files {
		require.NoError(t, os.MkdirAll(filepath.Dir(name), 0o755))
		require.NoError(t, os.WriteFile(name, []byte(content), 0o600))
		_, err = worktree.Add(name)
		require.NoError(t, err)
	}

	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test User", Email: "test@example.com"},
	})
	require.NoError(t, err)
}

func TestListFilesWithExplicitFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	initRepoWithFiles(t, map[string]string{
		"tracked.go":   "package main",
		"sub/inner.go": "package sub",
		".gitignore":   "ignored.go\n",
	})
	require.NoError(t, os.WriteFile("untracked.go", []byte("package main"), 0o600))
	require.NoError(t, os.WriteFile("ignored.go", []byte("package main"), 0o600))

	tests := []struct {
		name  string
		paths []string
		want  []string
	}{
		{
			name:  "existing files are returned as given, deduplicated",
			paths: []string{"tracked.go", "sub/inner.go", "untracked.go", "tracked.go"},
			want:  []string{"tracked.go", "sub/inner.go", "untracked.go"},
		},
		{
			name:  "explicitly named gitignored file is returned",
			paths: []string{"ignored.go"},
			want:  []string{"ignored.go"},
		},
		{
			name:  "explicitly named file is kept when another path needs discovery",
			paths: []string{"ignored.go", "missing.go"},
			want:  []string{"ignored.go"},
		},
		{
			name:  "explicit files and directories are combined without duplicates",
			paths: []string{"ignored.go", "sub/inner.go", "sub/"},
			want:  []string{"ignored.go", "sub/inner.go"},
		},
		{
			name:  "directory goes through git discovery",
			paths: []string{"sub/"},
			want:  []string{"sub/inner.go"},
		},
		{
			name:  "glob goes through git discovery and respects gitignore",
			paths: []string{"*.go"},
			want:  []string{"tracked.go", "untracked.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := listFiles(tt.paths)
			require.NoError(t, err)
			require.ElementsMatch(t, tt.want, files)
		})
	}
}

func TestListFilesWithExplicitFilesSkipsGitDiscovery(t *testing.T) {
	t.Chdir(t.TempDir())

	initRepoWithFiles(t, map[string]string{
		"a.go": "package a",
		"b.go": "package b",
	})

	// Delete b.go's blob so that walking the HEAD tree fails.
	blob := plumbing.ComputeHash(plumbing.BlobObject, []byte("package b")).String()
	require.NoError(t, os.Remove(filepath.Join(".git", "objects", blob[:2], blob[2:])))

	_, err := listFiles([]string{"**/*.go"})
	require.Error(t, err, "git discovery should fail on the broken repository")

	files, err := listFiles([]string{"a.go"})
	require.NoError(t, err)
	require.Equal(t, []string{"a.go"}, files)
}

func TestSplitExplicitFilesRejectsPathsGitWouldNotList(t *testing.T) {
	t.Chdir(t.TempDir())

	initRepoWithFiles(t, map[string]string{
		"tracked.go":   "package main",
		"sub/inner.go": "package sub",
	})

	files, rest := splitExplicitFiles([]string{"tracked.go", "sub/inner.go", "tracked.go"})
	require.Equal(t, []string{"tracked.go", "sub/inner.go"}, files)
	require.Empty(t, rest)

	// A nested repository and an uninitialized submodule, whose files Git-based
	// discovery does not list.
	require.NoError(t, os.MkdirAll(filepath.Join("nested", ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join("nested", "n.go"), []byte("package n"), 0o600))
	require.NoError(t, os.WriteFile(".gitmodules",
		[]byte("[submodule \"lib\"]\n\tpath = third_party/lib\n\turl = https://example.com/lib.git\n"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join("third_party", "lib"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join("third_party", "lib", "lib.go"), []byte("package lib"), 0o600))

	tests := []struct {
		name  string
		paths []string
	}{
		{name: "git metadata", paths: []string{".git/HEAD"}},
		{name: "git metadata spelled in another case", paths: []string{".GIT/HEAD"}},
		{name: "other spelling of an existing file", paths: []string{"TRACKED.go"}},
		{name: "other spelling of a parent directory", paths: []string{"SUB/inner.go"}},
		{name: "file inside a nested repository", paths: []string{"nested/n.go"}},
		{name: "file inside an uninitialized submodule", paths: []string{"third_party/lib/lib.go"}},
		{name: "missing file", paths: []string{"missing.go"}},
		{name: "directory", paths: []string{"sub"}},
		{name: "glob", paths: []string{"*.go"}},
	}

	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "x.go"), []byte("package x"), 0o600))
	if err := os.Symlink(outside, "link"); err == nil {
		require.NoError(t, os.Symlink(filepath.Join(outside, "x.go"), "leaf.go"))
		tests = append(tests, []struct {
			name  string
			paths []string
		}{
			{name: "file below a symlinked directory", paths: []string{"link/x.go"}},
			{name: "symlink to a file", paths: []string{"leaf.go"}},
		}...)
	} else {
		t.Logf("skipping symlinked directory case: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, rest := splitExplicitFiles(append([]string{"tracked.go"}, tt.paths...))
			require.Equal(t, []string{"tracked.go"}, files)
			require.Equal(t, tt.paths, rest)
		})
	}
}

func TestListFilesInBareRepository(t *testing.T) {
	t.Chdir(t.TempDir())

	_, err := git.PlainInit(".", true)
	require.NoError(t, err)

	_, err = listFiles([]string{"HEAD"})
	require.ErrorIs(t, err, git.ErrIsBareRepository)
}

func TestIsLiteralLocalPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"a.go", true},
		{"sub/inner.go", true},
		{".github/workflows/ci.yaml", true},
		{"", false},
		{".", true}, // literal, but listFiles still falls back because it is a directory
		{"./a.go", false},
		{"sub//inner.go", false},
		{"sub/./inner.go", false},
		{"sub/../a.go", false},
		{"../a.go", false},
		{"/abs/a.go", false},
		{"sub/", false},
		{"*.go", false},
		{"a?.go", false},
		{"[ab].go", false},
		{"{a,b}.go", false},
		{`sub\inner.go`, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			require.Equal(t, tt.want, isLiteralLocalPath(tt.path))
		})
	}
}

func TestMatchPaths(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		patterns []string
		expected bool
	}{
		{
			name:     "Exact file match",
			file:     "test.go",
			patterns: []string{"test.go"},
			expected: true,
		},
		{
			name:     "Glob pattern match",
			file:     "test.go",
			patterns: []string{"*.go"},
			expected: true,
		},
		{
			name:     "Double-star glob pattern match",
			file:     "pkg/header/check.go",
			patterns: []string{"**/*.go"},
			expected: true,
		},
		{
			name:     "Multiple patterns with match",
			file:     "test.go",
			patterns: []string{"*.java", "*.go", "*.py"},
			expected: true,
		},
		{
			name:     "Multiple patterns without match",
			file:     "test.go",
			patterns: []string{"*.java", "*.py"},
			expected: false,
		},
		{
			name:     "Directory pattern with trailing slash",
			file:     "pkg/header/check.go",
			patterns: []string{"pkg/header/"},
			expected: true,
		},
		{
			name:     "Directory pattern without trailing slash",
			file:     "pkg/header/check.go",
			patterns: []string{"pkg/header"},
			expected: false,
		},
		{
			name:     "Dot pattern",
			file:     "test.go",
			patterns: []string{"."},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchPaths(tt.file, tt.patterns)
			if result != tt.expected {
				t.Errorf(
					"MatchPaths(%q, %v) = %v, want %v",
					tt.file,
					tt.patterns,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestCheckPathsWithMultipleHeaders(t *testing.T) {
	t.Chdir(t.TempDir())

	initRepoWithFiles(t, map[string]string{
		"a/x.go": "package a\n",
		"b/y.go": "package b\n",
	})

	headerA := &ConfigHeader{License: LicenseConfig{Content: "Licensed under license A."}, Paths: []string{"a/**"}}
	headerB := &ConfigHeader{License: LicenseConfig{Content: "Licensed under license B."}, Paths: []string{"b/**"}}
	for _, h := range []*ConfigHeader{headerA, headerB} {
		require.NoError(t, h.Finalize())
	}

	for _, paths := range [][]string{
		{"a/x.go", "b/y.go"}, // explicit files, as a pre-commit hook passes them
		{"**/*.go"},          // a glob that goes through Git-based discovery
	} {
		t.Run(strings.Join(paths, ","), func(t *testing.T) {
			var resultA, resultB Result
			require.NoError(t, CheckPaths(headerA, paths, &resultA))
			require.NoError(t, CheckPaths(headerB, paths, &resultB))

			require.Equal(t, []string{"a/x.go"}, resultA.Failure)
			require.Equal(t, []string{"b/y.go"}, resultA.Ignored)
			require.Equal(t, []string{"b/y.go"}, resultB.Failure)
			require.Equal(t, []string{"a/x.go"}, resultB.Ignored)
		})
	}

	// Fixing the files given as paths inserts only the header that covers each file.
	paths := []string{"a/x.go", "b/y.go"}
	for _, h := range []*ConfigHeader{headerA, headerB} {
		var result Result
		require.NoError(t, CheckPaths(h, paths, &result))
		for _, file := range result.Failure {
			require.NoError(t, Fix(file, h, &result))
		}
	}

	for file, want := range map[string]struct{ has, hasNot string }{
		"a/x.go": {has: "license A", hasNot: "license B"},
		"b/y.go": {has: "license B", hasNot: "license A"},
	} {
		content, err := os.ReadFile(file)
		require.NoError(t, err)
		require.Contains(t, string(content), want.has, file)
		require.NotContains(t, string(content), want.hasNot, file)
	}

	for _, h := range []*ConfigHeader{headerA, headerB} {
		var result Result
		require.NoError(t, CheckPaths(h, paths, &result))
		require.False(t, result.HasFailure())
		require.Len(t, result.Success, 1)
	}
}

func TestCheckPathsSelection(t *testing.T) {
	license := LicenseConfig{Content: "Licensed under license A."}
	newHeader := func(paths, pathsIgnore []string) *ConfigHeader {
		h := &ConfigHeader{License: license, Paths: paths, PathsIgnore: pathsIgnore}
		require.NoError(t, h.Finalize())
		return h
	}
	writeFiles := func(files ...string) {
		for _, file := range files {
			require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
			require.NoError(t, os.WriteFile(file, []byte("package p\n"), 0o600))
		}
	}

	t.Run("a path does not cover files elsewhere with the same base name", func(t *testing.T) {
		t.Chdir(t.TempDir())
		initRepoWithFiles(t, map[string]string{"main.go": "package main\n", "sub/main.go": "package sub\n"})

		var result Result
		require.NoError(t, CheckPaths(newHeader([]string{"main.go"}, nil), []string{"sub/main.go"}, &result))
		require.Equal(t, []string{"sub/main.go"}, result.Ignored)
		require.Empty(t, result.Failure)
	})

	t.Run("a glob does not select the files inside the directories it matches", func(t *testing.T) {
		t.Chdir(t.TempDir()) // not a Git repository
		writeFiles("src/a.go", "src/nested/b.go")

		var result Result
		require.NoError(t, CheckPaths(newHeader(nil, nil), []string{"src/*"}, &result))
		require.Equal(t, []string{"src/a.go"}, result.Failure)
		require.Equal(t, []string{"src/nested/b.go"}, result.Ignored)
	})

	t.Run("an absolute path is matched relative to the current directory", func(t *testing.T) {
		t.Chdir(t.TempDir()) // not a Git repository
		writeFiles("src/a.go", "src/gen/b.go", "other/c.go")
		wd, err := os.Getwd()
		require.NoError(t, err)
		a, b, c := filepath.Join(wd, "src", "a.go"), filepath.Join(wd, "src", "gen", "b.go"), filepath.Join(wd, "other", "c.go")

		var result Result
		require.NoError(t, CheckPaths(newHeader([]string{"src/**"}, []string{"src/gen/**"}), []string{a, b, c}, &result))
		require.Equal(t, []string{a}, result.Failure)
		require.ElementsMatch(t, []string{b, c}, result.Ignored)
	})

	t.Run("a malformed glob is reported", func(t *testing.T) {
		t.Chdir(t.TempDir())
		initRepoWithFiles(t, map[string]string{"main.go": "package main\n"})

		var result Result
		require.ErrorIs(t, CheckPaths(newHeader(nil, nil), []string{"[ab.go", "main.go"}, &result), doublestar.ErrBadPattern)
	})
}

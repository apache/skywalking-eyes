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
	fileList, err := listFiles(config)
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

	fileList, err := listFiles(config)
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
	fileList2, err := listFiles(config)
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

	fileList3, err := listFiles(config)
	if err != nil {
		t.Fatal(err)
	}

	if len(fileList3) == 0 {
		t.Error("Expected to find files with valid commit")
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
			patterns: []string{"pkg/header}"},
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

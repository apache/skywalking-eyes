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
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	eyeignore "github.com/apache/skywalking-eyes/pkg/gitignore"
	lcs "github.com/apache/skywalking-eyes/pkg/license"
	"github.com/apache/skywalking-eyes/pkg/logger"

	"github.com/bmatcuk/doublestar/v2"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/sync/errgroup"
)

const currentDir = "./"

// Check checks the license headers of the specified paths/globs.
func Check(config *ConfigHeader, result *Result) error {
	fileList, err := listFiles(config)
	if err != nil {
		return err
	}

	g := new(errgroup.Group)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for _, file := range fileList {
		f := file
		g.Go(func() error {
			return CheckFile(f, config, result)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func listFiles(config *ConfigHeader) ([]string, error) {
	var fileList []string

	repo, err := git.PlainOpen(currentDir)

	if err != nil { // we're not in a Git workspace, fallback to glob paths
		var localFileList []string
		for _, pattern := range config.Paths {
			if pattern == "." {
				pattern = currentDir
			}
			files, err := doublestar.Glob(pattern)
			if err != nil {
				return nil, err
			}
			localFileList = append(localFileList, files...)
		}

		seen := make(map[string]bool)
		for _, file := range localFileList {
			files, err := walkFile(file, seen)
			if err != nil {
				return nil, err
			}
			fileList = append(fileList, files...)
		}
	} else {
		t, err := repo.Worktree()
		if err != nil {
			return nil, err
		}

		// Paths that name existing files directly, e.g. the changed files a
		// pre-commit hook passes on the command line, need no discovery. Git-based
		// discovery scans the whole worktree and HEAD tree, which is slow in large
		// repositories, so it only runs for the remaining paths.
		explicit, patterns := splitExplicitFiles(config.Paths)
		fileList = append(fileList, explicit...)
		if len(patterns) == 0 {
			logger.Log.Debugln("All paths name existing files, skipping Git-based file discovery.")
			return fileList, nil
		}

		var candidates []string

		if t.Excludes == nil {
			t.Excludes = make([]gitignore.Pattern, 0)
		}
		addIgnorePatterns(t)
		s, _ := t.Status()
		for file := range s {
			candidates = append(candidates, file)
		}

		head, err := repo.Head()
		if err != nil || head == nil {
			// Repository has no commits or invalid HEAD, skip git-based file discovery
			logger.Log.Debugf("Repository has no commits or invalid HEAD (head: %v), skipping git-based file discovery. Error: %v", head, err)
		} else {
			commit, err := repo.CommitObject(head.Hash())
			if err != nil {
				logger.Log.Debugln("Failed to get commit object:", err)
			} else {
				tree, err := commit.Tree()
				if err != nil {
					return nil, err
				}
				if err := tree.Files().ForEach(func(file *object.File) error {
					if file == nil {
						return errors.New("file pointer is nil")
					}
					candidates = append(candidates, file.Name)
					return nil
				}); err != nil {
					if errors.Is(err, plumbing.ErrObjectNotFound) {
						return nil, errors.New(
							"failed to read git repository. Run 'git fsck' to diagnose. If dangling objects are found, run: git prune && git gc --prune=now --aggressive",
						)
					}
					return nil, err
				}
			}
		}

		seen := make(map[string]bool)
		for _, file := range explicit {
			seen[file] = true
		}
		for _, candidate := range candidates {
			if !seen[candidate] {
				seen[candidate] = true
				_, err := os.Stat(candidate)
				if err == nil {
					// Filter candidates by the paths/patterns specified in config
					if MatchPaths(candidate, patterns) {
						fileList = append(fileList, candidate)
					}
				} else if !os.IsNotExist(err) {
					return nil, err
				}
			}
		}
	}

	return fileList, nil
}

// splitExplicitFiles separates the paths that name existing files directly from
// the rest, which need Git-based discovery. A path names a file directly when it
// is a literal, clean path inside the current directory (see isLiteralLocalPath)
// that Git-based discovery could return under the same spelling (see
// isDiscoverableFile). Duplicate files are dropped.
func splitExplicitFiles(paths []string) (files, rest []string) {
	submodules, err := submodulePaths()
	if err != nil {
		logger.Log.Debugln("Failed to read .gitmodules, using Git-based discovery for all paths:", err)
		return nil, paths
	}

	seen := make(map[string]bool, len(paths))
	dirs := make(map[string]*dirListing)
	for _, p := range paths {
		if !isLiteralLocalPath(p) || inSubmodule(p, submodules) || !isDiscoverableFile(p, dirs) {
			rest = append(rest, p)
			continue
		}
		if !seen[p] {
			seen[p] = true
			files = append(files, p)
		}
	}
	return files, rest
}

// submodulePaths returns the submodule paths declared in .gitmodules. Git-based
// discovery skips files in submodules, including uninitialized ones, which have
// no .git entry of their own.
func submodulePaths() ([]string, error) {
	data, err := os.ReadFile(".gitmodules")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	modules := gitconfig.NewModules()
	if err := modules.Unmarshal(data); err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(modules.Submodules))
	for _, m := range modules.Submodules {
		if p := strings.TrimSuffix(m.Path, "/"); p != "" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

func inSubmodule(p string, submodules []string) bool {
	for _, s := range submodules {
		if p == s || strings.HasPrefix(p, s+"/") {
			return true
		}
	}
	return false
}

// dirListing is the cached content of one directory.
type dirListing struct {
	entries map[string]fs.FileMode // entry name to type bits
	hasGit  bool                   // has an entry named .git, ignoring case
}

func listDir(dir string, dirs map[string]*dirListing) (*dirListing, error) {
	if l, ok := dirs[dir]; ok {
		return l, nil
	}
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	l := &dirListing{entries: make(map[string]fs.FileMode, len(des))}
	for _, de := range des {
		l.entries[de.Name()] = de.Type()
		l.hasGit = l.hasGit || strings.EqualFold(de.Name(), ".git")
	}
	dirs[dir] = l
	return l, nil
}

// isDiscoverableFile reports whether the literal local path p names a regular
// file that Git-based discovery could return under the same spelling. Each
// component must match a directory entry exactly, because case-insensitive
// filesystems also accept other spellings, which paths-ignore would not match.
// Every parent must be a real directory: not a symlink, which could lead outside
// the repository, and not .git or a directory holding its own .git (a submodule
// or nested repository). The file itself must be a regular file, not a symlink,
// whose target could also be outside the repository; Git-based discovery decides
// about symlinks as before. dirs caches directory listings across calls.
func isDiscoverableFile(p string, dirs map[string]*dirListing) bool {
	parts := strings.Split(p, "/")
	dir := "."
	for i, name := range parts {
		if strings.EqualFold(name, ".git") {
			return false
		}
		l, err := listDir(dir, dirs)
		if err != nil || (i > 0 && l.hasGit) {
			return false
		}
		mode, ok := l.entries[name]
		if !ok {
			return false
		}
		if i == len(parts)-1 {
			return mode.IsRegular()
		}
		if !mode.IsDir() {
			return false
		}
		dir = filepath.Join(dir, name)
	}

	return false
}

// isLiteralLocalPath reports whether p is a relative path inside the current
// directory, in clean slash-separated form (no "./", "..", or "//"), and free of
// glob syntax. Backslash is treated as glob syntax because doublestar uses it as
// the escape character.
func isLiteralLocalPath(p string) bool {
	return filepath.IsLocal(p) &&
		filepath.ToSlash(filepath.Clean(p)) == p &&
		!strings.ContainsAny(p, `*?[]{}\`)
}

func MatchPaths(file string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == "." {
			pattern = currentDir
		}
		matched, err := doublestar.Match(pattern, file)
		if err == nil && matched {
			return true
		}
		// Fallback for directory patterns (e.g., "pkg/header/" or "pkg/header")
		// to behave consistently with tryMatchPatten in config.go.
		// Ensure that pattern ends with a path separator so it clearly denotes a directory.
		if strings.HasSuffix(pattern, "/") ||
			strings.HasSuffix(pattern, string(os.PathSeparator)) {
			dirPattern := strings.TrimRight(pattern, "/")
			// Match if file's directory name equals pattern (exact directory match)
			if stat, err := os.Stat(file); err == nil {
				if stat.Name() == dirPattern {
					return true
				}
			}
			// Match if file is under the directory (prefix match)
			// Normalize both paths to use '/' so prefix matching is consistent across platforms.
			normalizedDir := strings.ReplaceAll(pattern, "\\", "/")
			normalizedFile := strings.ReplaceAll(file, "\\", "/")
			if strings.HasPrefix(normalizedFile, normalizedDir) {
				return true
			}
		}
	}
	return false
}

func addIgnorePatterns(t *git.Worktree) {
	if ignorePattens, err := gitignore.LoadGlobalPatterns(osfs.New("")); err == nil {
		t.Excludes = append(t.Excludes, ignorePattens...)
	}
	if ignorePattens, err := gitignore.LoadSystemPatterns(osfs.New("")); err == nil {
		t.Excludes = append(t.Excludes, ignorePattens...)
	}
	if ignorePatterns, err := eyeignore.LoadGlobalPatterns(); err == nil {
		t.Excludes = append(t.Excludes, ignorePatterns...)
	}
	if ignorePatterns, err := eyeignore.LoadSystemPatterns(); err == nil {
		t.Excludes = append(t.Excludes, ignorePatterns...)
	}
	if ignorePatterns, err := eyeignore.LoadGlobalIgnoreFile(); err == nil {
		t.Excludes = append(t.Excludes, ignorePatterns...)
	}
}

func walkFile(file string, seen map[string]bool) ([]string, error) {
	var files []string

	if seen[file] {
		return files, nil
	}
	seen[file] = true

	if stat, err := os.Stat(file); err == nil {
		switch mode := stat.Mode(); {
		case mode.IsRegular():
			files = append(files, file)
		case mode.IsDir():
			err := filepath.Walk(file, func(path string, info fs.FileInfo, _ error) error {
				if path == file {
					// when path is symbolic link file, it causes infinite recursive calls
					return nil
				}
				if seen[path] {
					return nil
				}
				seen[path] = true
				if info.Mode().IsRegular() {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return files, err
			}
		}
	}

	return files, nil
}

// CheckFile checks whether the file contains the configured license header.
func CheckFile(file string, config *ConfigHeader, result *Result) error {
	if yes, err := config.ShouldIgnore(file); yes || err != nil {
		result.Ignore(file)
		return err
	}

	logger.Log.Debugln("Checking file:", file)

	bs, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	content, encoding, err := decodeContent(bs)
	if err != nil {
		// Report the file instead of aborting the whole check. Fixing it is
		// rejected later by InsertComment, which leaves the file untouched.
		logger.Log.Warnln("Failed to decode file:", file, "; error:", err)
		result.Fail(file)
		return nil
	}
	// A BOM already implies textual content, so only files without one need
	// the MIME sniffing.
	if encoding == encodingRaw {
		if t := http.DetectContentType(bs); !strings.HasPrefix(t, "text/") {
			logger.Log.Debugln("Ignoring file:", file, "; type:", t)
			return nil
		}
	}

	normalized := lcs.NormalizeHeader(string(content))
	expected, pattern := config.NormalizedLicense(), config.NormalizedPattern()

	if satisfy(normalized, config, expected, pattern) {
		result.Succeed(file)
	} else {
		logger.Log.Debugln("Content is:", normalized)

		result.Fail(file)
	}

	return nil
}

func satisfy(content string, config *ConfigHeader, license string, pattern *regexp.Regexp) bool {
	if index := strings.Index(content, license); strings.TrimSpace(license) != "" && index >= 0 {
		return index < config.LicenseLocationThreshold
	}

	if pattern == nil {
		return false
	}
	index := pattern.FindStringIndex(content)

	return len(index) == 2 && index[0] < config.LicenseLocationThreshold
}

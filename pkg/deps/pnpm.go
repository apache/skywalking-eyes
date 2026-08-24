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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apache/skywalking-eyes/pkg/logger"
)

// PnpmLockFileName identifies a pnpm-managed project.
const PnpmLockFileName = "pnpm-lock.yaml"

// pnpmCleanInstallMajor is the first pnpm major with a `pnpm ci` command
// (aliased clean-install / install-clean). Below it, the lock-pinned install
// has to be spelled out as flags.
const pnpmCleanInstallMajor = 8

// PnpmResolver resolves the licenses of a pnpm-managed project's dependencies.
//
// It handles the same `package.json` as NpmResolver and reads licenses exactly
// the same way — a package resolves identically whichever manager installed it.
// What differs is finding the packages: pnpm links dependencies out of a
// content-addressed store rather than laying them out under
// `node_modules/<name>`, so neither `npm ls --parseable` nor the name npm
// infers from an install path applies. NpmResolver is embedded for the shared
// half, and only the listing is replaced.
type PnpmResolver struct {
	NpmResolver

	// Workspaces already listed. `pnpm licenses list` reports the whole
	// workspace whatever member it runs from, so a config naming several
	// package.json files would otherwise report every package once per member.
	listedWorkspaces map[string]bool
}

// CanResolve is deliberately narrower than NpmResolver's: it claims a
// package.json only when the project is pnpm-managed. Registration order in
// Resolvers therefore matters — this resolver must be offered the file first,
// or NpmResolver will claim every package.json and pnpm projects resolve as
// nothing.
func (resolver *PnpmResolver) CanResolve(file string) bool {
	if filepath.Base(file) != PkgFileName {
		return false
	}
	return isPnpmProject(filepath.Dir(file))
}

// isPnpmProject reports whether the project rooted at or above dir is managed
// by pnpm.
//
// The corepack `packageManager` field is the project's own statement of what it
// is built with, so it is believed first; a project that predates corepack, or
// simply omits the field, is identified by its lock file. Both live at the
// workspace ROOT while resolution runs from each MEMBER, so this climbs.
func isPnpmProject(dir string) bool {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	for {
		if content, err := os.ReadFile(filepath.Join(abs, PkgFileName)); err == nil {
			var manifest struct {
				PackageManager string `json:"packageManager"`
			}
			if err := json.Unmarshal(content, &manifest); err == nil && manifest.PackageManager != "" {
				name, _, _ := strings.Cut(manifest.PackageManager, "@")
				// A declared manager is decisive in BOTH directions: a project
				// naming yarn is not pnpm's to claim, whatever lock files a
				// stale working tree happens to contain.
				return name == "pnpm"
			}
		}
		if _, err := os.Stat(filepath.Join(abs, PnpmLockFileName)); err == nil {
			return true
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return false
		}
		abs = parent
	}
}

// pnpmWorkspaceRoot walks up from dir to the directory holding the pnpm lock
// file, which is the workspace a member belongs to. Empty when there is none.
func pnpmWorkspaceRoot(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, PnpmLockFileName)); err == nil {
			return abs
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return ""
		}
		abs = parent
	}
}

// Resolve resolves licenses of all dependencies declared in the package.json file.
func (resolver *PnpmResolver) Resolve(pkgFile string, config *ConfigDeps, report *Report) error {
	workDir := filepath.Dir(pkgFile)
	if err := os.Chdir(workDir); err != nil {
		return err
	}

	if needSkip := resolver.NeedSkipInstallPkgs(); !needSkip {
		resolver.InstallPkgs()
	}

	pkgs := resolver.GetInstalledPkgs(filepath.Join(workDir, "node_modules"))
	resolver.ResolvePackages(pkgs, config, report)
	return nil
}

// InstallPkgs runs `pnpm ci` to install node packages, the counterpart of
// `npm ci`: a clean, lock-pinned install, for reproducible builds.
func (resolver *PnpmResolver) InstallPkgs() {
	cmd := exec.Command("pnpm", "ci")
	if !pnpmSupportsCleanInstall() {
		cmd = exec.Command("pnpm", "install", "--frozen-lockfile")
	}
	logger.Log.Println(fmt.Sprintf("Run command: %v, please wait", cmd.String()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Log.Errorln(err)
	}
}

// pnpmSupportsCleanInstall reports whether the pnpm on PATH has `pnpm ci`.
// An unreadable or unparseable version is taken as "no", so the explicit flags
// are used rather than a command that may not exist.
func pnpmSupportsCleanInstall() bool {
	out, err := exec.Command("pnpm", "--version").Output()
	if err != nil {
		return false
	}
	major, _, _ := strings.Cut(strings.TrimSpace(string(out)), ".")
	n, err := strconv.Atoi(major)
	if err != nil {
		return false
	}
	return n >= pnpmCleanInstallMajor
}

// ListPkgPaths lists every production package pnpm has installed, one absolute
// path per line, matching what `npm ls --parseable` produces for npm.
//
// pnpm has no `--parseable`, and `pnpm ls` reports the dependency GRAPH rather
// than install locations. `pnpm licenses list` reports the locations directly,
// which is what this needs — the license it also reports is left to the normal
// resolution path, so a package is read the same way whichever manager
// installed it.
func (resolver *PnpmResolver) ListPkgPaths() (io.Reader, error) {
	// Keyed on the workspace ROOT, not the working directory: resolution runs
	// from each member in turn, so the cwd differs every time while the
	// workspace — and the package set — is the same one.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	root := pnpmWorkspaceRoot(cwd)
	if root == "" {
		root = cwd
	}
	if resolver.listedWorkspaces[root] {
		return &bytes.Buffer{}, nil
	}

	buffer := &bytes.Buffer{}
	cmd := exec.Command("pnpm", "licenses", "list", "--prod", "--json")
	cmd.Stdout = buffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list pnpm packages: %w", err)
	}

	var grouped map[string][]struct {
		Paths []string `json:"paths"`
	}
	if err := json.Unmarshal(buffer.Bytes(), &grouped); err != nil {
		return nil, fmt.Errorf("failed to parse `pnpm licenses list --json` output: %w", err)
	}

	// Every path, not one per entry: entries are grouped by the license PNPM
	// read, so one can carry several VERSIONS of a package, each at its own
	// path. Keeping only the first would leave a version unread — and a version
	// whose terms differ from its sibling's is exactly what a license scan
	// exists to catch.
	paths := &bytes.Buffer{}
	for _, entries := range grouped {
		for _, entry := range entries {
			for _, path := range entry.Paths {
				paths.WriteString(path)
				paths.WriteByte('\n')
			}
		}
	}
	if resolver.listedWorkspaces == nil {
		resolver.listedWorkspaces = make(map[string]bool)
	}
	resolver.listedWorkspaces[root] = true
	return paths, nil
}

// GetInstalledPkgs gathers all the installed packages' names and paths.
//
// The node_modules directory NpmResolver is handed is ignored: pnpm reports
// absolute paths, and its packages live under `node_modules/.pnpm` rather than
// directly beneath it, so there is nothing to resolve them against. The
// parameter is kept so this shadows the embedded method — a caller reaching
// GetInstalledPkgs on a pnpm project must not silently get npm's listing.
func (resolver *PnpmResolver) GetInstalledPkgs(_ string) []*Package {
	buffer, err := resolver.ListPkgPaths()
	if err != nil {
		logger.Log.Errorln(err)
		return nil
	}
	pkgs := make([]*Package, 0)
	sc := bufio.NewScanner(buffer)
	for sc.Scan() {
		absPath := sc.Text()
		name := pnpmPkgName(absPath)
		if name == "" {
			continue
		}
		pkgs = append(pkgs, &Package{
			Name: name,
			Path: absPath,
		})
	}
	return pkgs
}

// pnpmPkgName recovers a package's name from where pnpm installed it.
//
// npm installs to `node_modules/<name>`, so npm infers the name from the path
// relative to the node_modules directory. pnpm links out of its store, to
// `node_modules/.pnpm/<name>@<version>/node_modules/<name>` — where that same
// relative path yields the store key instead. The name is what follows the LAST
// node_modules segment, which is true of both layouts.
func pnpmPkgName(absPath string) string {
	normalized := filepath.ToSlash(absPath)
	const marker = "/node_modules/"
	if i := strings.LastIndex(normalized, marker); i >= 0 {
		return normalized[i+len(marker):]
	}
	return ""
}

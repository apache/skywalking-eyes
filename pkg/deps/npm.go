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
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/license"
)

// Constants for architecture names to avoid string duplication
const (
	archAMD64 = "amd64"
	archARM64 = "arm64"
	archARM   = "arm"
)

// Cross-platform package pattern recognition (for precise matching)
// These patterns work for both scoped (@scope/package-platform-arch) and
// non-scoped (package-platform-arch) npm packages, as the platform/arch
// suffix always appears at the end of the package name.
// Examples:
//   - Scoped:    @scope/foo-linux-x64
//   - Non-scoped: foo-linux-x64
//
// regex: matches package names ending with a specific string (e.g., "-linux-x64")
// os: target operating system (e.g., "linux", "darwin", "windows")
// arch: target CPU architecture (e.g., "x64", "arm64")
var platformPatterns = []struct {
	regex *regexp.Regexp
	os    string
	arch  string
}{
	// Android
	{regexp.MustCompile(`-android-arm64$`), "android", archARM64},
	{regexp.MustCompile(`-android-arm$`), "android", archARM},
	{regexp.MustCompile(`-android-x64$`), "android", "x64"},

	// Darwin (macOS)
	{regexp.MustCompile(`-darwin-arm64$`), "darwin", archARM64},
	{regexp.MustCompile(`-darwin-x64$`), "darwin", "x64"},

	// Linux
	{regexp.MustCompile(`-linux-arm64-glibc$`), "linux", archARM64},
	{regexp.MustCompile(`-linux-arm64-musl$`), "linux", archARM64},
	{regexp.MustCompile(`-linux-arm-glibc$`), "linux", archARM},
	{regexp.MustCompile(`-linux-arm-musl$`), "linux", archARM},
	{regexp.MustCompile(`-linux-x64-glibc$`), "linux", "x64"},
	{regexp.MustCompile(`-linux-x64-musl$`), "linux", "x64"},
	{regexp.MustCompile(`-linux-x64$`), "linux", "x64"},
	{regexp.MustCompile(`-linux-arm64$`), "linux", archARM64},
	{regexp.MustCompile(`-linux-arm$`), "linux", archARM},

	// Windows
	{regexp.MustCompile(`-win32-arm64$`), "windows", archARM64},
	{regexp.MustCompile(`-win32-ia32$`), "windows", "ia32"},
	{regexp.MustCompile(`-win32-x64$`), "windows", "x64"},

	// FreeBSD
	{regexp.MustCompile(`-freebsd-x64$`), "freebsd", "x64"},
}

type NpmResolver struct {
	Resolver
}

// Lcs represents the license style in package.json
type Lcs struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Package represents package.json
// License field has inconsistent styles, so we just store the byte array here to postpone unmarshalling
type Package struct {
	Name     string          `json:"name"`
	License  json.RawMessage `json:"license"`
	Licenses []Lcs           `json:"licenses"`
	Path     string          `json:"-"`
	Version  string          `json:"version"`
}

const PkgFileName = "package.json"

// CanResolve checks whether the given file is the npm package file
func (resolver *NpmResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return base == PkgFileName
}

// Resolve resolves licenses of all dependencies declared in the package.json file.
func (resolver *NpmResolver) Resolve(pkgFile string, config *ConfigDeps, report *Report) error {
	workDir := filepath.Dir(pkgFile)
	if err := os.Chdir(workDir); err != nil {
		return err
	}

	// Run command 'npm install' to install or update the required node packages
	// Query from the command line first whether to skip this procedure,
	// in case that the dependent packages are downloaded and brought up-to-date
	if needSkip := resolver.NeedSkipInstallPkgs(); !needSkip {
		resolver.InstallPkgs()
	}

	// Run command 'npm ls --all --parseable' to list all the installed packages' paths
	// Use a package directory's relative path from the node_modules directory, to infer its package name
	// Thus gathering all the installed packages' names and paths
	pkgDir := filepath.Join(workDir, "node_modules")
	pkgs := resolver.GetInstalledPkgs(pkgDir)

	// Walk through each package's root directory to resolve licenses
	// Resolve from a package's package.json file or its license file
	for _, pkg := range pkgs {
		if result := resolver.ResolvePackageLicense(pkg.Name, pkg.Path, config); result.LicenseSpdxID != "" {
			report.Resolve(result)
		} else if result.IsCrossPlatform {
			logger.Log.Warnf("Skipping cross-platform package %s (not for current platform %s %s)", pkg.Name, runtime.GOOS, runtime.GOARCH)
		} else {
			result.LicenseSpdxID = Unknown
			report.Skip(result)
			logger.Log.Warnln("Failed to resolve the license of dependency:", pkg.Name, result.ResolveErrors)
		}
	}
	return nil
}

// NeedSkipInstallPkgs queries whether to skip the procedure of installing or updating packages
func (resolver *NpmResolver) NeedSkipInstallPkgs() bool {
	const countdown = 5
	input := make(chan rune)
	logger.Log.Infoln(fmt.Sprintf("Try to install nodejs packages in %v seconds, press [s/S] and ENTER to skip", countdown))

	// Read the input character from console in a non-blocking way
	go func(ch chan rune) {
		reader := bufio.NewReader(os.Stdin)
		c, _, err := reader.ReadRune()
		if err != nil {
			close(ch)
			return
		}
		input <- c
	}(input)

	// Wait for the user to input a character or the countdown timer to elapse
	select {
	case input, ok := <-input:
		if ok && (input == 's' || input == 'S') {
			return true
		}
		logger.Log.Infoln("Unknown input, try to install packages")
		return false
	case <-time.After(countdown * time.Second):
		logger.Log.Infoln("Time out, try to install packages")
		return false
	}
}

// InstallPkgs runs command 'npm ci' to install node packages,
// using `npm ci` instead of `npm install` to ensure the reproducible builds.
// See https://blog.npmjs.org/post/171556855892/introducing-npm-ci-for-faster-more-reliable
func (resolver *NpmResolver) InstallPkgs() {
	cmd := exec.Command("npm", "ci")
	logger.Log.Println(fmt.Sprintf("Run command: %v, please wait", cmd.String()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Error occurs all the time in npm commands, so no return statement here
	if err := cmd.Run(); err != nil {
		logger.Log.Errorln(err)
	}
}

// ListPkgPaths runs npm command to list all the production only packages' absolute paths, one path per line
// Note that although the flag `--long` can show more information line like a package's name,
// its realization and printing format is not uniform in different npm-cli versions
func (resolver *NpmResolver) ListPkgPaths() (io.Reader, error) {
	pruneCmd := exec.Command("npm", "prune", "--production")
	pruneCmd.Stderr = io.Discard
	pruneCmd.Stdout = io.Discard
	if err := pruneCmd.Run(); err != nil {
		logger.Log.Debug("Failed to prune npm packages")
	}

	buffer := &bytes.Buffer{}
	cmd := exec.Command("npm", "ls", "--all", "--production", "--parseable")
	cmd.Stderr = os.Stderr
	cmd.Stdout = buffer
	// Error occurs all the time in npm commands, so no return statement here
	err := cmd.Run()
	return buffer, err
}

// GetInstalledPkgs gathers all the installed packages' names and paths
// it uses a package directory's relative path from the node_modules directory, to infer its package name
func (resolver *NpmResolver) GetInstalledPkgs(pkgDir string) []*Package {
	buffer, err := resolver.ListPkgPaths()
	// Error occurs all the time in npm commands, so no return statement here
	if err != nil {
		logger.Log.Errorln(err)
	}
	pkgs := make([]*Package, 0)
	sc := bufio.NewScanner(buffer)
	for sc.Scan() {
		absPath := sc.Text()
		rel, err := filepath.Rel(pkgDir, absPath)
		if err != nil {
			logger.Log.Errorln(err)
			continue
		}
		if rel == "" || rel == "." || rel == ".." {
			continue
		}
		pkgName := filepath.ToSlash(rel)
		pkgs = append(pkgs, &Package{
			Name: pkgName,
			Path: absPath,
		})
	}
	return pkgs
}

// ResolvePackageLicense resolves the licenses of the given packages.
// First, try to find and parse the package's package.json file to check the license file
// If the previous step fails, then try to identify the package's LICENSE file
// It's a necessary procedure to check the LICENSE file, because the resolver needs to record the license content
func (resolver *NpmResolver) ResolvePackageLicense(pkgName, pkgPath string, config *ConfigDeps) *Result {
	result := &Result{
		Dependency: pkgName,
	}

	if !resolver.isForCurrentPlatform(pkgName) {
		result.IsCrossPlatform = true
		return result
	}

	// resolve from the package.json file
	if err := resolver.ResolvePkgFile(result, pkgPath, config); err != nil {
		result.ResolveErrors = append(result.ResolveErrors, err)
	}

	// resolve from the LICENSE file
	if err := resolver.ResolveLcsFile(result, pkgPath, config); err != nil {
		result.ResolveErrors = append(result.ResolveErrors, err)
	}

	return result
}

// ResolvePkgFile tries to find and parse the package.json file to capture the license field
func (resolver *NpmResolver) ResolvePkgFile(result *Result, pkgPath string, config *ConfigDeps) error {
	expectedPkgFile := filepath.Join(pkgPath, PkgFileName)
	packageInfo, err := resolver.ParsePkgFile(expectedPkgFile)
	if err != nil {
		return err
	}

	result.Version = packageInfo.Version
	if l, ok := config.GetUserConfiguredLicense(packageInfo.Name, packageInfo.Version); ok {
		result.LicenseSpdxID = l
		return nil
	}

	if lcs, ok := resolver.ResolveLicenseField(packageInfo.License); ok {
		result.LicenseSpdxID = lcs
		return nil
	}

	if lcs, ok := resolver.ResolveLicensesField(packageInfo.Licenses); ok {
		result.LicenseSpdxID = lcs
		return nil
	}

	return fmt.Errorf(`cannot parse the "license"/"licenses" field`)
}

// ResolveLicenseField parses and validates the "license" field in package.json file
func (resolver *NpmResolver) ResolveLicenseField(rawData []byte) (string, bool) {
	if len(rawData) > 0 {
		switch rawData[0] {
		case '"':
			var lcs string
			_ = json.Unmarshal(rawData, &lcs)
			if lcs != "" {
				return lcs, true
			}
		case '{':
			var lcs Lcs
			_ = json.Unmarshal(rawData, &lcs)
			if t := lcs.Type; t != "" {
				return t, true
			}
		}
	}
	return "", false
}

// ResolveLicensesField parses and validates the "licenses" field in package.json file
// Additionally, the output is converted into the SPDX license expression syntax version 2.0 string, like "ISC OR GPL-3.0"
func (resolver *NpmResolver) ResolveLicensesField(licenses []Lcs) (string, bool) {
	var lcs []string
	for _, l := range licenses {
		lcs = append(lcs, l.Type)
	}
	if len(lcs) == 0 {
		return "", false
	}
	return strings.Join(lcs, " OR "), true
}

// ResolveLcsFile tries to find the license file to identify the license
func (resolver *NpmResolver) ResolveLcsFile(result *Result, pkgPath string, config *ConfigDeps) error {
	depFiles, err := os.ReadDir(pkgPath)
	if err != nil {
		return err
	}
	for _, info := range depFiles {
		if info.IsDir() || !possibleLicenseFileName.MatchString(info.Name()) {
			continue
		}
		licenseFilePath := filepath.Join(pkgPath, info.Name())
		result.LicenseFilePath = licenseFilePath
		content, err := os.ReadFile(licenseFilePath)
		if err != nil {
			return err
		}
		result.LicenseContent = string(content)
		if result.LicenseSpdxID != "" {
			return nil
		}
		if l, ok := config.GetUserConfiguredLicense(info.Name(), result.Version); ok {
			result.LicenseSpdxID = l
			return nil
		}
		identifier, err := license.Identify(string(content), config.Threshold)
		if err != nil {
			return err
		}
		result.LicenseSpdxID = identifier
		return nil
	}
	return fmt.Errorf("cannot find the license file")
}

// ParsePkgFile parses the content of the package file
func (resolver *NpmResolver) ParsePkgFile(pkgFile string) (*Package, error) {
	content, err := os.ReadFile(pkgFile)
	if err != nil {
		return nil, err
	}
	var packageInfo Package
	if err := json.Unmarshal(content, &packageInfo); err != nil {
		return nil, err
	}
	return &packageInfo, nil
}

// normalizeArch converts various architecture aliases into Go's canonical naming.
func normalizeArch(arch string) string {
	// Convert to lowercase to handle case variations (e.g., "AMD64").
	arch = strings.ToLower(arch)
	switch arch {
	// x86-64 family (64-bit Intel/AMD)
	case "x64", "x86_64", "amd64", "x86-64":
		return archAMD64
	// x86 32-bit family (legacy)
	case "ia32", "x86", "386", "i386", "i686":
		return "386"
	// ARM 64-bit
	case "arm64", "aarch64":
		return archARM64
	// ARM 32-bit
	case "arm", "armv7", "armhf", "armv7l", "armel":
		return archARM
	// Unknown architecture: return as-is (alternatively, could return empty to indicate incompatibility).
	default:
		return arch
	}
}

// analyzePackagePlatform extracts the target OS and architecture from a package name.
func (resolver *NpmResolver) analyzePackagePlatform(pkgName string) (pkgOS, pkgArch string) {
	for _, pattern := range platformPatterns {
		if pattern.regex.MatchString(pkgName) {
			return pattern.os, pattern.arch
		}
	}
	return "", ""
}

// isForCurrentPlatform checks whether the package is intended for the current OS and architecture.
func (resolver *NpmResolver) isForCurrentPlatform(pkgName string) bool {
	pkgPlatform, pkgArch := resolver.analyzePackagePlatform(pkgName)
	// If no platform/arch info is embedded in the package name, assume it's universal.
	if pkgPlatform == "" && pkgArch == "" {
		return true
	}

	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH

	// The package matches only if both OS and architecture are compatible.
	return pkgPlatform == currentOS && resolver.isArchCompatible(pkgArch, currentArch)
}

// isArchCompatible determines whether the package's architecture is compatible with the current machine's architecture.
func (resolver *NpmResolver) isArchCompatible(pkgArch, currentArch string) bool {
	return normalizeArch(pkgArch) == normalizeArch(currentArch)
}

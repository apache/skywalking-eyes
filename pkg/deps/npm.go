//
// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg/license"
)

type NpmResolver struct {
	Resolver
}

// Package represents package.json
type Package struct {
	Name    string `json:"name"`
	License string `json:"license"`
	Path    string
}

const PkgFileName = "package.json"

// CanResolve checks whether the given file is the npm package file
func (resolver *NpmResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return base == PkgFileName
}

// Resolve resolves licenses of all dependencies declared in the package.json file.
func (resolver *NpmResolver) Resolve(pkgFile string, report *Report) error {
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
		if err := resolver.ResolvePackageLicense(pkg.Name, pkg.Path, report); err != nil {
			logger.Log.Warnln("Failed to resolve the license of dependency:", pkg.Name, err)
			report.Skip(&Result{
				Dependency:    pkg.Name,
				LicenseSpdxID: Unknown,
			})
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

// InstallPkgs runs command 'npm install' to install node packages
func (resolver *NpmResolver) InstallPkgs() {
	cmd := exec.Command("npm", "install")
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
func (resolver *NpmResolver) ResolvePackageLicense(pkgName, pkgPath string, report *Report) error {
	var resolveErrs error
	expectedPkgFile := filepath.Join(pkgPath, PkgFileName)
	lcs, err := resolver.ResolvePkgFile(expectedPkgFile)
	if err == nil {
		report.Resolve(&Result{
			Dependency:    pkgName,
			LicenseSpdxID: lcs,
		})
		return nil
	}
	resolveErrs = err

	lcs, err = resolver.ResolveLcsFile(pkgName, pkgPath)
	if err == nil {
		report.Resolve(&Result{
			Dependency:    pkgName,
			LicenseSpdxID: lcs,
		})
		return nil
	}
	resolveErrs = fmt.Errorf("%+v; %+v", resolveErrs, err)
	return resolveErrs
}

// ResolvePkgFile tries to find and parse the package.json file to capture the license field
func (resolver *NpmResolver) ResolvePkgFile(pkgFile string) (string, error) {
	packageInfo, err := resolver.ParsePkgFile(pkgFile)
	if err != nil {
		return "", err
	}
	if packageInfo.License == "" {
		return "", fmt.Errorf("cannot capture the license field")
	}
	return packageInfo.License, nil
}

// ResolveLcsFile tries to find the license file to identify the license
func (resolver *NpmResolver) ResolveLcsFile(pkgName, pkgPath string) (string, error) {
	depFiles, err := ioutil.ReadDir(pkgPath)
	if err != nil {
		return "", err
	}
	for _, info := range depFiles {
		if info.IsDir() || !possibleLicenseFileName.MatchString(info.Name()) {
			continue
		}
		licenseFilePath := filepath.Join(pkgPath, info.Name())
		content, err := ioutil.ReadFile(licenseFilePath)
		if err != nil {
			return "", err
		}
		identifier, err := license.Identify(pkgName, string(content))
		if err != nil {
			return "", err
		}
		return identifier, nil
	}
	return "", fmt.Errorf("cannot find the license file")
}

// ParsePkgFile parses the content of the package file
func (resolver *NpmResolver) ParsePkgFile(pkgFile string) (*Package, error) {
	content, err := ioutil.ReadFile(pkgFile)
	if err != nil {
		return nil, err
	}
	var packageInfo Package
	if err := json.Unmarshal(content, &packageInfo); err != nil {
		return nil, err
	}
	return &packageInfo, nil
}
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
	"encoding/json"
	"errors"
	"fmt"
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
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	License      string            `json:"license"`
	Dependencies map[string]string `json:"dependencies"`
}

const packageFile = "package.json"

// CanResolve checks whether the given file is the npm package file
func (resolver *NpmResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return base == packageFile
}

// Resolve resolves licenses of all dependencies declared in the package.json file.
func (resolver *NpmResolver) Resolve(packageFile string, report *Report) error {
	// Parse the project package file to gather the required dependencies
	packageInfo, err := resolver.parsePackageFile(packageFile)
	if err != nil {
		return err
	}
	depNames := make([]string, 0, len(packageInfo.Dependencies))
	for dep := range packageInfo.Dependencies {
		depNames = append(depNames, dep)
	}

	// Run command 'npm install' to install or update the required node packages
	// Query from the command line first whether to skip this procedure
	// in case that the dependent packages are downloaded and brought up-to-date
	root := filepath.Dir(packageFile)
	if needSkip := resolver.NeedSkipInstallPkgs(); !needSkip {
		if err := resolver.InstallPkgs(root); err != nil {
			return fmt.Errorf("fail to install depNames: %+v", err)
		}
	}

	// Change working directory to node_modules
	depPath := filepath.Join(root, "node_modules")
	if err := os.Chdir(depPath); err != nil {
		return err
	}

	// Walk through each package's root directory to resolve licenses
	// First, try to parse the package's package.json file to check the license file
	// If the previous step fails, then try to identify the package's LICENSE file
	for _, depName := range depNames {
		if err := resolver.ResolvePackageLicense(depName, report); err != nil {
			logger.Log.Warnln("Failed to resolve the license of dependency:", depName, err)
			report.Skip(&Result{
				Dependency:    depName,
				LicenseSpdxID: Unknown,
			})
		}
	}
	return nil
}

// parsePackageFile parses the content of the package file
func (resolver *NpmResolver) parsePackageFile(packageFile string) (*Package, error) {
	content, err := ioutil.ReadFile(packageFile)
	if err != nil {
		return nil, err
	}
	var packageInfo Package
	if err := json.Unmarshal(content, &packageInfo); err != nil {
		return nil, err
	}
	return &packageInfo, nil
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
func (resolver *NpmResolver) InstallPkgs(root string) error {
	cmd := exec.Command("npm", "install")
	cmd.Dir = root
	logger.Log.Println(fmt.Sprintf("Run command: %v, please wait", cmd.String()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// ResolvePackageLicense resolves the licenses of the given packages.
func (resolver *NpmResolver) ResolvePackageLicense(depName string, report *Report) error {
	depFiles, err := ioutil.ReadDir(depName)
	if err != nil {
		return err
	}

	// Record the errors encountered when parsing the package.json file
	packageErr := errors.New("cannot find the package.json file")

	// STEP 1: Try to find and parse the package.json file to capture the license field
	for _, info := range depFiles {
		if info.IsDir() || info.Name() != packageFile {
			continue
		}
		packageFilePath := filepath.Join(depName, info.Name())
		packageInfo, err := resolver.parsePackageFile(packageFilePath)
		if err != nil {
			packageErr = err
			break
		}
		if packageInfo.License == "" {
			packageErr = fmt.Errorf("cannot capture the license field")
			break
		}
		report.Resolve(&Result{
			Dependency:      depName,
			LicenseFilePath: "",
			LicenseContent:  "",
			LicenseSpdxID:   packageInfo.License,
		})
		return nil
	}

	// Record the errors encountered when identifying the license file
	licenseErr := errors.New("cannot find the license file")

	// STEP 2: Try to find the license file to identify the license
	for _, info := range depFiles {
		if info.IsDir() || !possibleLicenseFileName.MatchString(info.Name()) {
			continue
		}
		licenseFilePath := filepath.Join(depName, info.Name())
		content, err := ioutil.ReadFile(licenseFilePath)
		if err != nil {
			licenseErr = err
			break
		}
		identifier, err := license.Identify(depName, string(content))
		if err != nil {
			licenseErr = err
			break
		}
		report.Resolve(&Result{
			Dependency:      depName,
			LicenseFilePath: licenseFilePath,
			LicenseContent:  string(content),
			LicenseSpdxID:   identifier,
		})
		return nil
	}

	return fmt.Errorf("%+v; %+v", packageErr, licenseErr)
}

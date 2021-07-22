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

func (resolver *NpmResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return base == "package.json"
}

// Resolve resolves licenses of all dependencies declared in the package.json file.
func (resolver *NpmResolver) Resolve(packageFile string, report *Report) error {
	deps, err := resolver.parseDeps(packageFile)
	if err != nil {
		return err
	}

	root := filepath.Dir(packageFile)
	if skip := resolver.QuerySkipInstallPkgs(); !skip {
		if err := resolver.InstallPkgs(root); err != nil {
			logger.Log.Errorln("Fail to install packages")
		}
	}

	depPath := filepath.Join(root, "node_modules")
	if err := os.Chdir(depPath); err != nil {
		return err
	}

	for _, dep := range deps {
		if err := resolver.ResolvePackageLicense(dep, report); err != nil {
			logger.Log.Warnln("Failed to resolve the license of dependency:", dep, err)
			report.Skip(&Result{
				Dependency:    dep,
				LicenseSpdxID: Unknown,
			})
		}
	}
	return nil
}

// parseDeps parses the content of the package file
func (resolver *NpmResolver) parseDeps(packageFile string) ([]string, error) {
	content, err := ioutil.ReadFile(packageFile)
	if err != nil {
		return nil, err
	}
	var packageInfo Package
	if err := json.Unmarshal(content, &packageInfo); err != nil {
		return nil, err
	}
	depNames := make([]string, 0, len(packageInfo.Dependencies))
	for dep := range packageInfo.Dependencies {
		depNames = append(depNames, dep)
	}
	return depNames, nil
}

// QuerySkipInstallPkgs queries whether to skip the procedure of installing or updating packages
func (resolver *NpmResolver) QuerySkipInstallPkgs() bool {
	const countdown = 5
	logger.Log.Infoln(fmt.Sprintf("Try to install nodejs packages in %v seconds, press [s/S] and ENTER to skip", countdown))
	input := make(chan rune)
	go func(ch chan rune) {
		reader := bufio.NewReader(os.Stdin)
		c, _, err := reader.ReadRune()
		if err != nil {
			close(ch)
			return
		}
		input <- c
	}(input)

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

// InstallPkgs runs command to install packages
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
func (resolver *NpmResolver) ResolvePackageLicense(dep string, report *Report) error {
	depFiles, err := ioutil.ReadDir(dep)
	if err != nil {
		return err
	}
	for _, info := range depFiles {
		if info.IsDir() || !possibleLicenseFileName.MatchString(info.Name()) {
			continue
		}
		licenseFilePath := filepath.Join(dep, info.Name())
		content, err := ioutil.ReadFile(licenseFilePath)
		if err != nil {
			continue
		}
		identifier, err := license.Identify(dep, string(content))
		if err != nil {
			return err
		}
		report.Resolve(&Result{
			Dependency:      dep,
			LicenseFilePath: licenseFilePath,
			LicenseContent:  string(content),
			LicenseSpdxID:   identifier,
		})
		return nil
	}
	return errors.New("cannot find the license description file")
}

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
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/pkg/license"
	"github.com/apache/skywalking-eyes/pkg/logger"

	"golang.org/x/tools/go/packages"
)

type GoModResolver struct {
	Resolver
}

const (
	goModFileName = "go.mod"
)

var (
	goModuleDirective       = regexp.MustCompile(`(?m)^\s*module\s+\S`)
	possibleLicenseFileName = regexp.MustCompile(`(?i)^(LICENSE|LICENCE|COPYING)(\.txt)?$`)
)

func (resolver *GoModResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return strings.HasSuffix(base, ".mod")
}

func validateGoModFile(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if !goModuleDirective.Match(content) {
		return fmt.Errorf("%v is not a valid Go module file", file)
	}
	return nil
}

// Resolve resolves licenses of all dependencies declared in the go.mod file.
func (resolver *GoModResolver) Resolve(goModFile string, config *ConfigDeps, report *Report) error {
	if err := validateGoModFile(goModFile); err != nil {
		return err
	}

	if err := os.Chdir(filepath.Dir(goModFile)); err != nil {
		return err
	}

	base := filepath.Base(goModFile)
	downloadArgs := []string{"mod", "download"}
	jsonArgs := []string{"mod", "download", "-json"}
	if base != goModFileName {
		downloadArgs = append(downloadArgs, "-modfile", base)
		jsonArgs = append(jsonArgs, "-modfile", base)
	}

	goModDownload := exec.Command("go", downloadArgs...)
	logger.Log.Debugf("Run command: %v, please wait", goModDownload.String())
	goModDownload.Stdout = os.Stdout
	goModDownload.Stderr = os.Stderr
	if err := goModDownload.Run(); err != nil {
		return err
	}

	output, err := exec.Command("go", jsonArgs...).Output()
	if err != nil {
		return err
	}

	modules := make([]*packages.Module, 0)
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var m packages.Module
		if err := decoder.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		modules = append(modules, &m)
	}

	logger.Log.Debugln("Module size:", len(modules))

	return resolver.ResolvePackages(modules, config, report)
}

// ResolvePackages resolves the licenses of the given packages.
func (resolver *GoModResolver) ResolvePackages(modules []*packages.Module, config *ConfigDeps, report *Report) error {
	for _, module := range modules {
		func() {
			if excluded, _ := config.IsExcluded(module.Path, module.Version); excluded {
				return
			}
			if l, ok := config.GetUserConfiguredLicense(module.Path, module.Version); ok {
				report.Resolve(&Result{
					Dependency:    module.Path,
					LicenseSpdxID: l,
					Version:       module.Version,
				})
				return
			}
			err := resolver.ResolvePackageLicense(config, module, report)
			if err != nil {
				logger.Log.Warnf("Failed to resolve the license of <%s@%s>: %v\n", module.Path, module.Version, err)
				report.Skip(&Result{
					Dependency:    module.Path,
					LicenseSpdxID: Unknown,
					Version:       module.Version,
				})
			}
		}()
	}
	return nil
}

func (resolver *GoModResolver) ResolvePackageLicense(config *ConfigDeps, module *packages.Module, report *Report) error {
	dir := module.Dir

	for {
		logger.Log.Debugf("Directory of %+v is %+v", module.Path, dir)
		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, info := range files {
			if info.IsDir() || !possibleLicenseFileName.MatchString(info.Name()) {
				continue
			}
			licenseFilePath := filepath.Join(dir, info.Name())
			content, err := os.ReadFile(licenseFilePath)
			if err != nil {
				return err
			}
			identifier, err := license.Identify(string(content), config.Threshold)
			if err != nil {
				return err
			}

			logger.Log.Debugf("\t- Found license: %v", identifier)
			report.Resolve(&Result{
				Dependency:      module.Path,
				LicenseFilePath: licenseFilePath,
				LicenseContent:  string(content),
				LicenseSpdxID:   identifier,
				Version:         module.Version,
			})
			return nil
		}
		if resolver.shouldStopAt(dir, module.Dir) {
			break
		}
		dir = filepath.Dir(dir)
	}
	return fmt.Errorf("cannot find license file")
}

func (resolver *GoModResolver) shouldStopAt(dir, moduleDir string) bool {
	return dir == moduleDir || dir == build.Default.GOPATH
}

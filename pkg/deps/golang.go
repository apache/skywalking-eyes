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
	"context"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg/license"
)

type GoModeResolver struct {
	Resolver
}

func (resolver *GoModeResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return base == "go.mod"
}

// Resolve resolves licenses of all dependencies declared in the go.mod file.
func (resolver *GoModeResolver) Resolve(goModFile string, report *Report) error {
	content, err := ioutil.ReadFile(goModFile)
	if err != nil {
		return err
	}

	file, err := modfile.Parse(goModFile, content, nil)
	if err != nil {
		return err
	}

	logger.Log.Debugln("Resolving module:", file.Module.Mod)

	if err := os.Chdir(filepath.Dir(goModFile)); err != nil {
		return err
	}

	requiredPkgNames := make([]string, len(file.Require))
	for i, require := range file.Require {
		requiredPkgNames[i] = require.Mod.Path
	}

	logger.Log.Debugln("Required packages:", requiredPkgNames)

	if err := resolver.ResolvePackages(requiredPkgNames, report); err != nil {
		return err
	}

	return nil
}

// ResolvePackages resolves the licenses of the given packages.
func (resolver *GoModeResolver) ResolvePackages(pkgNames []string, report *Report) error {
	requiredPkgs, err := packages.Load(&packages.Config{
		Context: context.Background(),
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps,
	}, pkgNames...)

	if err != nil {
		return err
	}

	packages.Visit(requiredPkgs, func(p *packages.Package) bool {
		err := resolver.ResolvePackageLicense(p, report)
		if err != nil {
			logger.Log.Warnln("Failed to resolve the license of dependency:", p.PkgPath, err)
			report.Skip(&Result{
				Dependency:    p.PkgPath,
				LicenseSpdxID: []string{Unknown},
			})
		}
		return true
	}, nil)

	return nil
}

var possibleLicenseFileName = regexp.MustCompile(`(?i)^LICENSE|LICENCE(\.txt)?$`)

func (resolver *GoModeResolver) ResolvePackageLicense(p *packages.Package, report *Report) error {
	var filesInPkg []string
	if len(p.GoFiles) > 0 {
		filesInPkg = p.GoFiles
	} else if len(p.CompiledGoFiles) > 0 {
		filesInPkg = p.CompiledGoFiles
	} else if len(p.OtherFiles) > 0 {
		filesInPkg = p.OtherFiles
	}

	if len(filesInPkg) == 0 {
		return fmt.Errorf("empty package")
	}

	absPath, err := filepath.Abs(filesInPkg[0])
	if err != nil {
		return err
	}
	dir := filepath.Dir(absPath)

	for {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, info := range files {
			if !possibleLicenseFileName.MatchString(info.Name()) {
				continue
			}
			licenseFilePath := filepath.Join(dir, info.Name())
			content, err := ioutil.ReadFile(licenseFilePath)
			if err != nil {
				return err
			}
			identifier, err := license.Identify(string(content))
			if err != nil {
				return err
			}
			report.Resolve(&Result{
				Dependency:      p.PkgPath,
				LicenseFilePath: licenseFilePath,
				LicenseContent:  string(content),
				LicenseSpdxID:   []string{identifier},
			})
			return nil
		}
		if resolver.shouldStopAt(dir) {
			break
		}
		dir = filepath.Dir(dir)
	}
	return nil
}

func (resolver *GoModeResolver) shouldStopAt(dir string) bool {
	for _, srcDir := range build.Default.SrcDirs() {
		if srcDir == dir {
			return true
		}
	}
	return false
}

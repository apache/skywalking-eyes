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
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/license"
)

type CargoMetadata struct {
	Packages []CargoPackage `json:"packages"`
}

type CargoPackage struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	License      string `json:"license"`
	LicenseFile  string `json:"license_file"`
	ManifestPath string `json:"manifest_path"`
}

type CargoTomlResolver struct {
	Resolver
}

func (resolver *CargoTomlResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return base == "Cargo.toml"
}

// Resolve resolves licenses of all dependencies declared in the Cargo.toml file.
func (resolver *CargoTomlResolver) Resolve(cargoTomlFile string, config *ConfigDeps, report *Report) error {
	dir := filepath.Dir(cargoTomlFile)

	download := exec.Command("cargo", "fetch")
	logger.Log.Debugf("Run command: %v, please wait", download.String())
	download.Stdout = os.Stdout
	download.Stderr = os.Stderr
	download.Dir = dir
	if err := download.Run(); err != nil {
		return err
	}

	cmd := exec.Command("cargo", "metadata", "--format-version=1", "--all-features")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	var metadata CargoMetadata
	if err := json.Unmarshal(output, &metadata); err != nil {
		return err
	}

	logger.Log.Debugln("Package size:", len(metadata.Packages))

	return resolver.ResolvePackages(metadata.Packages, config, report)
}

// ResolvePackages resolves the licenses of the given packages.
func (resolver *CargoTomlResolver) ResolvePackages(packages []CargoPackage, config *ConfigDeps, report *Report) error {
	for i := range packages {
		pkg := packages[i]

		if exclude, _ := config.IsExcluded(pkg.Name, pkg.Version); exclude {
			continue
		}
		if l, ok := config.GetUserConfiguredLicense(pkg.Name, pkg.Version); ok {
			report.Resolve(&Result{
				Dependency:    pkg.Name,
				LicenseSpdxID: l,
				Version:       pkg.Version,
			})
			continue
		}
		err := resolver.ResolvePackageLicense(config, &pkg, report)
		if err != nil {
			logger.Log.Warnf("Failed to resolve the license of <%s@%s>: %v\n", pkg.Name, pkg.Version, err)
			report.Skip(&Result{
				Dependency:    pkg.Name,
				LicenseSpdxID: Unknown,
				Version:       pkg.Version,
			})
		}
	}
	return nil
}

var cargoPossibleLicenseFileName = regexp.MustCompile(`(?i)^LICENSE|LICENCE(\.txt)?|LICENSE-.+|COPYING(\.txt)?$`)

// ResolvePackageLicense resolve the package license.
// The CargoPackage.LicenseFile is generally used for non-standard licenses and is ignored now.
func (resolver *CargoTomlResolver) ResolvePackageLicense(config *ConfigDeps, pkg *CargoPackage, report *Report) error {
	dir := filepath.Dir(pkg.ManifestPath)
	logger.Log.Debugf("Directory of %+v is %+v", pkg.Name, dir)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var licenseFilePath string
	var licenseContent []byte

	licenseID := pkg.License

	for _, info := range files {
		if !cargoPossibleLicenseFileName.MatchString(info.Name()) {
			continue
		}

		licenseFilePath = filepath.Join(dir, info.Name())
		licenseContent, err = os.ReadFile(licenseFilePath)
		if err != nil {
			return err
		}

		break
	}

	if licenseID == "" { // If pkg.License is empty, identify the license ID from the license file content
		if licenseID, err = license.Identify(string(licenseContent), config.Threshold); err != nil {
			return err
		}
	}

	report.Resolve(&Result{
		Dependency:      pkg.Name,
		LicenseFilePath: licenseFilePath,
		LicenseContent:  string(licenseContent),
		LicenseSpdxID:   licenseID,
		Version:         pkg.Version,
	})

	return nil
}

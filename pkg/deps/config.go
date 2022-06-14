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
	"os"
	"path/filepath"
	"strings"
)

// DefaultCoverageThreshold is the minimum percentage of the file
// that must contain license text for identifying a license.
// Reference: https://github.com/golang/pkgsite/blob/d43359e3a135fc391960db4f5800eb081d658412/internal/licenses/licenses.go#L48
const DefaultCoverageThreshold = 75

type ConfigDeps struct {
	Threshold int                 `yaml:"threshold"`
	Files     []string            `yaml:"files"`
	Licenses  []*ConfigDepLicense `yaml:"licenses"`
	Excludes  []Exclude           `yaml:"excludes"`
}

type ConfigDepLicense struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	License string `yaml:"license"`
}

type Exclude struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (config *ConfigDeps) Finalize(configFile string) error {
	configFileAbsPath, err := filepath.Abs(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for i, file := range config.Files {
		config.Files[i] = filepath.Join(filepath.Dir(configFileAbsPath), file)
	}

	if config.Threshold <= 0 {
		config.Threshold = DefaultCoverageThreshold
	}

	return nil
}

func (config *ConfigDeps) GetUserConfiguredLicense(name, version string) (string, bool) {
	for _, license := range config.Licenses {
		if matched, _ := filepath.Match(license.Name, name); !matched && license.Name != name {
			continue
		}
		if license.Version == "" {
			return license.License, true
		}
		for _, v := range strings.Split(license.Version, ",") {
			if v == version {
				return license.License, true
			}
		}
	}
	return "", false
}

func (config *ConfigDeps) IsExcluded(name, version string) bool {
	for _, license := range config.Excludes {
		if matched, _ := filepath.Match(license.Name, name); !matched && license.Name != name {
			continue
		}
		if license.Version == "" {
			return true
		}
		for _, v := range strings.Split(license.Version, ",") {
			if v == version {
				return true
			}
		}
	}
	return false
}

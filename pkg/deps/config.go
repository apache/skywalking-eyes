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
)

// DefaultCoverageThreshold is the minimum percentage of the file
// that must contain license text for identifying a license.
// Reference: https://github.com/golang/pkgsite/blob/d43359e3a135fc391960db4f5800eb081d658412/internal/licenses/licenses.go#L48
const DefaultCoverageThreshold = 75

type ConfigDeps struct {
	Threshold int                 `yaml:"threshold"`
	Files     []string            `yaml:"files"`
	Licenses  []*ConfigDepLicense `yaml:"licenses"`
}

type ConfigDepLicense struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	License string `yaml:"license"`
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

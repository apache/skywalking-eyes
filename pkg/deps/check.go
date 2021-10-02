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
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
)

type compatibilityMatrix struct {
	Compatible   []string `yaml:"compatible"`
	Incompatible []string `yaml:"incompatible"`
}

var matrices = make(map[string]compatibilityMatrix)

func init() {
	dir := "compatibility"
	files, err := assets.AssetDir(dir)
	if err != nil {
		logger.Log.Fatalln("Failed to list assets/compatibility directory:", err)
	}
	for _, file := range files {
		name := file.Name()
		matrix := compatibilityMatrix{}
		if bytes, err := assets.Asset(filepath.Join(dir, name)); err != nil {
			logger.Log.Fatalln("Failed to read compatibility file:", name, err)
		} else if err := yaml.Unmarshal(bytes, &matrix); err != nil {
			logger.Log.Fatalln("Failed to unmarshal compatibility file:", file, err)
		}
		matrices[strings.TrimSuffix(name, filepath.Ext(name))] = matrix
	}
}

func Check(mainLicenseSpdxID string, config *ConfigDeps) error {
	report := Report{}
	if err := Resolve(config, &report); err != nil {
		return nil
	}

	matrix := matrices[mainLicenseSpdxID]
	var incompatibleResults []*Result
	for _, result := range append(report.Resolved, report.Skipped...) {
		compare := func(list []string) bool {
			for _, com := range list {
				if result.LicenseSpdxID == com {
					return true
				}
			}
			return false
		}
		if compatible := compare(matrix.Compatible); compatible {
			continue
		}
		if incompatible := compare(matrix.Incompatible); incompatible {
			incompatibleResults = append(incompatibleResults, result)
		}
	}

	if len(incompatibleResults) > 0 {
		str := ""
		for _, r := range incompatibleResults {
			str += fmt.Sprintf("\nLicense: %v Dependency: %v", r.LicenseSpdxID, r.Dependency)
		}
		return fmt.Errorf("the following licenses are incompatible with the main license: %v %v", mainLicenseSpdxID, str)
	}

	return nil
}

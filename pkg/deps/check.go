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
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
)

type CompatibilityMatrix struct {
	Compatible   []string `yaml:"compatible"`
	Incompatible []string `yaml:"incompatible"`
}

var matrices = make(map[string]CompatibilityMatrix)

type LicenseOperator int

const (
	LicenseOperatorNone LicenseOperator = iota
	LicenseOperatorAND
	LicenseOperatorOR
	LicenseOperatorWITH
)

func init() {
	dir := "compatibility"
	files, err := assets.AssetDir(dir)
	if err != nil {
		logger.Log.Fatalln("Failed to list assets/compatibility directory:", err)
	}
	for _, file := range files {
		name := file.Name()
		matrix := CompatibilityMatrix{}
		if bytes, err := assets.Asset(filepath.Join(dir, name)); err != nil {
			logger.Log.Fatalln("Failed to read compatibility file:", name, err)
		} else if err := yaml.Unmarshal(bytes, &matrix); err != nil {
			logger.Log.Fatalln("Failed to unmarshal compatibility file:", file, err)
		}
		matrices[strings.TrimSuffix(name, filepath.Ext(name))] = matrix
	}
}

func Check(mainLicenseSpdxID string, config *ConfigDeps) error {
	matrix := matrices[mainLicenseSpdxID]

	report := Report{}
	if err := Resolve(config, &report); err != nil {
		return nil
	}

	return CheckWithMatrix(mainLicenseSpdxID, &matrix, &report)
}

func CheckWithMatrix(mainLicenseSpdxID string, matrix *CompatibilityMatrix, report *Report) error {
	var incompatibleResults []*Result
	for _, result := range append(report.Resolved, report.Skipped...) {
		compare := func(list []string, spdxID string) bool {
			for _, com := range list {
				if spdxID == com {
					return true
				}
			}
			return false
		}
		compareAll := func(spdxIDs []string, compare func(spdxID string) bool) bool {
			for _, spdxID := range spdxIDs {
				if !compare(spdxID) {
					return false
				}
			}
			return true
		}
		compareAny := func(spdxIDs []string, compare func(spdxID string) bool) bool {
			for _, spdxID := range spdxIDs {
				if compare(spdxID) {
					return true
				}
			}
			return false
		}

		operator, spdxIDs := parseLicenseExpression(result.LicenseSpdxID)

		switch operator {
		case LicenseOperatorAND:
			if compareAll(spdxIDs, func(spdxID string) bool {
				return compare(matrix.Compatible, spdxID)
			}) {
				continue
			}
			if compareAny(spdxIDs, func(spdxID string) bool {
				return compare(matrix.Incompatible, spdxID)
			}) {
				incompatibleResults = append(incompatibleResults, result)
			}

		case LicenseOperatorOR:
			if compareAny(spdxIDs, func(spdxID string) bool {
				return compare(matrix.Compatible, spdxID)
			}) {
				continue
			}
			if compareAll(spdxIDs, func(spdxID string) bool {
				return compare(matrix.Incompatible, spdxID)
			}) {
				incompatibleResults = append(incompatibleResults, result)
			}

		default:
			if compatible := compare(matrix.Compatible, spdxIDs[0]); compatible {
				continue
			}
			if incompatible := compare(matrix.Incompatible, spdxIDs[0]); incompatible {
				incompatibleResults = append(incompatibleResults, result)
			}
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

func parseLicenseExpression(s string) (operator LicenseOperator, spdxIDs []string) {
	if ss := strings.Split(s, " AND "); len(ss) > 1 {
		return LicenseOperatorAND, ss
	}
	if ss := strings.Split(s, " and "); len(ss) > 1 {
		return LicenseOperatorAND, ss
	}
	if ss := strings.Split(s, " OR "); len(ss) > 1 {
		return LicenseOperatorOR, ss
	}
	if ss := strings.Split(s, " or "); len(ss) > 1 {
		return LicenseOperatorOR, ss
	}
	if ss := strings.Split(s, " WITH "); len(ss) > 1 {
		return LicenseOperatorWITH, ss
	}
	if ss := strings.Split(s, " with "); len(ss) > 1 {
		return LicenseOperatorWITH, ss
	}
	return LicenseOperatorNone, []string{s}
}

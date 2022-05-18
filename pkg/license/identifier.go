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

package license

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/google/licensecheck"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
)

const (
	// coverageThreshold is the minimum percentage of the file that must contain license text.
	// Reference: https://github.com/golang/pkgsite/blob/d43359e3a135fc391960db4f5800eb081d658412/internal/licenses/licenses.go#L48
	coverageThreshold = 75

	licenseTemplatesDir = "lcs-templates"
)

var (
	_scanner    *licensecheck.Scanner
	scannerOnce sync.Once

	dualLicensePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)This project is covered by two different licenses: (?P<license>[^.]+)`),
	}
)

// scanner returns a licensecheck.Scanner instance with its build-in licenses.
// It will be initialized once.
func scanner() *licensecheck.Scanner {
	scannerOnce.Do(func() {
		var err error
		_scanner, err = licensecheck.NewScanner(licensecheck.BuiltinLicenses())
		if err != nil {
			logger.Log.Fatalf("licensecheck.NewScanner: %v", err)
		}
	})
	return _scanner
}

// Identify identifies the Spdx ID of the given license content.
func Identify(content string) (string, error) {
	for _, pattern := range dualLicensePatterns {
		matches := pattern.FindStringSubmatch(content)
		for i, name := range pattern.SubexpNames() {
			if name == "license" && len(matches) >= i {
				return matches[i], nil
			}
		}
	}

	coverage := scanner().Scan([]byte(content))
	if coverage.Percent < coverageThreshold {
		return "", fmt.Errorf("cannot identify the license, coverage: %.1f%%", coverage.Percent)
	}

	return coverage.Match[0].ID, nil
}

// GetLicenseContent returns the content of the license file with the given Spdx ID.
func GetLicenseContent(spdxID string) (string, error) {
	res, err := assets.Asset(filepath.Join(licenseTemplatesDir, spdxID+".txt"))
	if err != nil {
		return "", err
	}
	return string(res), nil
}

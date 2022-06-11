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
	"strings"
	"sync"

	"github.com/google/licensecheck"
	"gopkg.in/yaml.v3"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
)

const licenseTemplatesDir = "lcs-templates"

var (
	_scanner    *licensecheck.Scanner
	scannerOnce sync.Once
)

// scanner returns a licensecheck.Scanner instance with its build-in licenses.
// It will be initialized once.
func scanner() *licensecheck.Scanner {
	scannerOnce.Do(func() {
		licenses := licensecheck.BuiltinLicenses()
		if bs, err := assets.Asset("urls.yaml"); err == nil {
			licenseURLs := make(map[string][]string)
			if err := yaml.Unmarshal(bs, &licenseURLs); err == nil {
				logger.Log.Debug("license URLs:", licenseURLs)
				for id, urls := range licenseURLs {
					for _, url := range urls {
						licenses = append(licenses, licensecheck.License{
							ID:   id,
							URL:  strings.ToLower(url),
							Type: licensecheck.Unknown,
						})
					}
				}
			}
		}
		var err error
		_scanner, err = licensecheck.NewScanner(licenses)
		if err != nil {
			logger.Log.Fatalf("licensecheck.NewScanner: %v", err)
		}
	})
	return _scanner
}

// Identify identifies the Spdx ID of the given license content.
// If it's a dual-license, it will return `<Licenses 1> and <Licenses 2>`.
func Identify(content string, threshold int) (string, error) {
	coverage := scanner().Scan([]byte(content))
	if coverage.Percent < float64(threshold) {
		return "", fmt.Errorf("cannot identify the license, coverage: %.1f%%", coverage.Percent)
	}

	seen := make(map[string]bool)

	var sb strings.Builder
	sb.WriteString(coverage.Match[0].ID)
	seen[coverage.Match[0].ID] = true

	for i := 1; i < len(coverage.Match); i++ {
		if seen[coverage.Match[i].ID] {
			continue
		}
		seen[coverage.Match[i].ID] = true
		sb.WriteString(" and ")
		sb.WriteString(coverage.Match[i].ID)
	}

	return sb.String(), nil
}

// GetLicenseContent returns the content of the license file with the given Spdx ID.
func GetLicenseContent(spdxID string) (string, error) {
	res, err := assets.Asset(filepath.Join(licenseTemplatesDir, spdxID+".txt"))
	if err != nil {
		return "", err
	}
	return string(res), nil
}

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
package license

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/assets"
	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
)

var templatesDirs = []string{
	"lcs-templates",
	// Some projects simply use the header text as their LICENSE content...
	"header-templates",
}

var dualLicensePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)This project is covered by two different licenses: (?P<license>[^.]+)`),
}

var normalizedTemplates = make(map[string]string)

func init() {
	for _, dir := range templatesDirs {
		files, err := assets.AssetDir(dir)
		if err != nil {
			logger.Log.Fatalln("Failed to read license template directory:", dir, err)
		}
		for _, template := range files {
			name := template.Name()
			t, err := assets.Asset(filepath.Join(dir, name))
			if err != nil {
				logger.Log.Fatalln("Failed to read license template:", dir, err)
			}
			normalizedTemplates[dir+"/"+name] = Normalize(string(t))
		}
	}
}

// Identify identifies the Spdx ID of the given license content
func Identify(pkgPath, content string) (string, error) {
	for _, pattern := range dualLicensePatterns {
		matches := pattern.FindStringSubmatch(content)
		for i, name := range pattern.SubexpNames() {
			if name == "license" && len(matches) >= i {
				return matches[i], nil
			}
		}
	}

	content = Normalize(content)

	for name, license := range normalizedTemplates {
		if strings.Contains(content, license) {
			name = filepath.Base(name)
			return strings.TrimSuffix(name, filepath.Ext(name)), nil
		}
	}

	logger.Log.Debugf("Normalized content for %+v:\n%+v\n", pkgPath, content)

	return "", fmt.Errorf("cannot identify license content")
}

var seemLicense = regexp.MustCompile(`(?i)licen[sc]e|copyright|copying`)

// Seem determine whether the content of the file may be a license file
func Seem(content string) bool {
	return seemLicense.MatchString(content)
}

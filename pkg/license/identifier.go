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

	assets "github.com/apache/skywalking-eyes/license-eye"
)

const templatesDir = "assets/lcs-templates"

var dualLicensePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)This project is covered by two different licenses: (?P<license>[^.]+)`),
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

	templates, err := assets.AssetDir(templatesDir)
	if err != nil {
		return "", err
	}

	for _, template := range templates {
		templateName := template.Name()
		t, err := assets.Asset(filepath.Join(templatesDir, templateName))
		if err != nil {
			return "", err
		}
		license := string(t)
		license = Normalize(license)
		if license == content {
			return strings.TrimSuffix(templateName, filepath.Ext(templateName)), nil
		}
	}

	return "", fmt.Errorf("cannot identify license content")
}

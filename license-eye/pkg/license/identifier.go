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
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/assets"
)

const templatesDir = "assets/lcs-templates"

// Identify identifies the Spdx ID of the given license content
func Identify(content string) (string, error) {
	content = Normalize(content)

	templates, err := assets.AssetDir(templatesDir)
	if err != nil {
		return "", err
	}

	for _, template := range templates {
		t, err := assets.Asset(filepath.Join(templatesDir, template))
		if err != nil {
			return "", err
		}
		license := string(t)
		license = Normalize(license)
		if license == content {
			return strings.TrimSuffix(template, filepath.Ext(template)), nil
		}
	}

	return "", fmt.Errorf("cannot identify license content")
}

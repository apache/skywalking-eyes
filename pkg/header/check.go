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

package header

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/internal/logger"
	lcs "github.com/apache/skywalking-eyes/pkg/license"
)

// Check checks the license headers of the specified paths/globs.
func Check(config *ConfigHeader, result *Result) error {
	fileList, err := listFiles(config)
	if err != nil {
		return err
	}

	for _, file := range fileList {
		if err := CheckFile(file, config, result); err != nil {
			return nil
		}
	}

	return nil
}

func listFiles(config *ConfigHeader) ([]string, error) {
	var fileList []string

	repo, err := git.PlainOpen("./")

	if err != nil {
		return nil, err
	} else {
		head, _ := repo.Head()
		commit, _ := repo.CommitObject(head.Hash())
		tree, err := commit.Tree()
		if err != nil {
			return nil, err
		}
		err = tree.Files().ForEach(func(file *object.File) error {
			if file != nil {
				fileList = append(fileList, file.Name)
			}
			return errors.New("file pointer is nil")
		})
		if err != nil {
			return nil, err
		}
	}

	return fileList, nil
}

var seen = make(map[string]bool)

// CheckFile checks whether the file contains the configured license header.
func CheckFile(file string, config *ConfigHeader, result *Result) error {
	if yes, err := config.ShouldIgnore(file); yes || err != nil {
		if !seen[file] {
			result.Ignore(file)
		}
		return err
	}

	logger.Log.Debugln("Checking file:", file)

	bs, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if t := http.DetectContentType(bs); !strings.HasPrefix(t, "text/") {
		logger.Log.Debugln("Ignoring file:", file, "; type:", t)
		return nil
	}

	content := lcs.NormalizeHeader(string(bs))
	expected, pattern := config.NormalizedLicense(), config.NormalizedPattern()

	if satisfy(content, config, expected, pattern) {
		result.Succeed(file)
	} else {
		logger.Log.Debugln("Content is:", content)

		result.Fail(file)
	}

	return nil
}

func satisfy(content string, config *ConfigHeader, license string, pattern *regexp.Regexp) bool {
	if index := strings.Index(content, license); strings.TrimSpace(license) != "" && index >= 0 {
		return index < config.LicenseLocationThreshold
	}

	if pattern == nil {
		return false
	}
	index := pattern.FindStringIndex(content)

	return len(index) == 2 && index[0] < config.LicenseLocationThreshold
}

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
//
package header

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/internal/logger"
	lcs "github.com/apache/skywalking-eyes/pkg/license"

	"github.com/bmatcuk/doublestar/v2"
)

// Check checks the license headers of the specified paths/globs.
func Check(config *ConfigHeader, result *Result) error {
	for _, pattern := range config.Paths {
		if err := checkPattern(pattern, result, config); err != nil {
			return err
		}
	}

	return nil
}

var seen = make(map[string]bool)

func checkPattern(pattern string, result *Result, config *ConfigHeader) error {
	paths, err := glob(pattern)

	if err != nil {
		return err
	}

	for _, path := range paths {
		if yes, err := config.ShouldIgnore(path); yes || err != nil {
			result.Ignore(path)
			continue
		}
		if err := checkPath(path, result, config); err != nil {
			return err
		}
		seen[path] = true
	}

	return nil
}

func glob(pattern string) (matches []string, err error) {
	if pattern == "." {
		return doublestar.Glob("./")
	}

	return doublestar.Glob(pattern)
}

func checkPath(path string, result *Result, config *ConfigHeader) error {
	defer func() { seen[path] = true }()

	if seen[path] {
		return nil
	}

	if yes, err := config.ShouldIgnore(path); yes || err != nil {
		return err
	}

	pathInfo, err := os.Stat(path)

	if err != nil {
		return err
	}

	switch mode := pathInfo.Mode(); {
	case mode.IsDir():
		if err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if p == path { // when p is symbolic link file, it causes infinite recursive calls
				return nil
			}
			return checkPath(p, result, config)
		}); err != nil {
			return err
		}
	case mode.IsRegular():
		return CheckFile(path, config, result)
	}
	return nil
}

// CheckFile checks whether or not the file contains the configured license header.
func CheckFile(file string, config *ConfigHeader, result *Result) error {
	if yes, err := config.ShouldIgnore(file); yes || err != nil {
		if !seen[file] {
			result.Ignore(file)
		}
		return err
	}

	logger.Log.Debugln("Checking file:", file)

	bs, err := ioutil.ReadFile(file)
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

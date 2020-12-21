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
	"bufio"
	"github.com/bmatcuk/doublestar/v2"
	"license-checker/internal/logger"
	"os"
	"path/filepath"
	"strings"
)

const CommentChars = "/*#- !~"

// Check checks the license headers of the specified paths/globs.
func Check(config *Config, result *Result) error {
	for _, pattern := range config.Paths {
		if err := checkPattern(pattern, result, config); err != nil {
			return err
		}
	}

	return nil
}

func checkPattern(pattern string, result *Result, config *Config) error {
	paths, err := doublestar.Glob(pattern)

	if err != nil {
		return err
	}

	for _, path := range paths {
		if yes, err := config.ShouldIgnore(path); yes || err != nil {
			continue
		}
		if err = checkPath(path, result, config); err != nil {
			return err
		}
	}

	return nil
}

var seen = make(map[string]bool)

func checkPath(path string, result *Result, config *Config) error {
	defer func() { seen[path] = true }()

	if yes, err := config.ShouldIgnore(path); yes || seen[path] || err != nil {
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
			if err := checkPath(p, result, config); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	case mode.IsRegular():
		return CheckFile(path, config, result)
	}
	return nil
}

// CheckFile checks whether or not the file contains the configured license header.
func CheckFile(file string, config *Config, result *Result) error {
	if yes, err := config.ShouldIgnore(file); yes || err != nil {
		return err
	}

	logger.Log.Debugln("Checking file:", file)

	reader, err := os.Open(file)

	if err != nil {
		return nil
	}

	var lines []string

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), CommentChars)
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	if content := strings.Join(lines, " "); !strings.Contains(content, config.NormalizedLicense()) {
		logger.Log.Debugln("Content is:", content)
		logger.Log.Debugln("License is:", config.License)

		result.Fail(file)
	} else {
		result.Succeed(file)
	}

	return nil
}

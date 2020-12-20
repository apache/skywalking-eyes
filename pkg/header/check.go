// Copyright © 2020 Hoshea Jiang <hoshea@apache.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package header

import (
	"bufio"
	"fmt"
	"github.com/bmatcuk/doublestar/v2"
	"license-checker/internal/logger"
	"os"
	"path/filepath"
	"strings"
)

const CommentChars = "/*#- "

// Check checks the license headers of the specified paths/globs.
func Check(config *Config) error {
	var result Result

	for _, pattern := range config.Paths {
		if err := checkPattern(pattern, &result, config); err != nil {
			return err
		}
	}

	if len(result.Failure) > 0 {
		return fmt.Errorf(
			"The following files don't have a valid license header: \n%v",
			strings.Join(result.Failure, "\n"),
		)
	}

	return nil
}

func checkPattern(pattern string, result *Result, config *Config) error {
	paths, err := doublestar.Glob(pattern)

	if err != nil {
		return err
	}

	logger.Log.Infoln("Checking matched paths:", paths)

	for _, path := range paths {
		logger.Log.Debugln("Checking path:", path)

		if err = checkPath(path, result, config); err != nil {
			return err
		}
	}

	return nil
}

func checkPath(path string, result *Result, config *Config) error {
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
		return checkFile(path, result, config)
	}
	return nil
}

func checkFile(file string, result *Result, config *Config) error {
	skip := false
	for _, ignorePattern := range config.PathsIgnore {
		logger.Log.Debugln("Checking ignore pattern:", ignorePattern)

		if ignored, err := doublestar.Match(ignorePattern, file); ignored || err != nil {
			logger.Log.Infoln("Ignoring path:", file)
			skip = ignored
			break
		}
	}
	if skip {
		return nil
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

	if content := strings.Join(lines, " "); !strings.Contains(content, config.License) {
		logger.Log.Debugln("Content is:", content)
		logger.Log.Debugln("License is:", config.License)

		result.Failure = append(result.Failure, file)
	} else {
		result.Success = append(result.Success, file)
	}

	return nil
}

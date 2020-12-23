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
	"os"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg/license"

	"github.com/bmatcuk/doublestar/v2"
)

type ConfigHeader struct {
	License     string   `yaml:"license"`
	Pattern     string   `yaml:"pattern"`
	Paths       []string `yaml:"paths"`
	PathsIgnore []string `yaml:"paths-ignore"`
}

// NormalizedLicense returns the normalized string of the license content,
// "normalized" means the linebreaks and Punctuations are all trimmed.
func (config *ConfigHeader) NormalizedLicense() string {
	return license.Normalize(config.License)
}

func (config *ConfigHeader) NormalizedPattern() *regexp.Regexp {
	pattern := config.Pattern

	if pattern == "" || strings.TrimSpace(pattern) == "" {
		return nil
	}

	pattern = license.NormalizePattern(pattern)

	return regexp.MustCompile("(?i).*" + pattern + ".*")
}

func (config *ConfigHeader) ShouldIgnore(path string) (bool, error) {
	for _, ignorePattern := range config.PathsIgnore {
		if matched, err := doublestar.Match(ignorePattern, path); matched || err != nil {
			return matched, err
		}
	}

	if stat, err := os.Stat(path); err == nil {
		for _, ignorePattern := range config.PathsIgnore {
			ignorePattern = strings.TrimRight(ignorePattern, "/")
			if strings.HasPrefix(path, ignorePattern+"/") || stat.Name() == ignorePattern {
				return true, nil
			}
		}
	}

	return false, nil
}

func (config *ConfigHeader) Finalize() error {
	if len(config.Paths) == 0 {
		config.Paths = []string{"**"}
	}

	config.PathsIgnore = append(config.PathsIgnore, ".git")

	if file, err := os.Open(".gitignore"); err == nil {
		defer func() { _ = file.Close() }()

		for scanner := bufio.NewScanner(file); scanner.Scan(); {
			line := scanner.Text()
			if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
				continue
			}
			logger.Log.Debugln("Add ignore path from .gitignore:", line)
			config.PathsIgnore = append(config.PathsIgnore, strings.TrimSpace(line))
		}
	}

	logger.Log.Debugln("License header is:", config.NormalizedLicense())
	if p := config.NormalizedPattern(); p != nil {
		logger.Log.Debugln("Pattern is:", p)
	}

	return nil
}

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
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"

	"github.com/bmatcuk/doublestar/v2"
	"gopkg.in/yaml.v3"
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
	var lines []string
	for _, line := range strings.Split(config.License, "\n") {
		if len(line) > 0 {
			lines = append(lines, Punctuations.ReplaceAllString(line, " "))
		}
	}
	return strings.ToLower(regexp.MustCompile("(?m)[\\s\"']+").ReplaceAllString(strings.Join(lines, " "), " "))
}

func (config *ConfigHeader) NormalizedPattern() *regexp.Regexp {
	if config.Pattern == "" || strings.TrimSpace(config.Pattern) == "" {
		return nil
	}

	var lines []string
	for _, line := range strings.Split(config.Pattern, "\n") {
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	content := regexp.MustCompile("(?m)[\\s\"':;/\\-]+").ReplaceAllString(strings.Join(lines, " "), " ")
	return regexp.MustCompile("(?i).*" + content + ".*")
}

// Parse reads and parses the header check configurations in config file.
func (config *ConfigHeader) Parse(file string) error {
	logger.Log.Infoln("Loading configuration from file:", file)

	if bytes, err := ioutil.ReadFile(file); err != nil {
		return err
	} else if err := yaml.Unmarshal(bytes, config); err != nil {
		return err
	}

	logger.Log.Debugln("License header is:", config.NormalizedLicense())

	if len(config.Paths) == 0 {
		config.Paths = []string{"**"}
	}

	return nil
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
	logger.Log.Debugln("License header is:", config.NormalizedLicense())

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

	return nil
}

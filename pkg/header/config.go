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
	"github.com/bmatcuk/doublestar/v2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"license-checker/internal/logger"
	"strings"
)

type Config struct {
	License     string   `yaml:"license"`
	Paths       []string `yaml:"paths"`
	PathsIgnore []string `yaml:"paths-ignore"`
}

// NormalizedLicense returns the normalized string of the license content,
// "normalized" means the linebreaks and CommentChars are all trimmed.
func (config *Config) NormalizedLicense() string {
	var lines []string
	for _, line := range strings.Split(config.License, "\n") {
		if len(line) > 0 {
			lines = append(lines, strings.Trim(line, CommentChars))
		}
	}
	return strings.Join(lines, " ")
}

// Parse reads and parses the header check configurations in config file.
func (config *Config) Parse(file string) error {
	logger.Log.Infoln("Loading configuration from file:", file)

	if bytes, err := ioutil.ReadFile(file); err != nil {
		return err
	} else if err = yaml.Unmarshal(bytes, config); err != nil {
		return err
	}

	logger.Log.Debugln("License header is:", config.NormalizedLicense())

	if len(config.Paths) == 0 {
		config.Paths = []string{"**"}
	}

	return nil
}

func (config *Config) ShouldIgnore(path string) (bool, error) {
	for _, ignorePattern := range config.PathsIgnore {
		if matched, err := doublestar.Match(ignorePattern, path); matched || err != nil {
			return matched, err
		}
	}
	return false, nil
}

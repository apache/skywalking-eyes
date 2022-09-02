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
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/comments"
	"github.com/apache/skywalking-eyes/pkg/license"

	"github.com/bmatcuk/doublestar/v2"
)

type CommentOption string

var (
	Always    CommentOption = "always"
	Never     CommentOption = "never"
	OnFailure CommentOption = "on-failure"

	ASFNames = regexp.MustCompile("(?i)(the )?(Apache Software Foundation|ASF)")
)

type LicenseConfig struct {
	SpdxID         string `yaml:"spdx-id"`
	CopyrightOwner string `yaml:"copyright-owner"`
	SoftwareName   string `yaml:"software-name"`
	Content        string `yaml:"content"`
	Pattern        string `yaml:"pattern"`
}

type ConfigHeader struct {
	License     LicenseConfig `yaml:"license"`
	Paths       []string      `yaml:"paths"`
	PathsIgnore []string      `yaml:"paths-ignore"`
	Comment     CommentOption `yaml:"comment"`

	// LicenseLocationThreshold specifies the index threshold where the license header can be located,
	// after all, a "header" cannot be TOO far from the file start.
	LicenseLocationThreshold int                          `yaml:"license-location-threshold"`
	Languages                map[string]comments.Language `yaml:"language"`
}

// NormalizedLicense returns the normalized string of the license content,
// "normalized" means the linebreaks and Punctuations are all trimmed.
func (config *ConfigHeader) NormalizedLicense() string {
	return license.Normalize(config.GetLicenseContent())
}

func (config *ConfigHeader) LicensePattern(style *comments.CommentStyle) *regexp.Regexp {
	pattern := config.License.Pattern

	if pattern == "" || strings.TrimSpace(pattern) == "" {
		return nil
	}

	// Trim leading and trailing newlines
	pattern = strings.TrimSpace(pattern)
	lines := strings.Split(pattern, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = fmt.Sprintf("%v %v", style.Middle, line)
		} else {
			lines[i] = style.Middle
		}
	}

	lines = append(lines, "(("+style.Middle+"\n)*|\n*)")

	if style.Start != style.Middle {
		lines = append([]string{style.Start}, lines...)
	}

	if style.End != style.Middle {
		lines = append(lines, style.End)
	}

	pattern = strings.Join(lines, "\n")

	return regexp.MustCompile("(?s)" + pattern)
}

func (config *ConfigHeader) NormalizedPattern() *regexp.Regexp {
	pattern := config.License.Pattern

	if pattern == "" || strings.TrimSpace(pattern) == "" {
		return nil
	}

	pattern = license.NormalizePattern(pattern)

	return regexp.MustCompile("(?i).*" + pattern + ".*")
}

func (config *ConfigHeader) ShouldIgnore(path string) (bool, error) {
	var matched bool

	for _, ignorePattern := range config.Paths {
		m, err := doublestar.Match(ignorePattern, path)
		if err != nil {
			return true, err
		}

		if m {
			matched = m
			break
		}
	}

	if !matched {
		return true, nil
	}

	for _, ignorePattern := range config.PathsIgnore {
		if m, err := doublestar.Match(ignorePattern, path); m || err != nil {
			return true, err
		}
	}

	return false, nil
}

func (config *ConfigHeader) Finalize() error {
	if len(config.Paths) == 0 {
		config.Paths = []string{"**"}
	}

	comments.OverrideLanguageCommentStyle(config.Languages)

	config.PathsIgnore = append(config.PathsIgnore, "**/*.txt")

	logger.Log.Debugln("License header is:", config.NormalizedLicense())

	if p := config.NormalizedPattern(); p != nil {
		logger.Log.Debugln("Pattern is:", p)
	}

	if config.LicenseLocationThreshold <= 0 {
		config.LicenseLocationThreshold = 80
	}

	return nil
}

func (config *ConfigHeader) GetLicenseContent() string {
	if c := strings.TrimSpace(config.License.Content); c != "" {
		return config.License.Content // Do not change anything in user config
	}
	c, err := readLicenseFromSpdx(config)
	if err != nil {
		logger.Log.Warnln(err)
		return ""
	}
	return c
}

func readLicenseFromSpdx(config *ConfigHeader) (string, error) {
	spdxID, owner, name := config.License.SpdxID, config.License.CopyrightOwner, config.License.SoftwareName
	filename := fmt.Sprintf("header-templates/%v.txt", spdxID)

	if spdxID == "Apache-2.0" && ASFNames.MatchString(owner) {
		// Note that the Apache Software Foundation uses a different source header that is related to our use of a CLA.
		// Our instructions for our project's source headers are here (https://www.apache.org/legal/src-headers.html#headers).
		filename = "header-templates/Apache-2.0-ASF.txt"
	}

	content, err := assets.Asset(filename)
	if err != nil {
		return "", fmt.Errorf("failed to find a license template for spdx id %v, %w", spdxID, err)
	}
	template := string(content)
	template = strings.ReplaceAll(template, "[year]", strconv.Itoa(time.Now().Year()))
	template = strings.ReplaceAll(template, "[owner]", owner)
	template = strings.ReplaceAll(template, "[software-name]", name)

	return template, nil
}

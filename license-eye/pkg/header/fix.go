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
//
package header

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg"
	"github.com/apache/skywalking-eyes/license-eye/pkg/comments"
)

// Fix adds the configured license header to the given file.
func Fix(file string, config *ConfigHeader, result *pkg.Result) error {
	var r pkg.Result
	if err := CheckFile(file, config, &r); err != nil || !r.HasFailure() {
		logger.Log.Warnln("Try to fix a valid file, do nothing:", file)
		return err
	}

	style := comments.FileCommentStyle(file)

	if style == nil {
		return fmt.Errorf("unsupported file: %v", file)
	}

	if err := InsertComment(file, style, config, result); err != nil {
		return err
	}

	return nil
}

func InsertComment(file string, style *comments.CommentStyle, config *ConfigHeader, result *pkg.Result) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	licenseHeader, err := generateLicenseHeader(style, config)
	if err != nil {
		return err
	}

	content = rewriteContent(style, content, licenseHeader)

	if err := ioutil.WriteFile(file, content, stat.Mode()); err != nil {
		return err
	}

	result.Fix(file)

	return nil
}

func rewriteContent(style *comments.CommentStyle, content []byte, licenseHeader string) []byte {
	if style.After == "" {
		return append([]byte(licenseHeader), content...)
	}

	content = []byte(strings.TrimLeft(string(content), " \n"))
	afterPattern := regexp.MustCompile(style.After)
	location := afterPattern.FindIndex(content)
	if location == nil || len(location) != 2 {
		return append([]byte(licenseHeader), content...)
	}
	return append(content[0:location[1]],
		append(append([]byte("\n"), []byte(licenseHeader)...), content[location[1]+1:]...)...,
	)
}

func generateLicenseHeader(style *comments.CommentStyle, config *ConfigHeader) (string, error) {
	if err := style.Validate(); err != nil {
		return "", err
	}

	middleLines := strings.Split(config.License, "\n")
	for i, line := range middleLines {
		middleLines[i] = strings.TrimRight(fmt.Sprintf("%v %v", style.Middle, line), " ")
	}

	lines := fmt.Sprintf("%v\n%v\n", style.Start, strings.Join(middleLines, "\n"))
	if style.End != style.Middle {
		lines += style.End
	}

	return strings.TrimSpace(lines) + "\n", nil
}

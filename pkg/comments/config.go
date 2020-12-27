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
package comments

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/apache/skywalking-eyes/license-eye/assets"
)

type CommentStyle struct {
	ID     string `yaml:"id"`
	After  string `yaml:"after"`
	Start  string `yaml:"start"`
	Middle string `yaml:"middle"`
	End    string `yaml:"end"`
}

func (style *CommentStyle) Validate() error {
	if style.Start == "" || strings.TrimSpace(style.Start) == "" {
		return fmt.Errorf("comment style 'start' cannot be empty")
	}
	return nil
}

type Language struct {
	Type           string   `yaml:"type"`
	Extensions     []string `yaml:"extensions"`
	Filenames      []string `yaml:"filenames"`
	CommentStyleID string   `yaml:"comment_style_id"`
}

var languages map[string]Language
var comments = make(map[string]CommentStyle)
var commentStyles = make(map[string]CommentStyle)

func init() {
	initLanguages()

	initCommentStyles()

	for _, lang := range languages {
		for _, extension := range lang.Extensions {
			if lang.CommentStyleID == "" {
				continue
			}
			commentStyles[extension] = comments[lang.CommentStyleID]
		}
		for _, filename := range lang.Filenames {
			if lang.CommentStyleID == "" {
				continue
			}
			commentStyles[filename] = comments[lang.CommentStyleID]
		}
	}
}

func initLanguages() {
	content, err := assets.Asset("assets/languages.yaml")
	if err != nil {
		panic(fmt.Errorf("should never happen: %w", err))
	}

	if err := yaml.Unmarshal(content, &languages); err != nil {
		panic(err)
	}

	for s, language := range languages {
		languages[s] = language
	}
}

func initCommentStyles() {
	content, err := assets.Asset("assets/styles.yaml")
	if err != nil {
		panic(fmt.Errorf("should never happen: %w", err))
	}

	var styles []CommentStyle
	if err = yaml.Unmarshal(content, &styles); err != nil {
		panic(err)
	}

	for _, style := range styles {
		comments[style.ID] = style
	}
}

func FileCommentStyle(filename string) *CommentStyle {
	for extension, style := range commentStyles {
		if strings.HasSuffix(filename, extension) {
			return &style
		}
	}
	return nil
}

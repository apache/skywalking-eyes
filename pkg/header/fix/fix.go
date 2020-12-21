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
package fix

import (
	"fmt"
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg/header"
)

var suffixToFunc = map[string]func(string, *header.ConfigHeader, *header.Result) error{
	".go":   DoubleSlash,
	".adoc": DoubleSlash,

	".py":        Hashtag, // TODO: tackle shebang
	".sh":        Hashtag, // TODO: tackle shebang
	".yml":       Hashtag,
	".yaml":      Hashtag,
	".graphql":   Hashtag,
	"Makefile":   Hashtag,
	"Dockerfile": Hashtag,
	".gitignore": Hashtag,

	".md": AngleBracket,

	".java": SlashAsterisk,
}

// Fix adds the configured license header to the given file.
func Fix(file string, config *header.ConfigHeader, result *header.Result) error {
	var r header.Result
	if err := header.CheckFile(file, config, &r); err != nil || !r.HasFailure() {
		logger.Log.Warnln("Try to fix a valid file, returning:", file)
		return err
	}

	for suffix, fixFunc := range suffixToFunc {
		if strings.HasSuffix(file, suffix) {
			return fixFunc(file, config, result)
		}
	}

	return fmt.Errorf("file type is unsupported yet: %v", file)
}

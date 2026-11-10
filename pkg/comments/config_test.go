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

package comments

import "testing"

func TestConfig(t *testing.T) {
	if len(languages) == 0 {
		t.Fail()
	}
}

func TestLanguages(t *testing.T) {
	tests := []struct {
		lang      string
		extension string
	}{
		{lang: "Java", extension: ".java"},
		{lang: "Python", extension: ".py"},
		{lang: "JavaScript", extension: ".js"},
	}
	for _, test := range tests {
		t.Run(test.lang, func(t *testing.T) {
			for _, extension := range languages[test.lang].Extensions {
				if extension == test.extension {
					return
				}
			}
			t.Fail()
		})
	}
}

func TestCommentStyle(t *testing.T) {
	tests := []struct {
		filename       string
		commentStyleID string
	}{
		{filename: "Test.java", commentStyleID: "SlashAsterisk"},
		{filename: "Test.py", commentStyleID: "PythonStyle"},
		{filename: "Test.ps1", commentStyleID: "PowerShellStyle"},
		{filename: "Test.psm1", commentStyleID: "PowerShellStyle"},
		{filename: "Test.psd1", commentStyleID: "PowerShellStyle"},
		// These also end with an extension of another language (".inc",
		// ".frag", ".spec"); the more specific filename must win.
		{filename: "Makefile.inc", commentStyleID: "Hashtag"},
		{filename: "src/Makefile.inc", commentStyleID: "Hashtag"},
		{filename: "Makefile.frag", commentStyleID: "Hashtag"},
		{filename: "buildozer.spec", commentStyleID: "Semicolon"},
	}
	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			// Map iteration order is random, so look up more than once.
			for i := 0; i < 100; i++ {
				if style := FileCommentStyle(test.filename); test.commentStyleID != style.ID {
					t.Fatalf("Extension = %v, Comment style = %v, want %v", test.filename, style.ID, test.commentStyleID)
				}
			}
		})
	}
}

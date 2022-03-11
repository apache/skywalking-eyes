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
	}
	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			if style := FileCommentStyle(test.filename); test.commentStyleID != style.ID {
				t.Logf("Extension = %v, Comment style = %v", test.filename, style.ID)
				t.Fail()
			}
		})
	}
}

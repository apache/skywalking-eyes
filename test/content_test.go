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

package test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apache/skywalking-eyes/pkg/header"
	"gopkg.in/yaml.v3"
)

var c struct {
	Header header.ConfigHeader `yaml:"header"`
}

func init() {
	content, err := ioutil.ReadFile("testdata/test-spdx-content.yaml")

	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(content, &c); err != nil {
		panic(err)
	}
	if err := c.Header.Finalize(); err != nil {
		panic(err)
	}
}

func TestCheckFile(t *testing.T) {
	type args struct {
		name       string
		file       string
		result     *header.Result
		wantErr    bool
		hasFailure bool
	}

	tests := func() []args {
		files, err := filepath.Glob("testdata/spdx_content_test/*")
		if err != nil {
			t.Error(err)
		}

		var cases []args

		for _, file := range files {
			cases = append(cases, args{
				name:       file,
				file:       file,
				result:     &header.Result{},
				wantErr:    false,
				hasFailure: false,
			})
		}

		return cases
	}()

	if len(tests) == 0 {
		t.Errorf("Tests should not be empty")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.TrimSpace(c.Header.GetLicenseContent()) == "" {
				t.Errorf("License should not be empty")
			}
			if err := header.CheckFile(tt.file, &c.Header, tt.result); (err != nil) != tt.wantErr {
				t.Errorf("CheckFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(tt.result.Ignored) > 0 {
				t.Errorf("Should not ignore any file, %v", tt.result.Ignored)
			}
			if tt.result.HasFailure() != tt.hasFailure {
				t.Errorf("CheckFile() result has failure = %v, wanted = %v", tt.result.Failure, tt.hasFailure)
			}
		})
	}
}

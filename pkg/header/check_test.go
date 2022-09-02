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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCheckFile(t *testing.T) {
	type args struct {
		name   string
		file   string
		result *Result
	}

	var c struct {
		Header ConfigHeader `yaml:"header"`
	}

	content, err := os.ReadFile("../../test/testdata/.licenserc_for_test_check.yaml")
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(content, &c))
	require.NoError(t, c.Header.Finalize())

	t.Run("WithLicense", func(t *testing.T) {
		tests := func() []args {
			files, err := filepath.Glob("../../test/testdata/include_test/with_license/*")
			require.NoError(t, err)
			var cases []args
			for _, file := range files {
				cases = append(cases, args{
					name:   file,
					file:   file,
					result: &Result{},
				})
			}
			return cases
		}()
		require.NotEmpty(t, tests)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.NotEmpty(t, strings.TrimSpace(c.Header.GetLicenseContent()))
				require.NoError(t, CheckFile(tt.file, &c.Header, tt.result))
				require.Len(t, tt.result.Ignored, 0)
				require.False(t, tt.result.HasFailure())
			})
		}
	})

	t.Run("WithoutLicense", func(t *testing.T) {
		tests := func() []args {
			files, err := filepath.Glob("../../test/testdata/include_test/without_license/*")
			require.NoError(t, err)
			var cases []args
			for _, file := range files {
				cases = append(cases, args{
					name:   file,
					file:   file,
					result: &Result{},
				})
			}
			return cases
		}()
		require.NotEmpty(t, tests)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.NotEmpty(t, strings.TrimSpace(c.Header.GetLicenseContent()))
				require.NoError(t, CheckFile(tt.file, &c.Header, tt.result))
				require.Len(t, tt.result.Ignored, 0)
				require.True(t, tt.result.HasFailure())
			})
		}
	})
}

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
	"testing"

	"github.com/stretchr/testify/require"
)

var diffConfig = &ConfigHeader{
	License: LicenseConfig{
		Content: `Apache License 2.0
  http://www.apache.org/licenses/LICENSE-2.0
Apache License 2.0`,
	},
	LicenseLocationThreshold: 80,
}

func TestDiffFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		config   *ConfigHeader
		expected string
	}{
		{
			name:     "valid header",
			filename: "valid.go",
			content: `// Apache License 2.0
//   http://www.apache.org/licenses/LICENSE-2.0
// Apache License 2.0

package main
`,
			config:   diffConfig,
			expected: "",
		},
		{
			name:     "typo in the header",
			filename: "typo.go",
			content: `// Apache License 2.0
//   wwwhttp://www.apache.org/licenses/LICENSE-2.0
// Apache License 2.0

package main
`,
			config: diffConfig,
			expected: "apache license 2.0 " +
				"[-http://www.apache.org/licenses/license-2.0-] " +
				"{+wwwhttp://www.apache.org/licenses/license-2.0+} " +
				"apache license 2.0 ...",
		},
		{
			name:     "no header at all",
			filename: "missing.go",
			content: `package main

func main() {}
`,
			config: diffConfig,
			expected: "[-apache license 2.0 http://www.apache.org/licenses/license-2.0 apache license 2.0-] " +
				"...",
		},
		{
			name:     "header too far from the file start",
			filename: "far.go",
			content: `// aaaa bbbb cccc dddd eeee ffff gggg hhhh
// Apache License 2.0
//   http://www.apache.org/licenses/LICENSE-2.0
// Apache License 2.0

package main
`,
			config: &ConfigHeader{
				License:                  diffConfig.License,
				LicenseLocationThreshold: 10,
			},
			expected: "license header is found at normalized offset 40, " +
				"which exceeds the license-location-threshold 10, " +
				"move it closer to the file start",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), test.filename)
			require.NoError(t, os.WriteFile(file, []byte(test.content), 0o600))

			diff, err := DiffFile(file, test.config)
			require.NoError(t, err)
			require.Equal(t, test.expected, diff)
		})
	}
}

func TestDiffFileWithoutLicenseContent(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.go")
	require.NoError(t, os.WriteFile(file, []byte("package main\n"), 0o600))

	_, err := DiffFile(file, &ConfigHeader{LicenseLocationThreshold: 80})
	require.Error(t, err)
}

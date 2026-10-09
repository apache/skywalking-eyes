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
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/require"
)

// encodeUTF16 encodes s as UTF-16 with the given byte order, prefixed with bom.
// It is independent of the implementation under test, which uses x/text.
func encodeUTF16(s string, order binary.ByteOrder, bom []byte) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, 0, len(bom)+len(units)*2)
	out = append(out, bom...)
	var buf [2]byte
	for _, unit := range units {
		order.PutUint16(buf[:], unit)
		out = append(out, buf[0], buf[1])
	}
	return out
}

func decodeUTF16(t *testing.T, content []byte, order binary.ByteOrder) string {
	t.Helper()
	require.Zero(t, len(content)%2, "UTF-16 content must have an even number of bytes")
	units := make([]uint16, len(content)/2)
	for i := range units {
		units[i] = order.Uint16(content[i*2:])
	}
	return string(utf16.Decode(units))
}

// TestFixPreservesFileEncoding ensures that fixing a BOM encoded file keeps
// its encoding and byte-order mark, see
// https://github.com/apache/skywalking-eyes/pull/285#issuecomment-6081296040.
func TestFixPreservesFileEncoding(t *testing.T) {
	const source = "#!/usr/bin/env pwsh\nWrite-Host 'Hello World'\n"
	tests := []struct {
		name  string
		bom   []byte
		order binary.ByteOrder
	}{
		{name: "UTF-16LE with BOM", bom: []byte{0xFF, 0xFE}, order: binary.LittleEndian},
		{name: "UTF-16BE with BOM", bom: []byte{0xFE, 0xFF}, order: binary.BigEndian},
		{name: "UTF-8 with BOM", bom: []byte{0xEF, 0xBB, 0xBF}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := append(append([]byte{}, test.bom...), source...)
			if test.order != nil {
				content = encodeUTF16(source, test.order, test.bom)
			}
			file := filepath.Join(t.TempDir(), "test.ps1")
			require.NoError(t, os.WriteFile(file, content, 0o600))

			config := &ConfigHeader{
				License: LicenseConfig{Content: "Apache License 2.0"},
				Paths:   []string{"**"},
			}
			require.NoError(t, config.Finalize())

			var before Result
			require.NoError(t, CheckFile(file, config, &before))
			require.True(t, before.HasFailure(), "a headerless file must not pass the check")

			var fixed Result
			require.NoError(t, Fix(file, config, &fixed))
			require.Equal(t, []string{file}, fixed.Fixed)

			after, err := os.ReadFile(file)
			require.NoError(t, err)
			require.True(t, bytes.HasPrefix(after, test.bom), "the BOM must stay at byte 0")

			text := string(after[len(test.bom):])
			if test.order != nil {
				text = decodeUTF16(t, after[len(test.bom):], test.order)
			}
			require.Contains(t, text, "#!/usr/bin/env pwsh\n")
			require.Contains(t, text, "<#\n Apache License 2.0\n#>\n")
			require.Contains(t, text, "Write-Host 'Hello World'")

			var checked Result
			require.NoError(t, CheckFile(file, config, &checked))
			require.False(t, checked.HasFailure(), "the fixed file must pass the check")
			require.Equal(t, []string{file}, checked.Success)

			// Fixing an already valid file must not change it.
			var again Result
			require.NoError(t, Fix(file, config, &again))
			unchanged, err := os.ReadFile(file)
			require.NoError(t, err)
			require.Equal(t, after, unchanged)
		})
	}
}

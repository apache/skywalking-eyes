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
	"fmt"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// fileEncoding identifies the byte encoding of a file so that header fixes can
// preserve it, including its byte-order mark (BOM).
type fileEncoding int

const (
	encodingRaw fileEncoding = iota // UTF-8 without a BOM
	encodingUTF8BOM
	encodingUTF16LE
	encodingUTF16BE
)

var (
	utf8BOM    = []byte{0xEF, 0xBB, 0xBF}
	utf16LEBOM = []byte{0xFF, 0xFE}
	utf16BEBOM = []byte{0xFE, 0xFF}
	utf32LEBOM = []byte{0xFF, 0xFE, 0x00, 0x00}
	utf32BEBOM = []byte{0x00, 0x00, 0xFE, 0xFF}
)

// decodeContent converts the file content into UTF-8 and reports the encoding
// it was written in. Files without a BOM are returned unchanged, because their
// encoding cannot be determined reliably (e.g. UTF-16 without a BOM).
func decodeContent(content []byte) ([]byte, fileEncoding, error) {
	switch {
	case bytes.HasPrefix(content, utf32LEBOM), bytes.HasPrefix(content, utf32BEBOM):
		// The UTF-32LE BOM shares the UTF-16LE prefix, so check it first.
		return nil, encodingRaw, fmt.Errorf("unsupported encoding: UTF-32")
	case bytes.HasPrefix(content, utf16LEBOM):
		decoded, err := decodeWith(content[len(utf16LEBOM):], unicode.LittleEndian)
		return decoded, encodingUTF16LE, err
	case bytes.HasPrefix(content, utf16BEBOM):
		decoded, err := decodeWith(content[len(utf16BEBOM):], unicode.BigEndian)
		return decoded, encodingUTF16BE, err
	case bytes.HasPrefix(content, utf8BOM):
		return content[len(utf8BOM):], encodingUTF8BOM, nil
	default:
		return content, encodingRaw, nil
	}
}

// decodeContentLossless is decodeContent for callers that rewrite the file. The
// UTF-16 decoder replaces malformed input, such as an odd trailing byte or an
// unpaired surrogate, with U+FFFD instead of failing, so content that does not
// survive a round trip is rejected to keep the original bytes intact.
func decodeContentLossless(content []byte) ([]byte, fileEncoding, error) {
	decoded, encoding, err := decodeContent(content)
	if err != nil {
		return nil, encoding, err
	}
	encoded, err := encodeContent(decoded, encoding)
	if err != nil {
		return nil, encoding, err
	}
	// Only UTF-16 decoding is lossy, the other encodings always round trip.
	if !bytes.Equal(encoded, content) {
		return nil, encoding, fmt.Errorf("malformed UTF-16 content")
	}
	return decoded, encoding, nil
}

func decodeWith(content []byte, order unicode.Endianness) ([]byte, error) {
	decoded, _, err := transform.Bytes(unicode.UTF16(order, unicode.IgnoreBOM).NewDecoder(), content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode UTF-16 content: %w", err)
	}
	return decoded, nil
}

// encodeContent converts UTF-8 content back into the given encoding, restoring
// the BOM at byte 0.
func encodeContent(content []byte, encoding fileEncoding) ([]byte, error) {
	switch encoding {
	case encodingUTF16LE:
		encoded, err := encodeWith(content, unicode.LittleEndian)
		if err != nil {
			return nil, err
		}
		return append(append([]byte{}, utf16LEBOM...), encoded...), nil
	case encodingUTF16BE:
		encoded, err := encodeWith(content, unicode.BigEndian)
		if err != nil {
			return nil, err
		}
		return append(append([]byte{}, utf16BEBOM...), encoded...), nil
	case encodingUTF8BOM:
		return append(append([]byte{}, utf8BOM...), content...), nil
	default:
		return content, nil
	}
}

func encodeWith(content []byte, order unicode.Endianness) ([]byte, error) {
	encoded, _, err := transform.Bytes(unicode.UTF16(order, unicode.IgnoreBOM).NewEncoder(), content)
	if err != nil {
		return nil, fmt.Errorf("failed to encode UTF-16 content: %w", err)
	}
	return encoded, nil
}

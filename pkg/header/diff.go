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
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	lcs "github.com/apache/skywalking-eyes/pkg/license"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// DiffFile compares the license header of the file with the license configured
// in the config file, and returns a word-level diff of the two normalized texts,
// the same texts that CheckFile compares, so every difference in the diff is a
// real cause of the check failure.
//
// In the diff, [-text-] marks text that is expected by the configured license
// but missing in the file, and {+text+} marks text that is in the file but not
// expected by the configured license. An empty diff is returned when the file's
// license header is valid.
func DiffFile(file string, config *ConfigHeader) (string, error) {
	expected := config.NormalizedLicense()
	if expected == "" {
		return "", fmt.Errorf("no license content configured (spdx-id or content) to diff against")
	}

	bs, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	if t := http.DetectContentType(bs); !strings.HasPrefix(t, "text/") {
		return "", fmt.Errorf("not a text file: %v (%v)", file, t)
	}

	content := lcs.NormalizeHeader(string(bs))
	if satisfy(content, config, expected, config.NormalizedPattern()) {
		return "", nil
	}

	if index := strings.Index(content, expected); index >= 0 {
		return fmt.Sprintf(
			"license header is found at normalized offset %d, which exceeds the license-location-threshold %d, move it closer to the file start",
			index, config.LicenseLocationThreshold,
		), nil
	}

	// Only diff the region of the file where the license header is allowed to live,
	// the content after that region cannot contribute to a successful match anyway.
	end := len(expected) + config.LicenseLocationThreshold
	if end >= len(content) {
		end = len(content)
	} else {
		for end > 0 && !utf8.RuneStart(content[end]) {
			end--
		}
	}

	return renderDiff(wordDiff(expected, content[:end])), nil
}

// wordDiff diffs the two texts word by word, by mapping every word to a "line"
// and reusing the line-mode diff of diffmatchpatch.
func wordDiff(expected, actual string) []diffmatchpatch.Diff {
	dmp := diffmatchpatch.New()

	// The trailing "\n" makes the last word a complete "line" too, so that it can
	// match its occurrences in the middle of the other text.
	e, a, words := dmp.DiffLinesToChars(
		strings.ReplaceAll(expected, " ", "\n")+"\n",
		strings.ReplaceAll(actual, " ", "\n")+"\n",
	)
	diffs := dmp.DiffMain(e, a, false)

	return dmp.DiffCharsToLines(diffs, words)
}

func renderDiff(diffs []diffmatchpatch.Diff) string {
	const (
		// contextWords is the number of words to keep on each side when a long run of words is collapsed.
		contextWords = 5
		// maxEqualRun is the maximum number of words an unchanged run can have before being collapsed.
		maxEqualRun = 12
		// maxChangedRun is the maximum number of words a changed run can have before being collapsed,
		// it's larger than maxEqualRun because the changed words are what the user wants to see.
		maxChangedRun = 40
	)

	segments := make([]string, 0, len(diffs))
	for i, diff := range diffs {
		words := strings.Fields(diff.Text)
		if len(words) == 0 {
			continue
		}
		last := i == len(diffs)-1
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			segments = append(segments, collapseWords(words, maxEqualRun, contextWords, i == 0, last))
		case diffmatchpatch.DiffDelete:
			segments = append(segments, "[-"+collapseWords(words, maxChangedRun, contextWords*2, false, false)+"-]")
		case diffmatchpatch.DiffInsert:
			if last {
				// The trailing inserted words are the file contents after the license header
				// region, which are irrelevant to the diff.
				segments = append(segments, "...")
				continue
			}
			segments = append(segments, "{+"+collapseWords(words, maxChangedRun, contextWords*2, false, false)+"+}")
		}
	}

	return strings.Join(segments, " ")
}

// collapseWords joins the words with spaces, eliding the middle of runs longer
// than maxRun. dropHead/dropTail elide one entire side instead, for runs at the
// beginning/end of the diff where only the words next to a change matter.
func collapseWords(words []string, maxRun, context int, dropHead, dropTail bool) string {
	if len(words) <= maxRun {
		return strings.Join(words, " ")
	}

	head := strings.Join(words[:context], " ")
	tail := strings.Join(words[len(words)-context:], " ")
	switch {
	case dropHead:
		return "... " + tail
	case dropTail:
		return head + " ..."
	default:
		return head + " ... " + tail
	}
}

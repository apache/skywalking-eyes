//
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
package license

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
)

type Normalizer func(string) string

var (
	// normalizers is a list of Normalizer that can be applied to the license text, yet doesn't change the license's
	// meanings, according to the matching guide in https://spdx.dev/license-list/matching-guidelines.
	// The order matters.
	normalizers = []Normalizer{
		VariablesNormalizer,
		OneLineNormalizer,
		FlattenSpaceNormalizer,
		SubstantiveTextsNormalizer,
		FlattenSpaceNormalizer,
		strings.ToLower,
		strings.TrimSpace,
	}

	// 6. Code Comment Indicators (https://spdx.dev/license-list/matching-guidelines.)
	commentIndicators = []*regexp.Regexp{
		regexp.MustCompile(`(?m)^\s*#+`),    // #
		regexp.MustCompile(`(?m)^\s*//+`),   // //
		regexp.MustCompile(`(?m)^\s*"""+`),  // """
		regexp.MustCompile(`(?m)^\s*\(\*+`), // (*
		regexp.MustCompile(`(?m)^\s*;+`),    // ;

		regexp.MustCompile(`(?m)^\s*/\*+`), // /*
		regexp.MustCompile(`(?m)^\s*\*+/`), //  */
		regexp.MustCompile(`(?m)^\s*\*+`),  //  *

		regexp.MustCompile(`(?m)^\s*<!--+`), // <!--
		regexp.MustCompile(`(?m)^\s*--+>`),  // -->
		regexp.MustCompile(`(?m)^\s*--+`),   // --
		regexp.MustCompile(`(?m)^\s*~+`),    //   ~

		regexp.MustCompile(`(?m)^\s*{-+`), // {-
		regexp.MustCompile(`(?m)^\s*-}+`), // -}

		regexp.MustCompile(`(?m)^\s*::`),     // ::
		regexp.MustCompile(`(?m)^\s*\.\.`),   // ..
		regexp.MustCompile(`(?mi)^\s*@?REM`), // @REM
		regexp.MustCompile(`(?mi)^\s*%+`),    // % e.g. matlab
	}

	flattenSpace = regexp.MustCompile(`\s+`)

	substitutableTexts = []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`(?i)\backnowledgement\b`), "acknowledgment"},
		{regexp.MustCompile(`(?i)\banalog\b`), "analogue"},
		{regexp.MustCompile(`(?i)\banalyze\b`), "analyse"},
		{regexp.MustCompile(`(?i)\bartifact\b`), "artefact"},
		{regexp.MustCompile(`(?i)\bauthorization\b`), "authorisation"},
		{regexp.MustCompile(`(?i)\bauthorized\b`), "authorised"},
		{regexp.MustCompile(`(?i)\bcaliber\b`), "calibre"},
		{regexp.MustCompile(`(?i)\bcanceled\b`), "cancelled"},
		{regexp.MustCompile(`(?i)\bcapitalizations\b`), "capitalisations"},
		{regexp.MustCompile(`(?i)\bcatalog\b`), "catalogue"},
		{regexp.MustCompile(`(?i)\bcategorize\b`), "categorise"},
		{regexp.MustCompile(`(?i)\bcenter\b`), "centre"},
		{regexp.MustCompile(`(?i)\bcopyright holder\b`), "copyright owner"},
		{regexp.MustCompile(`(?i)\bemphasized\b`), "emphasised"},
		{regexp.MustCompile(`(?i)\bfavor\b`), "favour"},
		{regexp.MustCompile(`(?i)\bfavorite\b`), "favourite"},
		{regexp.MustCompile(`(?i)\bfulfill\b`), "fulfil"},
		{regexp.MustCompile(`(?i)\bfulfillment\b`), "fulfilment"},
		{regexp.MustCompile(`(?i)\binitialize\b`), "initialise"},
		{regexp.MustCompile(`(?i)\bjudgement\b`), "judgment"},
		{regexp.MustCompile(`(?i)\blabeling\b`), "labelling"},
		{regexp.MustCompile(`(?i)\blabor\b`), "labour"},
		{regexp.MustCompile(`(?i)\blicense\b`), "licence"},
		{regexp.MustCompile(`(?i)\bmaximize\b`), "maximise"},
		{regexp.MustCompile(`(?i)\bmodeled\b`), "modelled"},
		{regexp.MustCompile(`(?i)\bmodeling\b`), "modelling"},
		{regexp.MustCompile(`(?i)\bnoncommercial\b`), "non-commercial"},
		{regexp.MustCompile(`(?i)\boffense\b`), "offence"},
		{regexp.MustCompile(`(?i)\boptimize\b`), "optimise"},
		{regexp.MustCompile(`(?i)\borganization\b`), "organisation"},
		{regexp.MustCompile(`(?i)\borganize\b`), "organise"},
		{regexp.MustCompile(`(?i)\bpercent\b`), "per cent"},
		{regexp.MustCompile(`(?i)\bpractice\b`), "practise"},
		{regexp.MustCompile(`(?i)\bprogram\b`), "programme"},
		{regexp.MustCompile(`(?i)\brealize\b`), "realise"},
		{regexp.MustCompile(`(?i)\brecognize\b`), "recognise"},
		{regexp.MustCompile(`(?i)\bsignaling\b`), "signalling"},
		{regexp.MustCompile(`(?i)\bsublicense\b`), "sub-license"},
		{regexp.MustCompile(`(?i)\bsub-license\b`), "sub license"},
		{regexp.MustCompile(`(?i)\bsublicense\b`), "sub license"},
		{regexp.MustCompile(`(?i)\butilization\b`), "utilisation"},
		{regexp.MustCompile(`(?i)\bwhile\b`), "whilst"},
		{regexp.MustCompile(`(?i)\bwilfull\b`), "wilful"},

		{regexp.MustCompile(`©`), "Copyright "},
		{regexp.MustCompile(`\(c\)`), "Copyright "},
		{regexp.MustCompile(`\bhttps://`), "http://"},

		{regexp.MustCompile(`(?i)\b(the )?Apache Software Foundation( \(ASF\))?`), "the ASF"},
	}

	variables = []struct {
		regexp      *regexp.Regexp
		replacement string
	}{
		// BSD-3-Clause
		{
			regexp.MustCompile(`(?im)(^(\s*Copyright \(c\)) (\d{4}) (.+?) (All rights reserved\.)?$\n?)+`),
			"$2 [year] [owner]. $5",
		},
		{
			regexp.MustCompile(`(?i)(neither the name of) (.+?) (nor the names of)`),
			"$1 the copyright holder $3",
		},
		// MIT
		{ // remove optional header
			regexp.MustCompile(`(?im)^\s*The MIT License \(MIT\)$`),
			"",
		},
		{
			regexp.MustCompile(`(?im)^(\s*Copyright \(c\)) (\d{4}) (.+?)$`),
			"$1 [year] [owner]",
		},
		{
			regexp.MustCompile(`(?im)\(including the next paragraph\)`),
			"",
		},
	}
)

// NormalizePattern applies a chain of Normalizers to the license pattern to make it cleaner for identification.
func NormalizePattern(pattern string) string {
	for _, normalize := range normalizers {
		pattern = normalize(pattern)
	}
	return pattern
}

// NormalizeHeader applies a chain of Normalizers to the file header to make it cleaner for identification.
func NormalizeHeader(header string) string {
	ns := append([]Normalizer{CommentIndicatorNormalizer}, normalizers...)
	for _, normalize := range ns {
		logger.Log.Debugf("After normalized by %+v:", runtime.FuncForPC(reflect.ValueOf(normalize).Pointer()).Name())
		header = normalize(header)
		logger.Log.Debugln(header)
	}
	return header
}

// Normalize applies a chain of Normalizers to the license text to make it cleaner for identification.
func Normalize(license string) string {
	for _, normalize := range normalizers {
		license = normalize(license)
	}
	return license
}

// OneLineNormalizer simply removes all line breaks to flatten the license text into one line.
func OneLineNormalizer(text string) string {
	return regexp.MustCompile("[\n\r]+").ReplaceAllString(text, " ")
}

// SubstantiveTextsNormalizer normalizes the license text by substituting some words that
// doesn't change the meaning of the license.
func SubstantiveTextsNormalizer(text string) string {
	for _, s := range substitutableTexts {
		text = s.regex.ReplaceAllString(text, s.replacement)
	}
	return text
}

// CommentIndicatorNormalizer trims the leading characters of comments, such as /*, <!--, --, (*, etc..
func CommentIndicatorNormalizer(text string) string {
	for _, leadingChars := range commentIndicators {
		text = leadingChars.ReplaceAllString(text, "")
	}
	return text
}

// FlattenSpaceNormalizer flattens continuous spaces into a single space.
func FlattenSpaceNormalizer(text string) string {
	return flattenSpace.ReplaceAllString(text, " ")
}

// VariablesNormalizer replace the variables actual value into variable name.
func VariablesNormalizer(text string) string {
	for _, v := range variables {
		text = v.regexp.ReplaceAllString(text, v.replacement)
	}

	return text
}

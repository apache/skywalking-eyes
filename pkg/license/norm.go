package license

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/apache/skywalking-eyes/internal/logger"
)

type Normalizer func(string) string

var (
	// normalizers is a list of Normalizer that can be applied to the license text, yet doesn't change the license's
	// meanings, according to the matching guide in https://spdx.dev/license-list/matching-guidelines.
	// The order matters.
	normalizers = []Normalizer{
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
		regexp.MustCompile(`(?m)^\s*{#+`),    // {#
		regexp.MustCompile(`(?m)^\s*#+}`),    // #}
		regexp.MustCompile(`(?m)^\s*{\*+`),   // {*
		regexp.MustCompile(`(?m)^\s*\*+}`),   // *}
		regexp.MustCompile(`(?m)^\s*'+`),     // '
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
		{regexp.MustCompile(`(?i)\blicence\b`), "license"},
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
		{regexp.MustCompile(`(?i)\bsub licen[sc]e\b`), "sublicense"},
		{regexp.MustCompile(`(?i)\bsub-licen[sc]e\b`), "sublicense"},
		{regexp.MustCompile(`(?i)\butilization\b`), "utilisation"},
		{regexp.MustCompile(`(?i)\bwhile\b`), "whilst"},
		{regexp.MustCompile(`(?i)\bwilfull\b`), "wilful"},

		{regexp.MustCompile(`©`), "Copyright "},
		{regexp.MustCompile(`\(([cC])\)`), "Copyright "},
		{regexp.MustCompile(`\bhttps://`), "http://"},

		{regexp.MustCompile(`“+`), `'`},
		{regexp.MustCompile(`”+`), `'`},
		{regexp.MustCompile(`’+`), "'"},
		{regexp.MustCompile("`+"), "'"},
		{regexp.MustCompile(`"+`), "'"},
		{regexp.MustCompile(`'+`), "'"},

		{regexp.MustCompile(`(?i)\b(the )?Apache Software Foundation( \(ASF\))?`), "the ASF"},

		// Prettier chars
		{regexp.MustCompile(`[-=*]{3,}`), ""},

		// Mozilla Public License, Version 2.0
		// Mozilla Public License Version 2.0
		{
			regexp.MustCompile(`(?i)Mozilla Public License version 2\.0`),
			"Mozilla Public License, Version 2.0",
		},
		// Mozilla Public License, v. 2.0
		// ...
		{
			regexp.MustCompile(`(?i)Mozilla Public License,? v\. ?2\.0`),
			"Mozilla Public License, v. 2.0",
		},

		{
			regexp.MustCompile(`(?i)IN NO EVENT SHALL (.+?) BE LIABLE`),
			"IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE",
		},
		{
			regexp.MustCompile(`(?i)The names of (its|the) contributors may not be used to endorse`),
			"Neither the name of the copyright holder nor the names of its contributors may be used to endorse",
		},
		{
			regexp.MustCompile(`(?i)The name (.+?) may not be used to endorse`),
			"Neither the name of the copyright holder nor the names of its contributors may be used to endorse",
		},
		{
			regexp.MustCompile(`(?i)(neither the name)( of)? (.+?) (nor the names)( of( its authors and)?)?( its)?`),
			"$1 the copyright holder $4",
		},
		{
			regexp.MustCompile(`(?i)you may not use this (file|library) except`),
			"you may not use this file except",
		},

		{
			regexp.MustCompile(`(?i)THIS SOFTWARE IS PROVIDED BY (.+?)'AS IS'`),
			`THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS 'AS IS'`,
		},

		{
			regexp.MustCompile(`(?im)This license applies to .+? portions of this .+?\. The .+? maintains its own .+? license\.`),
			"",
		},

		{
			regexp.MustCompile(`(?im)\(including the next paragraph\)`),
			"",
		},
	}

	lineProcessors = []struct {
		regexp      *regexp.Regexp
		replacement string
	}{
		// BSD-3-Clause
		// MIT
		{ // remove optional header
			regexp.MustCompile(`(?im)^\s*\(?(The )?MIT License( \((MIT|Expat)\))?\)?\s*$`),
			"",
		},
		// The Three Clause BSD License (http://(http://en.wikipedia.org/wiki/bsd_licenses)
		{ // remove optional header
			regexp.MustCompile(`(?im)^\s*?the (two|three)? clause bsd license (\(http(s)?://(\w|\.|/)+\))?$`),
			"",
		},
		// BSD 3-Clause License
		{ // remove optional header
			regexp.MustCompile(`(?im)^\s*?bsd ([23])-clause license\s*$`),
			"",
		},
		// ISC
		{ // remove optional header
			regexp.MustCompile(`(?im)^\s*(The )?ISC License:?$`),
			"",
		},

		// leading chars such as >, * just for pretty printing
		{
			regexp.MustCompile(`(?m)^[>*]\s+`),
			" ",
		},
		// Listing bullets such as a., b., 1., 2.
		{
			regexp.MustCompile(`(?m)^\s*[a-z0-9]\. `),
			" ",
		},
		// Listing bullets such as (a), (b), (1), (2)
		{
			regexp.MustCompile(`(?m)^\s*\([a-z0-9]\) `),
			" ",
		},
		// trailing chars such as >, * just for pretty printing
		{
			regexp.MustCompile(`(?m)\s+[*]$`),
			" ",
		},
		// Copyright (c) .....
		// © Copyright .....
		{
			regexp.MustCompile(`(?m)^\s*([cC©])?\s*Copyright (\([cC©]\))?.+$`),
			"",
		},
		// Portions Copyright (C) ...
		{
			regexp.MustCompile(`(?m)^\s*Portions Copyright (\([cC©]\))?.+$`),
			"",
		},
		// All rights reserved
		{
			regexp.MustCompile(`(?m)^\s*All rights reserved\.?$`),
			"",
		},
		// ... is distributed under the Simplified BSD License:
		{
			regexp.MustCompile(`(?im)^\s*.+ is distributed under the Simplified BSD License\:?$`),
			"",
		},
		// Please consider promoting this project if you find it useful.
		{
			regexp.MustCompile(`(?im)^\s*Please consider promoting this project if you find it useful\.?$`),
			"",
		},

		// This should be the last one processor
		{
			regexp.MustCompile("[\n\r]+"),
			" ",
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
	ns := append([]Normalizer{CommentIndicatorNormalizer}, normalizers...)
	for _, normalize := range ns {
		license = normalize(license)
	}
	return license
}

// OneLineNormalizer normalizes the text line by line and finally merge them into one line.
func OneLineNormalizer(text string) string {
	for _, s := range lineProcessors {
		text = s.regexp.ReplaceAllString(text, s.replacement)
	}
	return text
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

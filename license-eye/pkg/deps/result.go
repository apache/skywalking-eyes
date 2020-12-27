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
package deps

type SpdxID string

const (
	Unknown string = "Unknown"
)

// Result is a single item that represents a resolved dependency license.
type Result struct {
	Dependency      string
	LicenseFilePath string
	LicenseContent  string
	LicenseSpdxID   []string
}

// Report is a collection of resolved Result.
type Report struct {
	Resolved []*Result
	Skipped  []*Result
}

// Resolve marks the dependency's license is resolved.
func (report *Report) Resolve(result *Result) {
	report.Resolved = append(report.Resolved, result)
}

// Skip marks the dependency's license is skipped for some reasons.
func (report *Report) Skip(result *Result) {
	report.Skipped = append(report.Skipped, result)
}

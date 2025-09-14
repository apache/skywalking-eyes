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

package deps

import "testing"

// Test that when requireFSFFree is enabled, a license must be marked FSF Free/Libre
// in the compatibility matrices to be considered compatible, even if listed as Compatible.
func TestCompareCompatible_FSFFreeRequirement(t *testing.T) {
	// Backup globals and restore after
	savedRequireFSF := requireFSFFree
	savedEntry, hadEntry := matrices["Test-FSF"]
	defer func() {
		requireFSFFree = savedRequireFSF
		if hadEntry {
			matrices["Test-FSF"] = savedEntry
		} else {
			delete(matrices, "Test-FSF")
		}
	}()

	matrix := &CompatibilityMatrix{Compatible: []string{"Test-FSF"}}

	// Case: Not FSF-free -> incompatible when requirement enabled
	requireFSFFree = true
	matrices["Test-FSF"] = CompatibilityMatrix{FSFFree: false}
	if compareCompatible(matrix, "Test-FSF", false) {
		t.Fatalf("expected Test-FSF to be incompatible when not FSF-free but requirement is enabled")
	}

	// Case: FSF-free -> compatible
	matrices["Test-FSF"] = CompatibilityMatrix{FSFFree: true}
	if !compareCompatible(matrix, "Test-FSF", false) {
		t.Fatalf("expected Test-FSF to be compatible when FSF-free and requirement is enabled")
	}
}

// Test that when requireOSIApproved is enabled, a license must be OSI-approved
// in the compatibility matrices to be considered compatible, even if listed as Compatible.
func TestCompareCompatible_OSIRequirement(t *testing.T) {
	// Backup globals and restore after
	savedRequireOSI := requireOSIApproved
	savedEntry, hadEntry := matrices["Test-OSI"]
	defer func() {
		requireOSIApproved = savedRequireOSI
		if hadEntry {
			matrices["Test-OSI"] = savedEntry
		} else {
			delete(matrices, "Test-OSI")
		}
	}()

	matrix := &CompatibilityMatrix{Compatible: []string{"Test-OSI"}}

	// Case: Not OSI-approved -> incompatible when requirement enabled
	requireOSIApproved = true
	matrices["Test-OSI"] = CompatibilityMatrix{OSIApproved: false}
	if compareCompatible(matrix, "Test-OSI", false) {
		t.Fatalf("expected Test-OSI to be incompatible when not OSI-approved but requirement is enabled")
	}

	// Case: OSI-approved -> compatible
	matrices["Test-OSI"] = CompatibilityMatrix{OSIApproved: true}
	if !compareCompatible(matrix, "Test-OSI", false) {
		t.Fatalf("expected Test-OSI to be compatible when OSI-approved and requirement is enabled")
	}
}

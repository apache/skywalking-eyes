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

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// Test that YAML keys map correctly into ConfigDeps fields and that
// those fields are applied to internal requirement flags.
func TestYAMLToConfigDepsAndApply(t *testing.T) {
	// Prepare YAML that matches the documented keys
	data := []byte("threshold: 10\nrequire_fsf_free: true\nrequire_osi_approved: false\n")

	var cfg ConfigDeps
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to unmarshal YAML into ConfigDeps: %v", err)
	}

	if cfg.RequireFSFFree != true {
		t.Fatalf("expected RequireFSFFree=true from YAML, got %v", cfg.RequireFSFFree)
	}
	if cfg.RequireOSIApproved != false {
		t.Fatalf("expected RequireOSIApproved=false from YAML, got %v", cfg.RequireOSIApproved)
	}

	// Backup and restore globals
	savedFSF := requireFSFFree
	savedOSI := requireOSIApproved
	defer func() {
		requireFSFFree = savedFSF
		requireOSIApproved = savedOSI
	}()

	// Apply and verify internal flags updated
	applyRequirementFlags(&cfg)
	if !requireFSFFree {
		t.Fatalf("requireFSFFree should be true after applying from config")
	}
	if requireOSIApproved {
		t.Fatalf("requireOSIApproved should be false after applying from config")
	}

	// Also test nil config resets to defaults (false)
	applyRequirementFlags(nil)
	if requireFSFFree || requireOSIApproved {
		t.Fatalf("expected both flags to be reset to false when config is nil, got fsf=%v osi=%v", requireFSFFree, requireOSIApproved)
	}
}

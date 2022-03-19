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

package config

import "testing"

func TestParseConfig(t *testing.T) {
	config := &Config{

	}
	configFile := ".licenserc_test.yaml"
	config.Parse(configFile)
	if len(config.Header.GetLicenseContent()) == 0 {
		t.Fail()
	}
	t.Logf("Parse config dependency files size = %v", len(config.Deps.Files))
	if len(config.Deps.Files) == 0 {
		t.Fail()
	}
	t.Logf("Parse config languages size = %v", len(config.Languages))
	if len(config.Languages) == 0 {
		t.Fail()
	}
}

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
	"strconv"
	"testing"
	"time"
)

func TestGetLicenseContent(t *testing.T) {
	{
		header := ConfigHeader{
			License: LicenseConfig{
				SpdxID:         "Apache-2.0",
				CopyrightOwner: "Foo",
				SoftwareName:   "Bar",
			},
		}
		expectContent := fmt.Sprintf(
			`Copyright %s Foo

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
`, strconv.Itoa(time.Now().Year()))
		actualContent := header.GetLicenseContent()
		if actualContent != expectContent {
			t.Errorf("GetLicenseContent() result has failure:\n\n%s\n\nWanted:\n\n%s\n", expectContent, actualContent)
		}
	}

	{
		header := ConfigHeader{
			License: LicenseConfig{
				SpdxID:         "Apache-2.0",
				CopyrightOwner: "Apache Software Foundation",
				SoftwareName:   "Bar",
			},
		}
		expectContent := `Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
`
		actualContent := header.GetLicenseContent()
		if actualContent != expectContent {
			t.Errorf("GetLicenseContent() result has failure:\n\n%s\n\nWanted:\n\n%s\n", expectContent, actualContent)
		}
	}

	{
		header := ConfigHeader{
			License: LicenseConfig{
				SpdxID:         "MulanPSL-2.0",
				CopyrightOwner: "Foo",
				SoftwareName:   "Bar",
			},
		}
		expectContent := fmt.Sprintf(
			`Copyright (c) %s Foo
Bar is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
`, strconv.Itoa(time.Now().Year()))
		actualContent := header.GetLicenseContent()
		if actualContent != expectContent {
			t.Errorf("GetLicenseContent() result has failure:\n\n%s\n\nWanted:\n\n%s\n", expectContent, actualContent)
		}
	}
}

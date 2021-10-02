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
package test

import (
	"strconv"
	"testing"
	"time"

	config2 "github.com/apache/skywalking-eyes/pkg/config"
)

func TestConfigHeaderSpdxASF(t *testing.T) {
	c := config2.Config{}
	if err := c.Parse("./testdata/test-spdx-asf.yaml"); err != nil {
		t.Error("unexpected error", err)
	}

	expected := `Licensed to the Apache Software Foundation (ASF) under one
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
	if actual := c.Header.GetLicenseContent(); actual != expected {
		t.Errorf("Actual: \n%v\nExpected: \n%v", actual, expected)
	}
}

func TestConfigHeaderSpdxPlain(t *testing.T) {
	c := config2.Config{}
	if err := c.Parse("./testdata/test-spdx.yaml"); err != nil {
		t.Error("unexpected error", err)
	}

	expected := `Copyright ` + strconv.Itoa(time.Now().Year()) + ` kezhenxu94

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
`
	if actual := c.Header.GetLicenseContent(); actual != expected {
		t.Errorf("Actual: \n%v\nExpected: \n%v", actual, expected)
	}
}

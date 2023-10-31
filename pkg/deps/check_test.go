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

package deps_test

import (
	"strings"
	"testing"

	"github.com/apache/skywalking-eyes/pkg/deps"
)

var TestMatrix = deps.CompatibilityMatrix{
	Compatible: []string{
		"Apache-2.0",
		"PHP-3.01",
		"BSD-3-Clause",
		"BSD-2-Clause",
		"PostgreSQL",
		"EPL-1.0",
		"ISC",
	},
	Incompatible: []string{
		"Unknown",
		"LGPL-2.0+",
		"LGPL-2.0",
		"LGPL-2.0-only",
		"LGPL-2.0-or-later",
		"LGPL-2.1+",
		"LGPL-2.1",
		"LGPL-2.1-only",
		"LGPL-2.1-or-later",
		"LGPL-3.0+",
		"LGPL-3.0",
		"GPL-3.0+",
		"GPL-3.0",
		"GPL-2.0+",
		"GPL-2.0",
		"GPL-2.0-only",
		"GPL-2.0-or-later",
	},
	WeakCompatible: []string{
		"MPL-2.0",
	},
}

func TestCheckWithMatrix(t *testing.T) {
	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "Apache-2.0",
			},
		},
	}, false); err != nil {
		t.Errorf("Shouldn't return error")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "Apache-2.0",
			},
			{
				Dependency:    "Bar",
				LicenseSpdxID: "LGPL-2.0",
			},
		},
	}, false); err == nil {
		t.Errorf("Should return error")
	} else if !strings.Contains(err.Error(), "Bar        | LGPL-2.0") {
		t.Errorf("Should return error and contains dependency Bar, now is `%s`", err.Error())
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "Apache-2.0",
			},
		},
		Skipped: []*deps.Result{
			{
				Dependency:    "Bar",
				LicenseSpdxID: "Unknown",
			},
		},
	}, false); err == nil {
		t.Errorf("Should return error")
	} else if !strings.Contains(err.Error(), "Bar        | Unknown") {
		t.Errorf("Should return error and has dependency Bar, now is `%s`", err.Error())
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "Apache-2.0 OR MIT",
			},
		},
	}, false); err != nil {
		t.Errorf("Shouldn't return error")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "GPL-3.0 and GPL-3.0-or-later",
			},
		},
	}, false); err == nil {
		t.Errorf("Should return error")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "LGPL-2.1-only AND MIT AND BSD-2-Clause",
			},
		},
	}, false); err == nil {
		t.Errorf("Should return error")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "GPL-2.0-or-later WITH Bison-exception-2.2",
			},
		},
	}, false); err == nil {
		t.Errorf("Should return error")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "MPL-2.0",
			},
		},
	}, false); err == nil {
		t.Errorf("Should return error since weak-compatible is turned off")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Bar",
				LicenseSpdxID: "MPL-2.0",
			},
		},
	}, true); err != nil {
		t.Errorf("Shouldn't return error")
	}
}

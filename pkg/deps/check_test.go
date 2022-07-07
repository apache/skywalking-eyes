package deps_test

import (
	"github.com/apache/skywalking-eyes/pkg/deps"
	"strings"
	"testing"
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
}

func TestCheckWithMatrix(t *testing.T) {
	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "Apache-2.0",
			},
		},
	}); err != nil {
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
	}); err == nil {
		t.Errorf("Should return error")
	} else if !strings.Contains(err.Error(), "License: LGPL-2.0 Dependency: Bar") {
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
	}); err == nil {
		t.Errorf("Should return error")
	} else if !strings.Contains(err.Error(), "License: Unknown Dependency: Bar") {
		t.Errorf("Should return error and has dependency Bar, now is `%s`", err.Error())
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "Apache-2.0 OR MIT",
			},
		},
	}); err != nil {
		t.Errorf("Shouldn't return error")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "GPL-3.0 and GPL-3.0-or-later",
			},
		},
	}); err == nil {
		t.Errorf("Should return error")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "LGPL-2.1-only AND MIT AND BSD-2-Clause",
			},
		},
	}); err == nil {
		t.Errorf("Should return error")
	}

	if err := deps.CheckWithMatrix("Apache-2.0", &TestMatrix, &deps.Report{
		Resolved: []*deps.Result{
			{
				Dependency:    "Foo",
				LicenseSpdxID: "GPL-2.0-or-later WITH Bison-exception-2.2",
			},
		},
	}); err == nil {
		t.Errorf("Should return error")
	}
}

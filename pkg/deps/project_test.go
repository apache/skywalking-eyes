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

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var licenses = []struct {
	Licenses    Licenses
	Have        bool
	AllLicenses string
	RawLicenses string
}{
	{Have: false, AllLicenses: "", RawLicenses: "<licenses></licenses>"},
	{Licenses: Licenses{Values: []License{}}, Have: false, AllLicenses: "", RawLicenses: "<licenses></licenses>"},
	{Licenses: Licenses{Values: []License{
		{Name: "The Apache Software License, Version 2.0", URL: "http://www.apache.org/licenses/LICENSE-2.0.txt"},
	}}, Have: true, AllLicenses: "The Apache Software License, Version 2.0", RawLicenses: "<licenses><license><name>The Apache Software License, Version 2.0</name><url>http://www.apache.org/licenses/LICENSE-2.0.txt</url></license></licenses>"},
	{Licenses: Licenses{Values: []License{
		{Name: "The Apache Software License, Version 1.0", URL: "http://www.apache.org/licenses/LICENSE-1.0.txt"},
		{Name: "The Apache Software License, Version 2.0", URL: "http://www.apache.org/licenses/LICENSE-2.0.txt"},
	}}, Have: true, AllLicenses: "The Apache Software License, Version 1.0, The Apache Software License, Version 2.0", RawLicenses: "<licenses><license><name>The Apache Software License, Version 1.0</name><url>http://www.apache.org/licenses/LICENSE-1.0.txt</url></license><license><name>The Apache Software License, Version 2.0</name><url>http://www.apache.org/licenses/LICENSE-2.0.txt</url></license></licenses>"},
}

func TestLicenses(t *testing.T) {
	pom := pomFileWrapper{
		PomFile: NewPomFile(),
	}
	for _, tt := range licenses {
		pom.Licenses = &tt.Licenses

		have := pom.HaveLicenses()
		require.Equal(t, tt.Have, have)

		allLicenses := pom.AllLicenses()
		require.Equal(t, tt.AllLicenses, allLicenses)

		rawLicenses := pom.Raw()
		require.Equal(t, tt.RawLicenses, rawLicenses)
	}
}

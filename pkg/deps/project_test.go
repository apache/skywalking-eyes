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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestProject(t *testing.T) {
	testDataPath, err := filepath.Abs("../../test/testdata/deps_test/maven")
	require.NoError(t, err)

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("%d", time.Now().Unix()))
	err = os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	files, err := ioutil.ReadDir(testDataPath)
	require.NoError(t, err)

	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			pomFile := filepath.Join(testDataPath, file.Name(), "pom.xml")
			err = os.Chdir(filepath.Dir(pomFile))
			require.NoError(t, err)

			project, err := findMaven().NewProject()
			require.NoError(t, err)
			project.splitModules()
			modules := project.modules

			dependenciesFile := filepath.Join(testDataPath, file.Name(), "dependency_tree.txt")
			dependenciesData, err := ioutil.ReadFile(dependenciesFile)
			require.NoError(t, err)
			depTrees := LoadDependenciesTree(dependenciesData)

			require.Equal(t, len(depTrees), len(modules))
			allDepTrees := make(map[string]bool)
			for i := 0; i < len(depTrees); i++ {
				allDepTrees[depTrees[i].Path()] = true
			}

			for i := 0; i < len(depTrees); i++ {
				t.Run(fmt.Sprintf("%d/%d:%s", i+1, len(depTrees), depTrees[i].Path()), func(t *testing.T) {
					expects := make(map[string]bool)
					for _, expect := range depTrees[i].Flatten(allDepTrees) {
						expects[expect.Path()] = true
					}

					module := modules[DepNameNoVersion(&depTrees[i].Dependency)]

					tmpFile, err := os.OpenFile(filepath.Join(tmpDir, pomFileName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
					require.NoError(t, err)

					_, err = tmpFile.Write(module.Encode())
					require.NoError(t, err)

					deps, err := project.loadDependencies(tmpFile.Name())
					require.NoError(t, err)

					actual := map[string]bool{}
					for _, dep := range deps {
						actual[dep.Path()] = true
					}

					for dep := range expects {
						require.True(t, actual[dep])
					}

					err = tmpFile.Close()
					require.NoError(t, err)
				})

			}
		})
	}
}

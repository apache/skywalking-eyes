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
package deps_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/apache/skywalking-eyes/pkg/deps"
	"github.com/stretchr/testify/require"
)

func TestCanResolvePomFile(t *testing.T) {
	resolver := new(deps.MavenPomResolver)
	for _, test := range []struct {
		fileName string
		exp      bool
	}{
		{"pom.xml", true},
		{"POM.XML", false},
		{"log4j-1.2.12.pom", false},
		{".pom", false},
	} {
		b := resolver.CanResolve(test.fileName)
		if b != test.exp {
			t.Errorf("MavenPomResolver.CanResolve(\"%v\") = %v, want %v", test.fileName, b, test.exp)
		}
	}
}

func TestResolveMaven(t *testing.T) {
	testDataPath, err := filepath.Abs("../../test/testdata/deps_test/maven")

	files, err := ioutil.ReadDir(testDataPath)
	require.NoError(t, err)

	for _, file := range files {
		resolver := new(deps.MavenPomResolver)
		pomFile := filepath.Join(testDataPath, file.Name(), "pom.xml")

		t.Run(file.Name(), func(t *testing.T) {
			if resolver.CanResolve(pomFile) {
				report := deps.Report{}
				err := resolver.Resolve(pomFile, &report)
				require.NoError(t, err)

				fmt.Println(report.String())
			}
		})
	}
}

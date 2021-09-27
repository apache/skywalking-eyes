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
	"testing"

	"github.com/apache/skywalking-eyes/pkg/deps"
)

func TestCanResolveJarFile(t *testing.T) {
	resolver := new(deps.JarResolver)
	for _, test := range []struct {
		fileName string
		exp      bool
	}{
		{"1.jar", true},
		{"/tmp/1.jar", true},
		{"1.jar2", false},
		{"protobuf-java-3.13.0.jar", true},
		{"slf4j-api-1.7.25.jar", true},
	} {
		b := resolver.CanResolve(test.fileName)
		if b != test.exp {
			t.Errorf("JarResolver.CanResolve(\"%v\") = %v, want %v", test.fileName, b, test.exp)
		}
	}
}

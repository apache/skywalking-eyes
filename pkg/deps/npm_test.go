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
	"io/ioutil"
	"testing"

	"github.com/apache/skywalking-eyes/pkg/deps"
)

var lcsString = `
{
  "license": "ISC"
}`
var lcsStruct = `
{
  "license": {
    "type" : "ISC",
    "url" : "https://opensource.org/licenses/ISC"
  }
}`
var lcss = `
{
  "licenses": [
    {
      "type": "MIT",
      "url": "https://www.opensource.org/licenses/mit-license.php"
    },
    {
      "type": "Apache-2.0",
      "url": "https://opensource.org/licenses/apache2.0.php"
    }
  ]
}`
var lcsStringEmpty = `
{
  "license": ""
}`
var lcsStructEmpty = `
{
  "license": {
  }
}`
var lcssEmpty = `
{
  "licenses": [
  ]
}`
var lcssInvalid = `
{
  "licenses": {
  }
}`

var TestData = []struct {
	data   string
	result string
	hasErr bool
}{
	{lcsString, "ISC", false},
	{lcsStruct, "ISC", false},
	{lcss, "MIT OR Apache-2.0", false},
	{lcsStringEmpty, "", true},
	{lcsStructEmpty, "", true},
	{lcssEmpty, "", true},
	{lcssInvalid, "", true},
}

func TestResolvePkgFile(t *testing.T) {
	dir := t.TempDir()
	resolver := new(deps.NpmResolver)
	for _, data := range TestData {
		result := &deps.Result{}
		f, err := ioutil.TempFile(dir, "*.json")
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.WriteString(data.data)
		if err != nil {
			t.Fatal(err)
		}
		err = resolver.ResolvePkgFile(result, f.Name())
		if result.LicenseSpdxID != data.result && (err != nil) == data.hasErr {
			t.Fail()
		}
		_ = f.Close()
	}
}

package deps_test

import (
	"io/ioutil"
	"os"
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
	dir, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
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

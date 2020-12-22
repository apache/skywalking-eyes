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
//
package header

import (
	"reflect"
	"testing"

	"github.com/apache/skywalking-eyes/license-eye/pkg/comments"
)

var config = &ConfigHeader{
	License: `Apache License 2.0
  http://www.apache.org/licenses/LICENSE-2.0
Apache License 2.0`,
}

func TestFix(t *testing.T) {
	tests := []struct {
		filename string
		comments string
	}{
		{
			filename: "Test.java",
			comments: `/*
 * Apache License 2.0
 *   http://www.apache.org/licenses/LICENSE-2.0
 * Apache License 2.0
 */
`,
		},
		{
			filename: "Test.py",
			comments: `#
# Apache License 2.0
#   http://www.apache.org/licenses/LICENSE-2.0
# Apache License 2.0
`,
		},
	}
	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			style := comments.FileCommentStyle(test.filename)
			if c, err := generateLicenseHeader(style, config); err != nil || c != test.comments {
				t.Log("Actual:", c)
				t.Log("Expected:", test.comments)
				t.Logf("Middle:'%v'\n", style.Middle)
				t.Log(err)
				t.Fail()
			}
		})
	}
}

func TestRewriteContent(t *testing.T) {
	tests := []struct {
		name            string
		style           *comments.CommentStyle
		content         string
		licenseHeader   string
		expectedContent string
	}{
		{
			name:  "Ocaml",
			style: comments.FileCommentStyle("test.ml"),
			content: `print_string "hello worlds!\n";;
`,
			licenseHeader: getLicenseHeader("test.ml", t.Error),
			expectedContent: `(*
(* Apache License 2.0
(*   http://www.apache.org/licenses/LICENSE-2.0
(* Apache License 2.0
print_string "hello worlds!\n";;
`},
		{
			name:  "Python with Shebang",
			style: comments.FileCommentStyle("test.py"),
			content: `
#!/usr/bin/env python3
if __name__ == '__main__':
    print('Hello World')
`,
			licenseHeader: getLicenseHeader("test.py", t.Error),
			expectedContent: `#!/usr/bin/env python3
#
# Apache License 2.0
#   http://www.apache.org/licenses/LICENSE-2.0
# Apache License 2.0
if __name__ == '__main__':
    print('Hello World')
`},
		{
			name:  "Python",
			style: comments.FileCommentStyle("test.py"),
			content: `
if __name__ == '__main__':
    print('Hello World')
`,
			licenseHeader: getLicenseHeader("test.py", t.Error),
			expectedContent: `#
# Apache License 2.0
#   http://www.apache.org/licenses/LICENSE-2.0
# Apache License 2.0
if __name__ == '__main__':
    print('Hello World')
`},
		{
			name:  "XML one line declaration",
			style: comments.FileCommentStyle("test.xml"),
			content: `
<?xml version="1.0" encoding="UTF-8"?>
<project>
  <modelVersion>4.0.0</modelVersion>
</project>
`,
			licenseHeader: getLicenseHeader("test.xml", t.Error),
			expectedContent: `<?xml version="1.0" encoding="UTF-8"?>
<!--
  ~ Apache License 2.0
  ~   http://www.apache.org/licenses/LICENSE-2.0
  ~ Apache License 2.0
-->
<project>
  <modelVersion>4.0.0</modelVersion>
</project>
`},
		{
			name:  "XML multi-line declaration",
			style: comments.FileCommentStyle("test.xml"),
			content: `
<?xml
  version="1.0"
  encoding="UTF-8"
?>
<project>
  <modelVersion>4.0.0</modelVersion>
</project>
`,
			licenseHeader: getLicenseHeader("test.xml", t.Error),
			expectedContent: `<?xml
  version="1.0"
  encoding="UTF-8"
?>
<!--
  ~ Apache License 2.0
  ~   http://www.apache.org/licenses/LICENSE-2.0
  ~ Apache License 2.0
-->
<project>
  <modelVersion>4.0.0</modelVersion>
</project>
`},
		{
			name:          "SQL",
			style:         comments.FileCommentStyle("test.sql"),
			content:       `select * from user;`,
			licenseHeader: getLicenseHeader("test.sql", t.Error),
			expectedContent: `--
-- Apache License 2.0
--   http://www.apache.org/licenses/LICENSE-2.0
-- Apache License 2.0
select * from user;`},
		{
			name:          "Haskell",
			style:         comments.FileCommentStyle("test.hs"),
			content:       `import Foundation.Hashing.Hashable`,
			licenseHeader: getLicenseHeader("test.hs", t.Error),
			expectedContent: `{-
 Apache License 2.0
   http://www.apache.org/licenses/LICENSE-2.0
 Apache License 2.0
-}
import Foundation.Hashing.Hashable`},
		{
			name:  "Vim",
			style: comments.FileCommentStyle("test.vim"),
			content: `echo 'Hello' | echo 'world!'
`,
			licenseHeader: getLicenseHeader("test.vim", t.Error),
			expectedContent: `"
" Apache License 2.0
"   http://www.apache.org/licenses/LICENSE-2.0
" Apache License 2.0
echo 'Hello' | echo 'world!'
`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := rewriteContent(test.style, []byte(test.content), test.licenseHeader)
			if !reflect.DeepEqual(content, []byte(test.expectedContent)) {
				t.Log("Actual\n" + string(content))
				t.Log("Expected\n" + test.expectedContent)
				t.Fail()
			}
		})
	}
}

func getLicenseHeader(filename string, tError func(args ...interface{})) string {
	s, err := generateLicenseHeader(comments.FileCommentStyle(filename), config)
	if err != nil {
		tError(err)
	}
	return s
}

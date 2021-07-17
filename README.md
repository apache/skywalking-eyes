# SkyWalking Eyes

<img src="http://skywalking.apache.org/assets/logo.svg" alt="Sky Walking logo" height="90px" align="right" />

A full-featured license tool to check and fix license headers and resolve dependencies' licenses.

[![Twitter Follow](https://img.shields.io/twitter/follow/asfskywalking.svg?style=for-the-badge&label=Follow&logo=twitter)](https://twitter.com/AsfSkyWalking)

## Usage

You can use License-Eye in GitHub Actions or in your local machine.

### GitHub Actions

To use License-Eye in GitHub Actions, add a step in your GitHub workflow.

```yaml
- name: Check License Header
  uses: apache/skywalking-eyes@main      # always prefer to use a revision instead of `main`.
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}  # needed only when you want License-Eye to comment on the pull request.
```

Add a `.licenserc.yaml` in the root of your project, for Apache Software Foundation projects, the following configuration should be enough.

```yaml
header:
  license:
    spdx-id: Apache-2.0
    copyright-owner: Apache Software Foundation

  paths-ignore:
    - 'dist'
    - 'licenses'
    - '**/*.md'
    - 'LICENSE'
    - 'NOTICE'

  comment: on-failure
```

**NOTE**: The full configurations can be found in [the configuration section](#configurations).

### Docker Image

```shell
$ docker run -it --rm -v $(pwd):/github/workspace apache/skywalking-eyes header check
$ docker run -it --rm -v $(pwd):/github/workspace apache/skywalking-eyes header fix
```

### Compile from Source

```bash
$ git clone https://github.com/apache/skywalking-eyes
$ cd skywalking-eyes
$ make build
```

#### Check License Header

```bash
$ bin/darwin/license-eye -c test/testdata/.licenserc_for_test_check.yaml header check

INFO Loading configuration from file: test/testdata/.licenserc_for_test_check.yaml 
INFO Totally checked 30 files, valid: 12, invalid: 12, ignored: 6, fixed: 0 
ERROR the following files don't have a valid license header: 
test/testdata/include_test/without_license/testcase.go
test/testdata/include_test/without_license/testcase.graphql
test/testdata/include_test/without_license/testcase.ini
test/testdata/include_test/without_license/testcase.java
test/testdata/include_test/without_license/testcase.md
test/testdata/include_test/without_license/testcase.php
test/testdata/include_test/without_license/testcase.py
test/testdata/include_test/without_license/testcase.sh
test/testdata/include_test/without_license/testcase.yaml
test/testdata/include_test/without_license/testcase.yml
test/testdata/test-spdx-asf.yaml
test/testdata/test-spdx.yaml 
exit status 1
```

#### Fix License Header

```bash
$ bin/darwin/license-eye -c test/testdata/.licenserc_for_test_fix.yaml header fix

INFO Loading configuration from file: test/testdata/.licenserc_for_test_fix.yaml
INFO Totally checked 20 files, valid: 10, invalid: 10, ignored: 0, fixed: 10 
```

#### Resolve Dependencies' licenses

```bash
$ bin/darwin/license-eye -c test/testdata/.licenserc_for_test_check.yaml dep resolve
INFO GITHUB_TOKEN is not set, license-eye won't comment on the pull request
INFO Loading configuration from file: test/testdata/.licenserc_for_test_check.yaml
WARNING Failed to resolve the license of dependency: gopkg.in/yaml.v3 cannot identify license content
Dependency                                  |      License
------------------------------------------- | ------------
github.com/bmatcuk/doublestar/v2            |          MIT
github.com/sirupsen/logrus                  |          MIT
golang.org/x/sys/unix                       | BSD-3-Clause
github.com/spf13/cobra                      |   Apache-2.0
github.com/spf13/pflag                      | BSD-3-Clause
vendor/golang.org/x/net/dns/dnsmessage      | BSD-3-Clause
vendor/golang.org/x/net/route               | BSD-3-Clause
golang.org/x/oauth2                         | BSD-3-Clause
golang.org/x/oauth2/internal                | BSD-3-Clause
vendor/golang.org/x/crypto/cryptobyte       | BSD-3-Clause
vendor/golang.org/x/crypto/cryptobyte/asn1  | BSD-3-Clause
golang.org/x/net/context/ctxhttp            | BSD-3-Clause
vendor/golang.org/x/crypto/chacha20poly1305 | BSD-3-Clause
vendor/golang.org/x/crypto/chacha20         | BSD-3-Clause
vendor/golang.org/x/crypto/internal/subtle  | BSD-3-Clause
vendor/golang.org/x/crypto/poly1305         | BSD-3-Clause
vendor/golang.org/x/sys/cpu                 | BSD-3-Clause
vendor/golang.org/x/crypto/curve25519       | BSD-3-Clause
vendor/golang.org/x/crypto/hkdf             | BSD-3-Clause
vendor/golang.org/x/net/http/httpguts       | BSD-3-Clause
vendor/golang.org/x/net/idna                | BSD-3-Clause
vendor/golang.org/x/text/secure/bidirule    | BSD-3-Clause
vendor/golang.org/x/text/transform          | BSD-3-Clause
vendor/golang.org/x/text/unicode/bidi       | BSD-3-Clause
vendor/golang.org/x/text/unicode/norm       | BSD-3-Clause
vendor/golang.org/x/net/http/httpproxy      | BSD-3-Clause
vendor/golang.org/x/net/http2/hpack         | BSD-3-Clause
gopkg.in/yaml.v3                            |      Unknown

ERROR failed to identify the licenses of following packages:
gopkg.in/yaml.v3
```

## Configurations

```yaml
header: # <1>
  license:
    spdx-id: Apache-2.0 # <2>
    copyright-owner: Apache Software Foundation # <3>
    content: | # <4>
      Licensed to Apache Software Foundation (ASF) under one or more contributor
      license agreements. See the NOTICE file distributed with
      this work for additional information regarding copyright
      ownership. Apache Software Foundation (ASF) licenses this file to you under
      the Apache License, Version 2.0 (the "License"); you may
      not use this file except in compliance with the License.
      You may obtain a copy of the License at

          http://www.apache.org/licenses/LICENSE-2.0

      Unless required by applicable law or agreed to in writing,
      software distributed under the License is distributed on an
      "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
      KIND, either express or implied.  See the License for the
      specific language governing permissions and limitations
      under the License.

    pattern: | # <5>
      Licensed to the Apache Software Foundation under one or more contributor
      license agreements. See the NOTICE file distributed with
      this work for additional information regarding copyright
      ownership. The Apache Software Foundation licenses this file to you under
      the Apache License, Version 2.0 \(the "License"\); you may
      not use this file except in compliance with the License.
      You may obtain a copy of the License at

          http://www.apache.org/licenses/LICENSE-2.0

      Unless required by applicable law or agreed to in writing,
      software distributed under the License is distributed on an
      "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
      KIND, either express or implied.  See the License for the
      specific language governing permissions and limitations
      under the License.

  paths: # <6>
    - '**'

  paths-ignore: # <7>
    - 'dist'
    - 'licenses'
    - '**/*.md'
    - '**/testdata/**'
    - '**/go.mod'
    - '**/go.sum'
    - 'LICENSE'
    - 'NOTICE'
    - '**/assets/languages.yaml'
    - '**/assets/assets.gen.go'

  comment: on-failure # <8>

dependency: # <9>
  files: # <10>
    - go.mod
```

1. The `header` section is configurations for source codes license header.
2. The [SPDX ID](https://spdx.org/licenses/) of the license, it’s convenient when your license is standard SPDX license, so that you can simply specify this identifier without copying the whole license `content` or `pattern`. This will be used as the content when `fix` command needs to insert a license header.
3. The copyright owner to replace the `[owner]` in the `SPDX-ID` license template.
4. If you are not using the standard license text, you can paste your license text here, this will be used as the content when `fix` command needs to insert a license header, if both `license` and `SPDX-ID` are specified, `license` wins.
5. The `pattern` is an optional regexp. You don’t need this if all the file headers are the same as `license` or the license of `SPDX-ID`, otherwise you need to compose a pattern that matches your license texts.
6. The `paths` are the path list that will be checked (and fixed) by license-eye, default is `['**']`. Formats like `**/*`.md and `**/bin/**` are supported.
7. The `paths-ignore` are the path list that will be ignored by license-eye. By default, `.git` and the content in `.gitignore` will be inflated into the `paths-ignore` list.
8. On what condition License-Eye will comment the check results on the pull request, `on-failure`, `always` or `never`. Options other than `never` require the environment variable `GITHUB_TOKEN` to be set.
9. `dependency` section is configurations for resolving dependencies' licenses.
10. `files` are the files that declare the dependencies of a project, typically, `go.mo` in Go project, `pom.xml` in maven project, and `package.json` in NodeJS project. If it's a relative path, it's relative to the `.licenserc.yaml`.

**NOTE**: When the `SPDX-ID` is Apache-2.0 and the owner is Apache Software foundation, the content would be [a dedicated license](https://www.apache.org/legal/src-headers.html#headers) specified by the ASF, otherwise, the license would be [the standard one](https://www.apache.org/foundation/license-faq.html#Apply-My-Software).

## Supported File Types

The `header check` command theoretically supports all kinds of file types, while the supported file types of `header fix` command can be found [in this YAML file](assets/languages.yaml). In the YAML file, if the language has a non-empty property `comment_style_id`, and the comment style id is declared in [the comment styles file](assets/styles.yaml), then the language is supported by `fix` command.

- [assets/languages.yaml](assets/languages.yaml)

  ```yaml
  Java:
    type: programming
    tm_scope: source.java
    ace_mode: java
    codemirror_mode: clike
    codemirror_mime_type: text/x-java
    color: "#b07219"
    extensions:
      - ".java"
    language_id: 181
    comment_style_id: SlashAsterisk
  ```

- [assets/languages.yaml](assets/languages.yaml)

  ```yaml
  - id: SlashAsterisk     # (i)
    start: '/*'           # (ii)
    middle: ' *'          # (iii)
    end: ' */'            # (iv)
  ```

  1. The `comment_style_id` used in [assets/languages.yaml](assets/languages.yaml).
  2. The leading characters of the starting of a block comment.
  3. The leading characters of the middle lines of a block comment.
  4. The leading characters of the ending line of a block comment.

## Technical Documentation

- There is an [activity diagram](https://www.plantuml.com/plantuml/proxy?cache=yes&format=svg&src=https://raw.githubusercontent.com/apache/skywalking-eyes/main/docs/header_fix_logic.plantuml) explaining the implemented license header 
  fixing mechanism in-depth. The diagram's source file can be found [here](./docs/header_fix_logic.plantuml).

## Contribution

- If you find any file type should be supported by the aforementioned configurations, but it's not listed there, feel free to [open a pull request](https://github.com/apache/skywalking-eyes/pulls) to add the configuration into the two files.
- If you find the license template of an SPDX ID is not supported, feel free to [open a pull request](https://github.com/apache/skywalking-eyes/pulls) to add it into [the template folder](assets/header-templates).

## License

[Apache License 2.0](https://github.com/apache/skywalking-eyes/blob/master/LICENSE)

## Contact Us
* Submit [an issue](https://github.com/apache/skywalking/issues/new) by using [INFRA] as title prefix.
* Mail list: **dev@skywalking.apache.org**. Mail to dev-subscribe@skywalking.apache.org, follow the reply to subscribe the mail list.
* Join `skywalking` channel at [Apache Slack](http://s.apache.org/slack-invite). If the link is not working, find the latest one at [Apache INFRA WIKI](https://cwiki.apache.org/confluence/display/INFRA/Slack+Guest+Invites).
* Twitter, [ASFSkyWalking](https://twitter.com/ASFSkyWalking)

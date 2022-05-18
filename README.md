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
  # with:
      # log: debug # optional: set the log level. The default value is `info`.
      # config: .licenserc.yaml # optional: set the config file. The default value is `.licenserc.yaml`.
      # token: # optional: the token that license eye uses when it needs to comment on the pull request. Set to empty ("") to disable commenting on pull request. The default value is ${{ github.token }}
      # mode: # optional: Which mode License-Eye should be run in. Choices are `check` or `fix`. The default value is `check`.
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

#### Using the Action in Fix Mode

By default the action runs License-Eye in check mode, which will raise an error
if any of the processed files are missing license headers. If `mode` is set to
`fix`, the action will instead apply the license header to any processed file
that is missing a license header. The fixed files can then be pushed back to the
pull request using another GitHub action. For example:

```yaml
- name: Check License Header
  uses: apache/skywalking-eyes@main
    with:
      mode: fix
- name: Apply Changes
  uses: EndBug/add-and-commit@v4
    with:
      author_name: License Bot
      author_email: license_bot@github.com
      message: 'Automatic application of license header'
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Note**: The exit code of fix mode is always 0 and can not be used to block CI
status. Consider running the action in check mode if you would like CI to fail
when a file is missing a license header.

### Docker Image

```shell
docker run -it --rm -v $(pwd):/github/workspace apache/skywalking-eyes header check
docker run -it --rm -v $(pwd):/github/workspace apache/skywalking-eyes header fix
```

### Docker Image from the latest codes

For users and developers who want to help to test the latest codes on main branch, we publish Docker image to GitHub
Container Registry for every commit in main branch, tagged with the commit sha, if it's the latest commit in main
branch, it's also tagged with `latest`.

**Note**: these Docker images are not official Apache releases. For official releases, please refer to
[the download page](https://skywalking.apache.org/downloads/#SkyWalkingEyes) for executable binary and
[the Docker hub](https://hub.docker.com/r/apache/skywalking-eyes) for Docker images.

```shell
docker run -it --rm -v $(pwd):/github/workspace ghcr.io/apache/skywalking-eyes/license-eye header check
docker run -it --rm -v $(pwd):/github/workspace ghcr.io/apache/skywalking-eyes/license-eye header fix
```

### Compile from Source

```bash
git clone https://github.com/apache/skywalking-eyes
cd skywalking-eyes
make build
```

If you have Go SDK installed, you can also use `go install` command to install the latest code.

```bash
go install github.com/apache/skywalking-eyes/cmd/license-eye@latest
```

#### Check License Header

```bash
license-eye -c test/testdata/.licenserc_for_test_check.yaml header check
```

<details>
<summary>Header Check Result</summary>

```
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

</details>

#### Fix License Header

```bash
bin/darwin/license-eye -c test/testdata/.licenserc_for_test_fix.yaml header fix
```

<details>
<summary>Header Fix Result</summary>

```
INFO Loading configuration from file: test/testdata/.licenserc_for_test_fix.yaml
INFO Totally checked 20 files, valid: 10, invalid: 10, ignored: 0, fixed: 10
```

</details>

#### Resolve Dependencies' licenses

This command serves as assistance for human beings to audit the dependencies license, it's exit code is always 0.

We also support two flags:

|Flag name|Short name|Description|
|---------|----------|-----------|
|`--output`|`-o`|Save the dependencies' `LICENSE` files to a specified directory so that you can put them in distribution package if needed.|
|`--summary`|`-s`|Based on the template, aggregate all dependency information and generate a `LICENSE` file.|

```bash
license-eye -c test/testdata/.licenserc_for_test_check.yaml dep resolve -o ./dependencies/licenses -s LICENSE.tpl
```

<details>
<summary>Dependency Resolve Result</summary>

```
INFO GITHUB_TOKEN is not set, license-eye won't comment on the pull request
INFO Loading configuration from file: test/testdata/.licenserc_for_test_check.yaml
WARNING Failed to resolve the license of <github.com/gogo/protobuf>: cannot identify license content
WARNING Failed to resolve the license of <github.com/kr/logfmt>: cannot find license file
WARNING Failed to resolve the license of <github.com/magiconair/properties>: cannot identify license content
WARNING Failed to resolve the license of <github.com/miekg/dns>: cannot identify license content
WARNING Failed to resolve the license of <github.com/pascaldekloe/goe>: cannot identify license content
WARNING Failed to resolve the license of <github.com/russross/blackfriday/v2>: cannot identify license content
WARNING Failed to resolve the license of <gopkg.in/check.v1>: cannot identify license content
Dependency                                         |        License |                              Version
-------------------------------------------------- | -------------- | ------------------------------------
cloud.google.com/go                                |     Apache-2.0 |                              v0.46.3
cloud.google.com/go/bigquery                       |     Apache-2.0 |                               v1.0.1
cloud.google.com/go/datastore                      |     Apache-2.0 |                               v1.0.0
cloud.google.com/go/firestore                      |     Apache-2.0 |                               v1.1.0
cloud.google.com/go/pubsub                         |     Apache-2.0 |                               v1.0.1
cloud.google.com/go/storage                        |     Apache-2.0 |                               v1.0.0
dmitri.shuralyov.com/gpu/mtl                       |   BSD-3-Clause |   v0.0.0-20190408044501-666a987793e9
github.com/BurntSushi/toml                         |            MIT |                               v0.3.1
github.com/BurntSushi/xgb                          |   BSD-3-Clause |   v0.0.0-20160522181843-27f122750802
github.com/OneOfOne/xxhash                         |     Apache-2.0 |                               v1.2.2
github.com/alecthomas/template                     |   BSD-3-Clause |   v0.0.0-20160405071501-a0175ee3bccc
github.com/alecthomas/units                        |            MIT |   v0.0.0-20151022065526-2efee857e7cf
github.com/armon/circbuf                           |            MIT |   v0.0.0-20150827004946-bbbad097214e
github.com/armon/go-metrics                        |            MIT |   v0.0.0-20180917152333-f0300d1749da
github.com/armon/go-radix                          |            MIT |   v0.0.0-20180808171621-7fddfc383310
github.com/beorn7/perks                            |            MIT |                               v1.0.0
github.com/bgentry/speakeasy                       |            MIT |                               v0.1.0
github.com/bketelsen/crypt                         |            MIT | v0.0.3-0.20200106085610-5cbc8cc4026c
github.com/bmatcuk/doublestar/v2                   |            MIT |                               v2.0.4
github.com/cespare/xxhash                          |            MIT |                               v1.1.0
github.com/client9/misspell                        |            MIT |                               v0.3.4
github.com/coreos/bbolt                            |            MIT |                               v1.3.2
github.com/coreos/etcd                             |     Apache-2.0 |                 v3.3.13+incompatible
github.com/coreos/go-semver                        |     Apache-2.0 |                               v0.3.0
github.com/coreos/go-systemd                       |     Apache-2.0 |   v0.0.0-20190321100706-95778dfbb74e
github.com/coreos/pkg                              |     Apache-2.0 |   v0.0.0-20180928190104-399ea9e2e55f
github.com/cpuguy83/go-md2man/v2                   |            MIT |                               v2.0.0
github.com/davecgh/go-spew                         |            ISC |                               v1.1.1
github.com/dgrijalva/jwt-go                        |            MIT |                  v3.2.0+incompatible
github.com/dgryski/go-sip13                        |            MIT |   v0.0.0-20181026042036-e10d5fee7954
github.com/fatih/color                             |            MIT |                               v1.7.0
github.com/fsnotify/fsnotify                       |   BSD-3-Clause |                               v1.4.7
github.com/ghodss/yaml                             |            MIT |                               v1.0.0
github.com/go-gl/glfw                              |   BSD-3-Clause |   v0.0.0-20190409004039-e6da0acd62b1
github.com/go-kit/kit                              |            MIT |                               v0.8.0
github.com/go-logfmt/logfmt                        |            MIT |                               v0.4.0
github.com/go-stack/stack                          |            MIT |                               v1.8.0
github.com/golang/glog                             |     Apache-2.0 |   v0.0.0-20160126235308-23def4e6c14b
github.com/golang/groupcache                       |     Apache-2.0 |   v0.0.0-20190129154638-5b532d6fd5ef
github.com/golang/mock                             |     Apache-2.0 |                               v1.3.1
github.com/golang/protobuf                         |   BSD-3-Clause |                               v1.3.2
github.com/google/btree                            |     Apache-2.0 |                               v1.0.0
github.com/google/go-cmp                           |   BSD-3-Clause |                               v0.3.0
github.com/google/go-github/v33                    |   BSD-3-Clause |                              v33.0.0
github.com/google/go-querystring                   |   BSD-3-Clause |                               v1.0.0
github.com/google/martian                          |     Apache-2.0 |                  v2.1.0+incompatible
github.com/google/pprof                            |     Apache-2.0 |   v0.0.0-20190515194954-54271f7e092f
github.com/google/renameio                         |     Apache-2.0 |                               v0.1.0
github.com/googleapis/gax-go/v2                    |   BSD-3-Clause |                               v2.0.5
github.com/gopherjs/gopherjs                       |   BSD-2-Clause |   v0.0.0-20181017120253-0766667cb4d1
github.com/gorilla/websocket                       |   BSD-2-Clause |                               v1.4.2
github.com/grpc-ecosystem/go-grpc-middleware       |     Apache-2.0 |                               v1.0.0
github.com/grpc-ecosystem/go-grpc-prometheus       |     Apache-2.0 |                               v1.2.0
github.com/grpc-ecosystem/grpc-gateway             |   BSD-3-Clause |                               v1.9.0
github.com/hashicorp/consul/api                    |        MPL-2.0 |                               v1.1.0
github.com/hashicorp/consul/sdk                    |        MPL-2.0 |                               v0.1.1
github.com/hashicorp/errwrap                       |        MPL-2.0 |                               v1.0.0
github.com/hashicorp/go-cleanhttp                  |        MPL-2.0 |                               v0.5.1
github.com/hashicorp/go-immutable-radix            |        MPL-2.0 |                               v1.0.0
github.com/hashicorp/go-msgpack                    |   BSD-3-Clause |                               v0.5.3
github.com/hashicorp/go-multierror                 |        MPL-2.0 |                               v1.0.0
github.com/hashicorp/go-rootcerts                  |        MPL-2.0 |                               v1.0.0
github.com/hashicorp/go-sockaddr                   |        MPL-2.0 |                               v1.0.0
github.com/hashicorp/go-syslog                     |            MIT |                               v1.0.0
github.com/hashicorp/go-uuid                       |        MPL-2.0 |                               v1.0.1
github.com/hashicorp/go.net                        |   BSD-3-Clause |                               v0.0.1
github.com/hashicorp/golang-lru                    |        MPL-2.0 |                               v0.5.1
github.com/hashicorp/hcl                           |        MPL-2.0 |                               v1.0.0
github.com/hashicorp/logutils                      |        MPL-2.0 |                               v1.0.0
github.com/hashicorp/mdns                          |            MIT |                               v1.0.0
github.com/hashicorp/memberlist                    |        MPL-2.0 |                               v0.1.3
github.com/hashicorp/serf                          |        MPL-2.0 |                               v0.8.2
github.com/inconshreveable/mousetrap               |     Apache-2.0 |                               v1.0.0
github.com/jonboulle/clockwork                     |     Apache-2.0 |                               v0.1.0
github.com/json-iterator/go                        |            MIT |                               v1.1.6
github.com/jstemmer/go-junit-report                |            MIT |   v0.0.0-20190106144839-af01ea7f8024
github.com/jtolds/gls                              |            MIT |                 v4.20.0+incompatible
github.com/julienschmidt/httprouter                |   BSD-3-Clause |                               v1.2.0
github.com/kisielk/errcheck                        |            MIT |                               v1.1.0
github.com/kisielk/gotool                          |            MIT |                               v1.0.0
github.com/konsorten/go-windows-terminal-sequences |            MIT |                               v1.0.1
github.com/kr/pretty                               |            MIT |                               v0.1.0
github.com/kr/pty                                  |            MIT |                               v1.1.1
github.com/kr/text                                 |            MIT |                               v0.1.0
github.com/mattn/go-colorable                      |            MIT |                               v0.0.9
github.com/mattn/go-isatty                         |            MIT |                               v0.0.3
github.com/matttproud/golang_protobuf_extensions   |     Apache-2.0 |                               v1.0.1
github.com/mitchellh/cli                           |        MPL-2.0 |                               v1.0.0
github.com/mitchellh/go-homedir                    |            MIT |                               v1.1.0
github.com/mitchellh/go-testing-interface          |            MIT |                               v1.0.0
github.com/mitchellh/gox                           |        MPL-2.0 |                               v0.4.0
github.com/mitchellh/iochan                        |            MIT |                               v1.0.0
github.com/mitchellh/mapstructure                  |            MIT |                               v1.1.2
github.com/modern-go/concurrent                    |     Apache-2.0 |   v0.0.0-20180306012644-bacd9c7ef1dd
github.com/modern-go/reflect2                      |     Apache-2.0 |                               v1.0.1
github.com/mwitkow/go-conntrack                    |     Apache-2.0 |   v0.0.0-20161129095857-cc309e4a2223
github.com/oklog/ulid                              |     Apache-2.0 |                               v1.3.1
github.com/pelletier/go-toml                       |            MIT |                               v1.2.0
github.com/pkg/errors                              |   BSD-2-Clause |                               v0.8.1
github.com/pmezard/go-difflib                      |   BSD-3-Clause |                               v1.0.0
github.com/posener/complete                        |            MIT |                               v1.1.1
github.com/prometheus/client_golang                |     Apache-2.0 |                               v0.9.3
github.com/prometheus/client_model                 |     Apache-2.0 |   v0.0.0-20190129233127-fd36f4220a90
github.com/prometheus/common                       |     Apache-2.0 |                               v0.4.0
github.com/prometheus/procfs                       |     Apache-2.0 |   v0.0.0-20190507164030-5867b95ac084
github.com/prometheus/tsdb                         |     Apache-2.0 |                               v0.7.1
github.com/rogpeppe/fastuuid                       |   BSD-3-Clause |   v0.0.0-20150106093220-6724a57986af
github.com/rogpeppe/go-internal                    |   BSD-3-Clause |                               v1.3.0
github.com/ryanuber/columnize                      |            MIT |   v0.0.0-20160712163229-9b3edd62028f
github.com/sean-/seed                              |            MIT |   v0.0.0-20170313163322-e2103e2c3529
github.com/shurcooL/sanitized_anchor_name          |            MIT |                               v1.0.0
github.com/sirupsen/logrus                         |            MIT |                               v1.7.0
github.com/smartystreets/assertions                |            MIT |   v0.0.0-20180927180507-b2de0cb4f26d
github.com/smartystreets/goconvey                  |            MIT |                               v1.6.4
github.com/soheilhy/cmux                           |     Apache-2.0 |                               v0.1.4
github.com/spaolacci/murmur3                       |   BSD-3-Clause |   v0.0.0-20180118202830-f09979ecbc72
github.com/spf13/afero                             |     Apache-2.0 |                               v1.1.2
github.com/spf13/cast                              |            MIT |                               v1.3.0
github.com/spf13/cobra                             |     Apache-2.0 |                               v1.1.1
github.com/spf13/jwalterweatherman                 |            MIT |                               v1.0.0
github.com/spf13/pflag                             |   BSD-3-Clause |                               v1.0.5
github.com/spf13/viper                             |            MIT |                               v1.7.0
github.com/stretchr/objx                           |            MIT |                               v0.1.1
github.com/stretchr/testify                        |            MIT |                               v1.3.0
github.com/subosito/gotenv                         |            MIT |                               v1.2.0
github.com/tmc/grpc-websocket-proxy                |            MIT |   v0.0.0-20190109142713-0ad062ec5ee5
github.com/xiang90/probing                         |            MIT |   v0.0.0-20190116061207-43a291ad63a2
github.com/yuin/goldmark                           |            MIT |                               v1.3.5
go.etcd.io/bbolt                                   |            MIT |                               v1.3.2
go.opencensus.io                                   |     Apache-2.0 |                              v0.22.0
go.uber.org/atomic                                 |            MIT |                               v1.4.0
go.uber.org/multierr                               |            MIT |                               v1.1.0
go.uber.org/zap                                    |            MIT |                              v1.10.0
golang.org/x/crypto                                |   BSD-3-Clause |   v0.0.0-20191011191535-87dc89f01550
golang.org/x/exp                                   |   BSD-3-Clause |   v0.0.0-20191030013958-a1ab85dbe136
golang.org/x/image                                 |   BSD-3-Clause |   v0.0.0-20190802002840-cff245a6509b
golang.org/x/lint                                  |   BSD-3-Clause |   v0.0.0-20190930215403-16217165b5de
golang.org/x/mobile                                |   BSD-3-Clause |   v0.0.0-20190719004257-d2bd2a29d028
golang.org/x/mod                                   |   BSD-3-Clause |                               v0.4.2
golang.org/x/net                                   |   BSD-3-Clause |   v0.0.0-20210726213435-c6fcb2dbf985
golang.org/x/oauth2                                |   BSD-3-Clause |   v0.0.0-20190604053449-0f29369cfe45
golang.org/x/sync                                  |   BSD-3-Clause |   v0.0.0-20210220032951-036812b2e83c
golang.org/x/sys                                   |   BSD-3-Clause |   v0.0.0-20210510120138-977fb7262007
golang.org/x/term                                  |   BSD-3-Clause |   v0.0.0-20201126162022-7de9c90e9dd1
golang.org/x/text                                  |   BSD-3-Clause |                               v0.3.6
golang.org/x/time                                  |   BSD-3-Clause |   v0.0.0-20190308202827-9d24e82272b4
golang.org/x/tools                                 |   BSD-3-Clause |                               v0.1.5
golang.org/x/xerrors                               |   BSD-3-Clause |   v0.0.0-20200804184101-5ec99f83aff1
google.golang.org/api                              |   BSD-3-Clause |                              v0.13.0
google.golang.org/appengine                        |     Apache-2.0 |                               v1.6.1
google.golang.org/genproto                         |     Apache-2.0 |   v0.0.0-20191108220845-16a3f7862a1a
google.golang.org/grpc                             |     Apache-2.0 |                              v1.21.1
gopkg.in/alecthomas/kingpin.v2                     |            MIT |                               v2.2.6
gopkg.in/errgo.v2                                  |   BSD-3-Clause |                               v2.1.0
gopkg.in/ini.v1                                    |     Apache-2.0 |                              v1.51.0
gopkg.in/resty.v1                                  |            MIT |                              v1.12.0
gopkg.in/yaml.v2                                   |     Apache-2.0 |                               v2.2.8
gopkg.in/yaml.v3                                   | MIT and Apache |   v3.0.0-20200615113413-eeeca48fe776
honnef.co/go/tools                                 |            MIT |                      v0.0.1-2019.2.3
rsc.io/binaryregexp                                |   BSD-3-Clause |                               v0.2.0
github.com/gogo/protobuf                           |        Unknown |                               v1.2.1
github.com/kr/logfmt                               |        Unknown |   v0.0.0-20140226030751-b84e30acd515
github.com/magiconair/properties                   |        Unknown |                               v1.8.1
github.com/miekg/dns                               |        Unknown |                              v1.0.14
github.com/pascaldekloe/goe                        |        Unknown |   v0.0.0-20180627143212-57f6aae5913c
github.com/russross/blackfriday/v2                 |        Unknown |                               v2.0.1
gopkg.in/check.v1                                  |        Unknown |   v1.0.0-20180628173108-788fd7840127

ERROR failed to identify the licenses of following packages (7):
github.com/gogo/protobuf
github.com/kr/logfmt
github.com/magiconair/properties
github.com/miekg/dns
github.com/pascaldekloe/goe
github.com/russross/blackfriday/v2
gopkg.in/check.v1
```

</details>

##### Summary Template

The summary is a template to generate the summary of dependencies' licenses based on the [Golang Template](https://pkg.go.dev/text/template). It includes these variables:

|Name|Type|Example|Description|
|----|----|-------|-----------|
|LicenseContent|string|`{{.LicenseContent}}`|The project license content, it's the license of `header.license.spdx-id` (if set), otherwise it's the `header.license.content`. |
|Groups|list structure|`{{ range .Groups }}`|The dependency groups, all licenses are grouped by the same license [SPDX ID](https://spdx.org/licenses/). |
|Groups.LicenseID|string|`{{.LicenseID}}`|The [SPDX ID](https://spdx.org/licenses/) of dependency. |
|Groups.Deps|list structure|`{{ range .Deps }}`|All dependencies with the same [SPDX ID](https://spdx.org/licenses/). |
|Groups.Deps.Name|string|`{{.Name}}`|The name of the dependency. |
|Groups.Deps.Version|string|`{{.Version}}`|The version of the dependency. |
|Groups.Deps.LicenseID|string|`{{.LicenseID}}`|The [SPDX ID](https://spdx.org/licenses/) of the dependency license. |

<details>
<summary>Summary Template Generate</summary>

Summary template content:
```
{{.LicenseContent }}
{{ range .Groups }}
========================================================================
{{.LicenseID}} licenses
========================================================================
{{range .Deps}}
    {{.Name}} {{.Version}} {{.LicenseID}}
{{- end }}
{{ end }}
```

Generate LICENSE file content:
```
                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.

      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.

      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.

      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity. For the purposes of this definition,
      "control" means (i) the power, direct or indirect, to cause the
      direction or management of such entity, whether by contract or
      otherwise, or (ii) ownership of fifty percent (50%) or more of the
      outstanding shares, or (iii) beneficial ownership of such entity.

      "You" (or "Your") shall mean an individual or Legal Entity
      exercising permissions granted by this License.

      "Source" form shall mean the preferred form for making modifications,
      including but not limited to software source code, documentation
      source, and configuration files.

      "Object" form shall mean any form resulting from mechanical
      transformation or translation of a Source form, including but
      not limited to compiled object code, generated documentation,
      and conversions to other media types.

      "Work" shall mean the work of authorship, whether in Source or
      Object form, made available under the License, as indicated by a
      copyright notice that is included in or attached to the work
      (an example is provided in the Appendix below).

      "Derivative Works" shall mean any work, whether in Source or Object
      form, that is based on (or derived from) the Work and for which the
      editorial revisions, annotations, elaborations, or other modifications
      represent, as a whole, an original work of authorship. For the purposes
      of this License, Derivative Works shall not include works that remain
      separable from, or merely link (or bind by name) to the interfaces of,
      the Work and Derivative Works thereof.

      "Contribution" shall mean any work of authorship, including
      the original version of the Work and any modifications or additions
      to that Work or Derivative Works thereof, that is intentionally
      submitted to Licensor for inclusion in the Work by the copyright owner
      or by an individual or Legal Entity authorized to submit on behalf of
      the copyright owner. For the purposes of this definition, "submitted"
      means any form of electronic, verbal, or written communication sent
      to the Licensor or its representatives, including but not limited to
      communication on electronic mailing lists, source code control systems,
      and issue tracking systems that are managed by, or on behalf of, the
      Licensor for the purpose of discussing and improving the Work, but
      excluding communication that is conspicuously marked or otherwise
      designated in writing by the copyright owner as "Not a Contribution."

      "Contributor" shall mean Licensor and any individual or Legal Entity
      on behalf of whom a Contribution has been received by Licensor and
      subsequently incorporated within the Work.

   2. Grant of Copyright License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      copyright license to reproduce, prepare Derivative Works of,
      publicly display, publicly perform, sublicense, and distribute the
      Work and such Derivative Works in Source or Object form.

   3. Grant of Patent License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      (except as stated in this section) patent license to make, have made,
      use, offer to sell, sell, import, and otherwise transfer the Work,
      where such license applies only to those patent claims licensable
      by such Contributor that are necessarily infringed by their
      Contribution(s) alone or by combination of their Contribution(s)
      with the Work to which such Contribution(s) was submitted. If You
      institute patent litigation against any entity (including a
      cross-claim or counterclaim in a lawsuit) alleging that the Work
      or a Contribution incorporated within the Work constitutes direct
      or contributory patent infringement, then any patent licenses
      granted to You under this License for that Work shall terminate
      as of the date such litigation is filed.

   4. Redistribution. You may reproduce and distribute copies of the
      Work or Derivative Works thereof in any medium, with or without
      modifications, and in Source or Object form, provided that You
      meet the following conditions:

      (a) You must give any other recipients of the Work or
          Derivative Works a copy of this License; and

      (b) You must cause any modified files to carry prominent notices
          stating that You changed the files; and

      (c) You must retain, in the Source form of any Derivative Works
          that You distribute, all copyright, patent, trademark, and
          attribution notices from the Source form of the Work,
          excluding those notices that do not pertain to any part of
          the Derivative Works; and

      (d) If the Work includes a "NOTICE" text file as part of its
          distribution, then any Derivative Works that You distribute must
          include a readable copy of the attribution notices contained
          within such NOTICE file, excluding those notices that do not
          pertain to any part of the Derivative Works, in at least one
          of the following places: within a NOTICE text file distributed
          as part of the Derivative Works; within the Source form or
          documentation, if provided along with the Derivative Works; or,
          within a display generated by the Derivative Works, if and
          wherever such third-party notices normally appear. The contents
          of the NOTICE file are for informational purposes only and
          do not modify the License. You may add Your own attribution
          notices within Derivative Works that You distribute, alongside
          or as an addendum to the NOTICE text from the Work, provided
          that such additional attribution notices cannot be construed
          as modifying the License.

      You may add Your own copyright statement to Your modifications and
      may provide additional or different license terms and conditions
      for use, reproduction, or distribution of Your modifications, or
      for any such Derivative Works as a whole, provided Your use,
      reproduction, and distribution of the Work otherwise complies with
      the conditions stated in this License.

   5. Submission of Contributions. Unless You explicitly state otherwise,
      any Contribution intentionally submitted for inclusion in the Work
      by You to the Licensor shall be under the terms and conditions of
      this License, without any additional terms or conditions.
      Notwithstanding the above, nothing herein shall supersede or modify
      the terms of any separate license agreement you may have executed
      with Licensor regarding such Contributions.

   6. Trademarks. This License does not grant permission to use the trade
      names, trademarks, service marks, or product names of the Licensor,
      except as required for reasonable and customary use in describing the
      origin of the Work and reproducing the content of the NOTICE file.

   7. Disclaimer of Warranty. Unless required by applicable law or
      agreed to in writing, Licensor provides the Work (and each
      Contributor provides its Contributions) on an "AS IS" BASIS,
      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
      implied, including, without limitation, any warranties or conditions
      of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A
      PARTICULAR PURPOSE. You are solely responsible for determining the
      appropriateness of using or redistributing the Work and assume any
      risks associated with Your exercise of permissions under this License.

   8. Limitation of Liability. In no event and under no legal theory,
      whether in tort (including negligence), contract, or otherwise,
      unless required by applicable law (such as deliberate and grossly
      negligent acts) or agreed to in writing, shall any Contributor be
      liable to You for damages, including any direct, indirect, special,
      incidental, or consequential damages of any character arising as a
      result of this License or out of the use or inability to use the
      Work (including but not limited to damages for loss of goodwill,
      work stoppage, computer failure or malfunction, or any and all
      other commercial damages or losses), even if such Contributor
      has been advised of the possibility of such damages.

   9. Accepting Warranty or Additional Liability. While redistributing
      the Work or Derivative Works thereof, You may choose to offer,
      and charge a fee for, acceptance of support, warranty, indemnity,
      or other liability obligations and/or rights consistent with this
      License. However, in accepting such obligations, You may act only
      on Your own behalf and on Your sole responsibility, not on behalf
      of any other Contributor, and only if You agree to indemnify,
      defend, and hold each Contributor harmless for any liability
      incurred by, or claims asserted against, such Contributor by reason
      of your accepting any such warranty or additional liability.


========================================================================
MIT licenses
========================================================================

    github.com/BurntSushi/toml v0.3.1 MIT
    github.com/alecthomas/units v0.0.0-20151022065526-2efee857e7cf MIT
    github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e MIT
    github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da MIT
    github.com/armon/go-radix v0.0.0-20180808171621-7fddfc383310 MIT
    github.com/beorn7/perks v1.0.0 MIT
    github.com/bgentry/speakeasy v0.1.0 MIT
    github.com/bketelsen/crypt v0.0.3-0.20200106085610-5cbc8cc4026c MIT
    github.com/bmatcuk/doublestar/v2 v2.0.4 MIT

========================================================================
ISC licenses
========================================================================

    github.com/davecgh/go-spew v1.1.1 ISC

========================================================================
BSD-2-Clause licenses
========================================================================

    github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 BSD-2-Clause
    github.com/gorilla/websocket v1.4.2 BSD-2-Clause
    github.com/pkg/errors v0.8.1 BSD-2-Clause
    github.com/russross/blackfriday/v2 v2.0.1 BSD-2-Clause

========================================================================
MPL-2.0-no-copyleft-exception licenses
========================================================================

    github.com/hashicorp/consul/api v1.1.0 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/consul/sdk v0.1.1 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/go-cleanhttp v0.5.1 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/go-immutable-radix v1.0.0 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/go-multierror v1.0.0 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/go-rootcerts v1.0.0 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/go-sockaddr v1.0.0 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/go-uuid v1.0.1 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/golang-lru v0.5.1 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/logutils v1.0.0 MPL-2.0-no-copyleft-exception
    github.com/hashicorp/memberlist v0.1.3 MPL-2.0-no-copyleft-exception

========================================================================
MPL-2.0 licenses
========================================================================

    github.com/hashicorp/errwrap v1.0.0 MPL-2.0
    github.com/hashicorp/hcl v1.0.0 MPL-2.0
    github.com/hashicorp/serf v0.8.2 MPL-2.0
    github.com/mitchellh/cli v1.0.0 MPL-2.0
    github.com/mitchellh/gox v0.4.0 MPL-2.0

========================================================================
MIT and Apache licenses
========================================================================

    gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 MIT and Apache

========================================================================
Apache-2.0 licenses
========================================================================

    cloud.google.com/go v0.46.3 Apache-2.0
    cloud.google.com/go/bigquery v1.0.1 Apache-2.0
    cloud.google.com/go/datastore v1.0.0 Apache-2.0
    cloud.google.com/go/firestore v1.1.0 Apache-2.0
    cloud.google.com/go/pubsub v1.0.1 Apache-2.0
    cloud.google.com/go/storage v1.0.0 Apache-2.0

========================================================================
BSD-3-Clause licenses
========================================================================

    dmitri.shuralyov.com/gpu/mtl v0.0.0-20190408044501-666a987793e9 BSD-3-Clause
    github.com/BurntSushi/xgb v0.0.0-20160522181843-27f122750802 BSD-3-Clause
    github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc BSD-3-Clause
    github.com/fsnotify/fsnotify v1.4.7 BSD-3-Clause
```

</details>

#### Check Dependencies' licenses

This command can be used to perform automatic license compatibility check, when there is incompatible licenses found,
the command will exit with status code 1 and fail the command.

```bash
license-eye -c test/testdata/.licenserc_for_test_check.yaml dep check
```

<details>
<summary>Dependency Check Result</summary>

```
INFO GITHUB_TOKEN is not set, license-eye won't comment on the pull request
INFO Loading configuration from file: .licenserc.yaml
WARNING Failed to resolve the license of <github.com/gogo/protobuf>: cannot identify license content
WARNING Failed to resolve the license of <github.com/kr/logfmt>: cannot find license file
WARNING Failed to resolve the license of <github.com/magiconair/properties>: cannot identify license content
WARNING Failed to resolve the license of <github.com/miekg/dns>: cannot identify license content
WARNING Failed to resolve the license of <github.com/pascaldekloe/goe>: cannot identify license content
WARNING Failed to resolve the license of <github.com/russross/blackfriday/v2>: cannot identify license content
WARNING Failed to resolve the license of <gopkg.in/check.v1>: cannot identify license content
ERROR the following licenses are incompatible with the main license: Apache-2.0
License: Unknown Dependency: github.com/gogo/protobuf
License: Unknown Dependency: github.com/kr/logfmt
License: Unknown Dependency: github.com/magiconair/properties
License: Unknown Dependency: github.com/miekg/dns
License: Unknown Dependency: github.com/pascaldekloe/goe
License: Unknown Dependency: github.com/russross/blackfriday/v2
License: Unknown Dependency: gopkg.in/check.v1
exit status 1
```

</details>

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

  license-location-threshold: 80 # <9>

  language: # <10>
    Go: # <11>
      extensions: #<12>
        - ".go"
      filenames: #<13>
        - "config.go"
        - "config_test.go"
      comment_style_id: DoubleSlash # <14>

dependency: # <15>
  files: # <16>
    - go.mod
  licenses: # <17>
    - name: dependency-name # <18>
      version: dependency-version # <19>
      license: Apache-2.0 # <20>
```

1. The `header` section is configurations for source codes license header.
2. The [SPDX ID](https://spdx.org/licenses/) of the license, it’s convenient when your license is standard SPDX license, so that you can simply specify this identifier without copying the whole license `content` or `pattern`. This will be used as the content when `fix` command needs to insert a license header.
3. The copyright owner to replace the `[owner]` in the `SPDX-ID` license template.
4. If you are not using the standard license text, you can paste your license text here, this will be used as the content when `fix` command needs to insert a license header, if both `license` and `SPDX-ID` are specified, `license` wins.
5. The `pattern` is an optional regexp. You don’t need this if all the file headers are the same as `license` or the license of `SPDX-ID`, otherwise you need to compose a pattern that matches your existing license texts so that `license-eye` won't complain about the existing license headers. If you want to replace your existing license headers, you can compose a `pattern` that matches your existing license headers, and modify the `content` to what you want to have, then `license-eye header fix` would rewrite all the existing license headers to the wanted `content`.
6. The `paths` are the path list that will be checked (and fixed) by license-eye, default is `['**']`. Formats like `**/*`.md and `**/bin/**` are supported.
7. The `paths-ignore` are the path list that will be ignored by license-eye. By default, `.git` and the content in `.gitignore` will be inflated into the `paths-ignore` list.
8. On what condition License-Eye will comment the check results on the pull request, `on-failure`, `always` or `never`. Options other than `never` require the environment variable `GITHUB_TOKEN` to be set.
9. The `license-location-threshold` specifies the index threshold where the license header can be located.
10. The `language` is an optional configuration. You can set the language license header comment style. If it doesn't exist, it will use the default configuration at the `languages.yaml`. An [example](test/testdata/.licenserc_language_config_test.yaml) is to use block comment style for Go codes.
11. Specify the programming language identifier. You can set different configurations for multiple languages.
12. The `extensions` are the files with these extensions which the configuration will take effect.
13. The `filenames` are the specified files which the configuration will take effect.
14. The `comment_style_id` set the license header comment style, it's the `id` at the `styles.yaml`.
15. The `dependency` section is configurations for resolving dependencies' licenses.
16. The `files` are the files that declare the dependencies of a project, typically, `go.mod` in Go project, `pom.xml` in maven project, and `package.json` in NodeJS project. If it's a relative path, it's relative to the `.licenserc.yaml`.
17. Declare the licenses which cannot be identified by this tool.
18. The `name` of the dependency, The name is different for different projects, `PackagePath` in Go project, `GroupID:ArtifactID` in maven project, `PackageName` in NodeJS project.
19. The `version` of the dependency, it's locked, preventing license changed between different versions.
20. The [SPDX ID](https://spdx.org/licenses/) of the dependency license.

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

- [assets/styles.yaml](assets/styles.yaml)

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

- There is an [activity diagram](./docs/header_fix_logic.svg) explaining the implemented license header
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

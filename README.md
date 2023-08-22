# SkyWalking Eyes

<img src="http://skywalking.apache.org/assets/logo.svg" alt="Sky Walking logo" height="90px" align="right" />

A full-featured license tool to check and fix license headers and resolve dependencies' licenses.

[![Twitter Follow](https://img.shields.io/twitter/follow/asfskywalking.svg?style=for-the-badge&label=Follow&logo=twitter)](https://twitter.com/AsfSkyWalking)

## Usage

You can use License-Eye in GitHub Actions or in your local machine.

### GitHub Actions

First of all, add a `.licenserc.yaml` in the root of your project, for Apache Software Foundation projects, the following configuration should be enough.

> **Note**: The full configurations can be found in [the configuration section](#configurations).

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

# If you don't want to check dependencies' license compatibility, remove the following part
dependency:
  files:
    - pom.xml           # If this is a maven project.
    - Cargo.toml        # If this is a rust project.
    - package.json      # If this is a npm project.
    - go.mod            # If this is a Go project.
```

#### Check License Headers

To check license headers in GitHub Actions, add a step in your GitHub workflow.

```yaml
- name: Check License Header
  uses: apache/skywalking-eyes/header@main      # always prefer to use a revision instead of `main`.
  # with:
      # log: debug # optional: set the log level. The default value is `info`.
      # config: .licenserc.yaml # optional: set the config file. The default value is `.licenserc.yaml`.
      # token: # optional: the token that license eye uses when it needs to comment on the pull request. Set to empty ("") to disable commenting on pull request. The default value is ${{ github.token }}
      # mode: # optional: Which mode License-Eye should be run in. Choices are `check` or `fix`. The default value is `check`.
```

#### Fix License Headers

By default the action runs License-Eye in check mode, which will raise an error
if any of the processed files are missing license headers. If `mode` is set to
`fix`, the action will instead apply the license header to any processed file
that is missing a license header. The fixed files can then be pushed back to the
pull request using another GitHub action. For example:

```yaml
- name: Fix License Header
  uses: apache/skywalking-eyes/header@main
  with:
    mode: fix
- name: Apply Changes
  uses: EndBug/add-and-commit@v4
  env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  with:
    author_name: License Bot
    author_email: license_bot@github.com
    message: 'Automatic application of license header'
```

> **Warning**: The exit code of fix mode is always 0 and can not be used to block CI
status. Consider running the action in check mode if you would like CI to fail
when a file is missing a license header.

> **Note**: In 0.3.0 and earlier versions, GitHub Actions `apache/skywalking-eyes`
> only works for header check/fix, since 0.4.0, we have a dedicate GitHub Actions
> `apache/skywalking-eyes/header` for header check/fix and a GitHub Actions
> `apache/skywalking-eyes/dependency` for dependency resolve/check.
> Now `apache/skywalking-eyes` is equivalent to `apache/skywalking-eyes/header` in
> order not to break existing usages of `apache/skywalking-eyes`.

#### Check Dependencies' License

To check dependencies license in GitHub Actions, add a step in your GitHub workflow.

```yaml
- name: Check Dependencies' License
  uses: apache/skywalking-eyes/dependency@main      # always prefer to use a revision instead of `main`.
  # with:
      # log: debug # optional: set the log level. The default value is `info`.
      # config: .licenserc.yaml # optional: set the config file. The default value is `.licenserc.yaml`.
      # mode: # optional: Which mode License-Eye should be run in. Choices are `check` or `resolve`. The default value is `check`.
      # flags: # optional: Extra flags appended to the command, for example, `--summary=path/to/template.tmpl`
```

### Docker Image

```shell
docker run -it --rm -v $(pwd):/github/workspace apache/skywalking-eyes header check
docker run -it --rm -v $(pwd):/github/workspace apache/skywalking-eyes header fix
```

### Docker Image from the latest codes

For users and developers who want to help to test the latest codes on main branch, we publish a Docker image to the GitHub
Container Registry for every commit in main branch, tagged with the commit sha. If it's the latest commit in main
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

If you have the Go SDK installed, you can also use the `go install` command to install the latest code.

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

This command assists human audits of the dependencies licenses. It's exit code is always 0.

It supports two flags:

| Flag name   | Short name | Description                                                                                                                            |
|-------------|------------|----------------------------------------------------------------------------------------------------------------------------------------|
| `--output`  | `-o`       | Save the dependencies' `LICENSE` files to a specified directory so that you can put them in distribution package if needed.            |
| `--summary` | `-s`       | Based on the template, aggregate all dependency information and generate a `LICENSE` file.                                             |
| `--license` | `-l`       | The output path to the LICENSE file to be generated. The default summary format will be used if summary template file is not specified |

```bash
license-eye -c test/testdata/.licenserc_for_test_check.yaml dep resolve -o ./dependencies/licenses -s LICENSE.tpl
```

<details>
<summary>Dependency Resolve Result</summary>

```
INFO Loading configuration from file: .licenserc.yaml
WARNING Failed to resolve the license of <github.com/acomagu/bufpipe@v1.0.3>: cannot find license file
Dependency                           |              License |                                  Version
------------------------------------ | -------------------- | ----------------------------------------
github.com/Masterminds/goutils       |           Apache-2.0 |                                   v1.1.1
github.com/Masterminds/semver/v3     |                  MIT |                                   v3.1.1
github.com/Masterminds/sprig/v3      |                  MIT |                                   v3.2.2
github.com/Microsoft/go-winio        |                  MIT |                                   v0.5.2
github.com/ProtonMail/go-crypto      |         BSD-3-Clause |       v0.0.0-20220824120805-4b6e5c587895
github.com/bmatcuk/doublestar/v2     |                  MIT |                                   v2.0.4
github.com/cloudflare/circl          |         BSD-3-Clause |                                   v1.2.0
github.com/davecgh/go-spew           |                  ISC |                                   v1.1.1
github.com/emirpasic/gods            | BSD-2-Clause and ISC |                                  v1.18.1
github.com/go-git/gcfg               |         BSD-3-Clause |                                   v1.5.0
github.com/go-git/go-billy/v5        |           Apache-2.0 |                                   v5.3.1
github.com/go-git/go-git/v5          |           Apache-2.0 |                                   v5.4.2
github.com/golang/protobuf           |         BSD-3-Clause |                                   v1.5.2
github.com/google/go-github/v33      |         BSD-3-Clause |                                  v33.0.0
github.com/google/go-querystring     |         BSD-3-Clause |                                   v1.1.0
github.com/google/licensecheck       |         BSD-3-Clause |                                   v0.3.1
github.com/google/uuid               |         BSD-3-Clause |                                   v1.1.1
github.com/huandu/xstrings           |                  MIT |                                   v1.3.1
github.com/imdario/mergo             |         BSD-3-Clause |                                  v0.3.13
github.com/inconshreveable/mousetrap |           Apache-2.0 |                                   v1.0.0
github.com/jbenet/go-context         |                  MIT |       v0.0.0-20150711004518-d14ea06fba99
github.com/kevinburke/ssh_config     |                  MIT |                                   v1.2.0
github.com/mitchellh/copystructure   |                  MIT |                                   v1.0.0
github.com/mitchellh/go-homedir      |                  MIT |                                   v1.1.0
github.com/mitchellh/reflectwalk     |                  MIT |                                   v1.0.0
github.com/pmezard/go-difflib        |         BSD-3-Clause |                                   v1.0.0
github.com/sergi/go-diff             |                  MIT |                                   v1.2.0
github.com/shopspring/decimal        |                  MIT |                                   v1.2.0
github.com/sirupsen/logrus           |                  MIT |                                   v1.8.1
github.com/spf13/cast                |                  MIT |                                   v1.3.1
github.com/spf13/cobra               |           Apache-2.0 |                                   v1.4.0
github.com/spf13/pflag               |         BSD-3-Clause |                                   v1.0.5
github.com/stretchr/testify          |                  MIT |                                   v1.7.0
github.com/xanzy/ssh-agent           |           Apache-2.0 |                                   v0.3.2
golang.org/x/crypto                  |         BSD-3-Clause |       v0.0.0-20220829220503-c86fa9a7ed90
golang.org/x/mod                     |         BSD-3-Clause | v0.6.0-dev.0.20220106191415-9b9b3d81d5e3
golang.org/x/net                     |         BSD-3-Clause |       v0.0.0-20220826154423-83b083e8dc8b
golang.org/x/oauth2                  |         BSD-3-Clause |       v0.0.0-20220411215720-9780585627b5
golang.org/x/sys                     |         BSD-3-Clause |       v0.0.0-20220829200755-d48e67d00261
golang.org/x/text                    |         BSD-3-Clause |                                   v0.3.7
golang.org/x/tools                   |         BSD-3-Clause |                                  v0.1.10
golang.org/x/xerrors                 |         BSD-3-Clause |       v0.0.0-20220517211312-f3a8303e98df
google.golang.org/appengine          |           Apache-2.0 |                                   v1.6.7
google.golang.org/protobuf           |         BSD-3-Clause |                                  v1.28.0
gopkg.in/warnings.v0                 |         BSD-2-Clause |                                   v0.1.2
gopkg.in/yaml.v3                     |   MIT and Apache-2.0 |                                   v3.0.0
github.com/acomagu/bufpipe           |              Unknown |                                   v1.0.3

ERROR failed to identify the licenses of following packages (1):
github.com/acomagu/bufpipe
exit status 1
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
    copyright-year: '1993-2022' # <25>
    software-name: skywalking-eyes # <4>
    content: | # <5>
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

    pattern: | # <6>
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

  paths: # <7>
    - '**'

  paths-ignore: # <8>
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

  comment: on-failure # <9>

  license-location-threshold: 80 # <10>

  language: # <11>
    Go: # <12>
      extensions: #<13>
        - ".go"
      filenames: #<14>
        - "config.go"
        - "config_test.go"
      comment_style_id: DoubleSlash # <15>

dependency: # <16>
  files: # <17>
    - go.mod
  licenses: # <18>
    - name: dependency-name # <19>
      version: dependency-version # <20>
      license: Apache-2.0 # <21>
  threshold: 75 # <22>
  excludes: # <23>
    - name: dependency-name # the same format as <19>
      version: dependency-version # the same format as <20>
      recursive: true # whether to exclude all transitive dependencies brought by <dependency-name>, now only maven project supports this <24>
```

1. The `header` section is configurations for source codes license header. If you have multiple modules or packages in your project that have differing licenses, this section may contain a list of licenses:
```yaml
header:
  - path: "/path/to/module/a"
    license:
      spdx-id: Apache-2.0
  - path: "/path/to/module/b"
    license:
      spdx-id: MPL-2.0
```
2. The [SPDX ID](https://spdx.org/licenses/) of the license, it’s convenient when your license is standard SPDX license, so that you can simply specify this identifier without copying the whole license `content` or `pattern`. This will be used as the content when `fix` command needs to insert a license header.
3. The copyright owner to replace the `[owner]` in the `SPDX-ID` license template.
4. The copyright software name to replace the `[software-name]` in the `SPDX-ID` license template.
5. If you are not using the standard license text, you can paste your license text here, this will be used as the content when `fix` command needs to insert a license header, if both `license` and `SPDX-ID` are specified, `license` wins.
6. The `pattern` is an optional regexp. You don’t need this if all the file headers are the same as `license` or the license of `SPDX-ID`, otherwise you need to compose a pattern that matches your existing license texts so that `license-eye` won't complain about the existing license headers. If you want to replace your existing license headers, you can compose a `pattern` that matches your existing license headers, and modify the `content` to what you want to have, then `license-eye header fix` would rewrite all the existing license headers to the wanted `content`.
7. The `paths` are the path list that will be checked (and fixed) by license-eye, default is `['**']`. Formats like `**/*`.md and `**/bin/**` are supported.
8. The `paths-ignore` are the path list that will be ignored by license-eye. By default, `.git` and the content in `.gitignore` will be inflated into the `paths-ignore` list.
9. On what condition License-Eye will comment the check results on the pull request, `on-failure`, `always` or `never`. Options other than `never` require the environment variable `GITHUB_TOKEN` to be set.
10. The `license-location-threshold` specifies the index threshold where the license header can be located.
11. The `language` is an optional configuration. You can set the language license header comment style. If it doesn't exist, it will use the default configuration at the `languages.yaml`. An [example](test/testdata/.licenserc_language_config_test.yaml) is to use block comment style for Go codes.
12. Specify the programming language identifier. You can set different configurations for multiple languages.
13. The `extensions` are the files with these extensions which the configuration will take effect.
14. The `filenames` are the specified files which the configuration will take effect.
15. The `comment_style_id` set the license header comment style, it's the `id` at the `styles.yaml`.
16. The `dependency` section is configurations for resolving dependencies' licenses.
17. The `files` are the files that declare the dependencies of a project, typically, `go.mod` in Go project, `pom.xml` in maven project, and `package.json` in NodeJS project. If it's a relative path, it's relative to the `.licenserc.yaml`.
18. Declare the licenses which cannot be identified by this tool.
19. The `name` of the dependency, The name is different for different projects, `PackagePath` in Go project, `GroupID:ArtifactID` in maven project, `PackageName` in NodeJS project. You can use file pattern as described in [the doc](https://pkg.go.dev/path/filepath#Match).
20. The `version` of the dependency, comma seperated string (such as `1.0,2.0,3.0`), if this is empty, it means all versions of the dependency.
21. The [SPDX ID](https://spdx.org/licenses/) of the dependency license.
22. The minimum percentage of the file that must contain license text for identifying a license, default is `75`.
23. The dependencies that should be excluded when analyzing the licenses, this is useful when you declare the dependencies in `pom.xml` with `compile` scope but don't distribute them in package. (Note that non-`compile` scope dependencies are automatically excluded so you don't need to put them here).
24. The transitive dependencies brought by <23> should be recursively excluded when analyzing the licenses, currently only maven project supports this.
25. The copyright year of the work, if it's empty, it will be set to the current year. If you don't want to update the license year anually, you can set this to the year of the first publication of your work, such as `1994`, or `1994-2023`.

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

## Papers

- CCF-A [AUGER: Automatically Generating Review Comments with Pre-training Models](https://2022.esec-fse.org/details/fse-2022-research-papers/21/AUGER-Automatically-Generating-Review-Comments-with-Pre-training-Models)
- CCF-C [DeepRelease: Language-agnostic Release Notes Generation from Pull Requests of Open-source Software](https://www.computer.org/csdl/proceedings-article/apsec/2021/378400a101/1B4mbM3STVm)

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

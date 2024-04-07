## 0.6.0

* Add instructions to fix header issues in markdown comment.
* Add Eclipse Foundation specific Apache 2.0 license header.
* Add support for OPA policy files, protobuf.
* Add weak-compatible check to dependency check.

## 0.5.0

* feat(header templates): add support for AGPL-3.0 by @elijaholmos in https://github.com/apache/skywalking-eyes/pull/125
* Upgrade go version to 1.18 by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/126
* Add MulanPSL-2.0 support. by @jmjoy in https://github.com/apache/skywalking-eyes/pull/127
* New Header Template: GPL-3.0-or-later by @ddlees in https://github.com/apache/skywalking-eyes/pull/128
* Update README.md by @rovast in https://github.com/apache/skywalking-eyes/pull/129
* Add more `.env.[mode]` support for VueJS project by @rovast in https://github.com/apache/skywalking-eyes/pull/130
* Docker Multiple Architecture Support :fixes#9089 by @mohammedtabish0 in https://github.com/apache/skywalking-eyes/pull/132
* Polish maven test for convenient debug by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/134
* feat: list files by git when possible by @tisonkun in https://github.com/apache/skywalking-eyes/pull/133
* Switch to npm ci for reliable builds by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/135
* Fix optional dependencies are not excluded by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/136
* Fix exclude not work for transitive dependencies and add recursive config by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/137
* Add some tests for maven resovler by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/138
* feat(header-fix): add Svelte support by @elijaholmos in https://github.com/apache/skywalking-eyes/pull/139
* dep: do not write license files if they already exist by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/140
* fix: not ignore *.txt to make sure files like CMakeLists.txt can be checked by @acelyc111 in https://github.com/apache/skywalking-eyes/pull/141
* fix license header normalizer by @xiaoyawei in https://github.com/apache/skywalking-eyes/pull/142
* Substitute variables in license content for header command by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/143
* Correct indent in Apache-2.0 template by @tisonkun in https://github.com/apache/skywalking-eyes/pull/144
* Add copyright-year configuration by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/145
* dep/maven: use output file to store the dep tree for cleaner result by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/146
* dep/maven: resolve dependencies before analysis by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/147
* gha: switch to composite running mode and set up cache by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/149
* Fix GitHub Actions wrong path by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/151
* Normalize license for cargo. by @jmjoy in https://github.com/apache/skywalking-eyes/pull/153
* Remove space characters in license for cargo. by @jmjoy in https://github.com/apache/skywalking-eyes/pull/154
* Bump up dependencies to fix CVE by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/155
* Bump up GHA to depress warnings by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/156
* Leverage the built-in cache in setup-go@v4 by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/157
* Dependencies check should report unknown licneses by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/158
* Fix wrong indentation in doc by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/159
* Add EPL-2.0 header template by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/160
* Dependency resolve with default template and specified output of license by @crholm in https://github.com/apache/skywalking-eyes/pull/163
* Bump up go git to support .gitconfig user path by @kezhenxu94 in https://github.com/apache/skywalking-eyes/pull/164

## 0.4.0
- Reorganize GHA by header and dependency. (#123)
- Add rust cargo support for dep command. (#121)
- Support license expression in dep check. (#120)
- Prune npm packages before listing all dependencies (#119)
- Add support for multiple licenses in the header config section (#118)
- Add `excludes` to `license resolve` config (#117)
- maven: set `group:artifact` as dependency name and extend functions in summary template (#116)
- Stablize summary context to perform consistant output (#115)
- Add custom license urls for identification (#114)
- Lazy initialize GitHub client for comment (#111)
- Make license identifying threshold configurable (#110)
- Use Google's licensecheck to identify licenses (#107)
- dep: short circuit if user declare dep license (#108)

## 0.3.0

- Dependency License
  - Fix license check in go library testify (#93)

- License Header
  - `fix` command supports more languages:
    - Add comment style for cmake language (#86)
    - Add comment style for hcl (#89)
    - Add mpl-2.0 header template (#87)
    - Support fix license header for tcl files (#102)
    - Add python docstring comment style (#100)
    - Add comment style for makefile & editorconfig (#90)
  - Support config license header comment style (#97)
  - Trim leading and trailing newlines before rewrite license header cotent (#94)
  - Replace already existing license header based on pattern (#98)
  - [docs] add the usage for config the license header comment style (#99)

- Project
  - Obtain default github token in github actions (#82)
  - Add tests for bare spdx license header content (#92)
  - Add github action step summary for better experience (#104)
  - Adds an option to the action to run in `fix` mode (#84)
  - Provide `--summary` flag to generate the license summary file (#103)
  - Add .exe suffix to windows binary (#101)
  - Fix wrong file path and exclude binary files in src release (#81)
  - Use t.tempdir to create temporary test directory (#95)
  - Config: fix incorrect log message (#91)
  - [docs] correct spelling mistakes (#96)

## 0.2.0

- Dependency License
  - Support resolving go.mod for Go
  - Support resolving pom.xml for maven (#50)
  - Support resolving jars' licenses (#53)
  - Support resolving npm dependencies' licenses (#48)
  - Support saving dependencies' licenses (#69)
  - Add `dependency check` to check dependencies license compatibilities (#58)

- License Header
  - `fix` command supports more languages:
    - Add support for plantuml (#42)
    - Add support for PHP (#40)
    - Add support for Twig template language (#39)
    - Add support for Smarty template language (#38)
    - Add support for MatLab files (#37)
    - Add support for TypeScript language files (#73)
    - Add support for nextflow files (#65)
    - Add support for perl files (#63)
    - Add support for ini extension (#24)
    - Add support for R files (#64)
    - Add support for .rst files and allow fixing header of a single file (#25)
    - Add support for Rust files (#29)
    - Add support for bat files (#32)
  - Remove .tsx from XML language extensions
  - Honor Python's coding directive (#68)
  - Fix file extension conflict between RenderScript and Rust (#66)
  - Add comment type to cython declaration (#62)
  - header fix: respect user configured license content (#60)
  - Expose `license-location-threshold` as config item (#34)
  - Fix infinite recursive calls when containing symbolic files (#33)
  - defect: avoid crash when no comment style is found (#23)

- Project
  - Enhance license identification (#79)
  - Support installing via go install (#76)
  - Speed up the initialization phase (#75)
  - Resolve absolute path in `.gitignore` to relative path (#67)
  - Reduce img size and add npm env (#59)
  - Make the config file and log level in GitHub Action configurable (#56, #57)
  - doc: add a PlantUML activity diagram of header fixing mechanism (#41)
  - Fix bug: license file is not found but reported message is nil (#49)
  - Add all well-known licenses and polish normalizers (#47)
  - Fix compatibility issues in Windows (#44)
  - feature: add reasonable default config to allow running in a new repo without copying config file (#28)
  - chore: only build linux binary when building inside docker (#26)
  - chore: upgrade to go 1.16 and remove `go-bindata` (#22)
  - Add documentation about how to use via docker image (#20)

## 0.1.0

- License Header
  + Add `check` and `fix` command.
  + `check` results can be reported to pull request as comments.
  + `fix` suggestions can be filed on pull request as edit suggestions.

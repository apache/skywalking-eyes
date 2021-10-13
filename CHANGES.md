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

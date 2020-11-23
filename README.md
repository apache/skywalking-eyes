# license-checker

A CLI tool for checking license headers, which theoretically supports checking all types of files.

## Install

```bash 
git clone 
cd license-checker
make
```

## Usage

```bash
Usage: license-checker [flags]

license-checker walks the specified path recursively and checks 
if the specified files have the license header in the config file.

Usage:
  license-checker [flags]

Flags:
  -c, --config string   the config file (default ".licenserc.json")
  -h, --help            help for license-checker
  -l, --loose           loose mode
  -p, --path string     the path to check (default ".")
  -v, --verbose         verbose mode
```

## Test

```bash
bin/license-checker -p test -c test/.licenserc_for_test.json 
```

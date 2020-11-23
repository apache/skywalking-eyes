# license-checker

A CLI tool for checking license headers, which theoretically supports checking all types of files.

## Install

```bash 
git clone https://github.com/fgksgf/license-checker.git
cd license-checker
make
```

## Usage

```
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

## Configuration

```json
{
  "licenseStrict": [
    "Licensed to the Apache Software Foundation (ASF) under one or more",
    "contributor license agreements.  See the NOTICE file distributed with",
    "..."
  ],
  "licenseLoose": [
    "Apache License, Version 2.0"
  ],
  "targetFiles": [
    "java",
    "go",
    "py",
    "sh",
    "graphql",
    "yaml",
    "yml"
  ],
  "exclude": {
    "files": [
      ".gitignore",
      "NOTICE",
      "go.mod",
      "go.sum",
      ".DS_Store",
      "LICENSE"
    ],
    "extensions": [
      "md",
      "xml",
      "json"
    ],
    "directories": [
      "bin",
      ".github",
      ".git",
      ".idea",
      "test"
    ]
  }
}
```

## Test

```bash
bin/license-checker -p test -c test/.licenserc_for_test.json 
```

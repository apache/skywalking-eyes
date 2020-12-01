# Copyright © 2020 Hoshea Jiang <hoshea@apache.org>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PROJECT = license-checker
VERSION ?= latest
OUT_DIR = bin
CURDIR := $(shell pwd)
FILES := $$(find .$$($(PACKAGE_DIRECTORIES)) -name "*.go")
FAIL_ON_STDOUT := awk '{ print } END { if (NR > 0) { exit 1 } }'

GO := GO111MODULE=on go
GO_PATH = $(shell $(GO) env GOPATH)
GO_BUILD = $(GO) build
GO_GET = $(GO) get
GO_TEST = $(GO) test
GO_LINT = $(GO_PATH)/bin/golangci-lint
GO_LICENSER = $(GO_PATH)/bin/go-licenser

all: clean deps lint test build

tools:
	mkdir -p $(GO_PATH)/bin
	#$(GO_LICENSER) -version || GO111MODULE=off $(GO_GET) -u github.com/elastic/go-licenser

deps: tools
	$(GO_GET) -v -t -d ./...

.PHONY: lint
lint: tools
	$(GO_LINT) version || curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_PATH)/bin v1.21.0
	@gofmt -s -l -w $(FILES) 2>&1 | $(FAIL_ON_STDOUT)
	$(GO_LINT) run -v ./...

.PHONE: test
test: clean lint
	$(GO_TEST) ./...
	@>&2 echo "Great, all tests passed."

.PHONY: build
build: deps
	$(GO_BUILD) -o $(OUT_DIR)/$(PROJECT)

#.PHONY: license
#license: clean tools
#	$(GO_LICENSER) -d -license='ASL2' .

.PHONY: fix
fix: tools
	$(GO_LINT) run -v --fix ./...

.PHONY: clean
clean: tools
	-rm -rf bin

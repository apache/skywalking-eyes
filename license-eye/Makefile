# Licensed to Apache Software Foundation (ASF) under one or more contributor
# license agreements. See the NOTICE file distributed with
# this work for additional information regarding copyright
# ownership. Apache Software Foundation (ASF) licenses this file to you under
# the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

PROJECT = license-eye
VERSION ?= latest
OUT_DIR = bin
ARCH := $(shell uname)
OSNAME := $(if $(findstring Darwin,$(ARCH)),darwin,linux)

GO := GO111MODULE=on go
GO_PATH = $(shell $(GO) env GOPATH)
GO_BUILD = $(GO) build
GO_TEST = $(GO) test
GO_LINT = $(GO_PATH)/bin/golangci-lint
GO_BUILD_LDFLAGS = -X github.com/apache/skywalking-eyes/license-eye/commands.version=$(VERSION)

GO_BINDATA = $(GO_PATH)/bin/go-bindata

PLATFORMS := windows linux darwin
os = $(word 1, $@)
ARCH = amd64

RELEASE_BIN = skywalking-license-eye-$(VERSION)-bin
RELEASE_SRC = skywalking-license-eye-$(VERSION)-src

all: clean lint license test build

.PHONY: lint
lint: codegen
	$(GO_LINT) version || curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_PATH)/bin
	$(GO_LINT) run -v ./...

.PHONY: fix-lint
fix-lint:
	$(GO_LINT) run -v --fix ./...

.PHONY: license
license: clean codegen
	$(GO) run cmd/license-eye/main.go header check -c ../.licenserc.yaml

.PHONY: codegen
codegen: clean
	$(GO_BINDATA) -v || curl -Lo $(GO_BINDATA) https://github.com/kevinburke/go-bindata/releases/download/v3.11.0/go-bindata-$(OSNAME)-amd64 && chmod +x $(GO_BINDATA)
	$(GO_BINDATA) --nocompress --nometadata --pkg assets --ignore '.*\.go' -o "assets/assets.gen.go" assets/...

.PHONY: test
test: clean codegen
	$(GO_TEST) ./... -coverprofile=coverage.txt -covermode=atomic
	@>&2 echo "Great, all tests passed."

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p $(OUT_DIR)
	GOOS=$(os) GOARCH=$(ARCH) $(GO_BUILD) $(GO_BUILD_FLAGS) -ldflags "$(GO_BUILD_LDFLAGS)" -o $(OUT_DIR)/$(os)/$(PROJECT) cmd/license-eye/main.go

.PHONY: build
build: codegen windows linux darwin

.PHONY: docker
docker:
	docker build . -t license-eye:latest

.PHONY: clean
clean:
	-rm -rf bin
	-rm -rf assets/*.gen.go
	-rm -rf coverage.txt
	-rm -rf "$(RELEASE_BIN)"*
	-rm -rf "$(RELEASE_SRC)"*

.PHONY: verify
verify: clean license lint test

release-src: clean
	-mkdir $(RELEASE_SRC)
	-cp ../NOTICE $(RELEASE_SRC)
	-rsync -av . $(RELEASE_SRC) --exclude $(RELEASE_SRC) --exclude .DS_Store
	-tar -zcvf $(RELEASE_SRC).tgz $(RELEASE_SRC)
	-rm -rf "$(RELEASE_SRC)"

release-bin: build
	-mkdir $(RELEASE_BIN)
	-cp -R bin $(RELEASE_BIN)
	-cp -R dist/* $(RELEASE_BIN)
	-cp -R CHANGES.md $(RELEASE_BIN)
	-cp -R README.adoc $(RELEASE_BIN)
	-cp -R ../NOTICE $(RELEASE_BIN)
	-tar -zcvf $(RELEASE_BIN).tgz $(RELEASE_BIN)
	-rm -rf "$(RELEASE_BIN)"

release: verify release-src release-bin
	gpg --batch --yes --armor --detach-sig $(RELEASE_SRC).tgz
	shasum -a 512 $(RELEASE_SRC).tgz > $(RELEASE_SRC).tgz.sha512
	gpg --batch --yes --armor --detach-sig $(RELEASE_BIN).tgz
	shasum -a 512 $(RELEASE_BIN).tgz > $(RELEASE_BIN).tgz.sha512

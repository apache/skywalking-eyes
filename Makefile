# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

HUB ?= docker.io/apache
PROJECT ?= license-eye
VERSION ?= latest
OUT_DIR = bin
ARCH := $(shell uname)
OSNAME := $(if $(findstring Darwin,$(ARCH)),darwin,linux)

GO := GO111MODULE=on go
GO_PATH = $(shell $(GO) env GOPATH)
GO_BUILD = $(GO) build
GO_TEST = $(GO) test
GO_LINT = $(GO_PATH)/bin/golangci-lint
GO_BUILD_LDFLAGS = -X github.com/apache/skywalking-eyes/commands.version=$(VERSION)

PLANTUML_VERSION = 1.2021.9

PLATFORMS := windows linux darwin
os = $(word 1, $@)
ARCH = amd64

RELEASE_BIN = skywalking-$(PROJECT)-$(VERSION)-bin
RELEASE_SRC = skywalking-$(PROJECT)-$(VERSION)-src

all: clean lint license test build

$(GO_LINT):
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_PATH)/bin v1.46.2

.PHONY: lint
lint: $(GO_LINT)
	$(GO_LINT) run -v ./...

.PHONY: fix-lint
fix-lint: $(GO_LINT)
	$(GO_LINT) run -v --fix ./...

.PHONY: license
license: clean
	$(GO) run cmd/$(PROJECT)/main.go header check

.PHONY: test
test: clean
	$(GO_TEST) ./... -coverprofile=coverage.txt -covermode=atomic
	@>&2 echo "Great, all tests passed."

windows: PROJECT_SUFFIX=.exe

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p $(OUT_DIR)
	GOOS=$(os) GOARCH=$(ARCH) $(GO_BUILD) $(GO_BUILD_FLAGS) -ldflags "$(GO_BUILD_LDFLAGS)" -o $(OUT_DIR)/$(os)/$(PROJECT)$(PROJECT_SUFFIX) cmd/$(PROJECT)/main.go

.PHONY: build
build: windows linux darwin

.PHONY: docker
docker:
	docker build . -t $(HUB)/$(PROJECT):$(VERSION) -t $(HUB)/$(PROJECT):latest

.PHONY: docker-push
docker-push:
	docker buildx create --use --driver docker-container --name skywalking_eyes_main
	docker buildx build --push --platform linux/amd64,linux/arm64 -t $(HUB)/$(PROJECT):$(VERSION) -t $(HUB)/$(PROJECT):latest .
	docker buildx rm skywalking_eyes_main

.PHONY: docker-release
docker-release: docker docker-push

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
	-tar -zcvf $(RELEASE_SRC)/$(RELEASE_SRC).tgz \
	--exclude $(RELEASE_SRC).tgz \
	--exclude bin \
	--exclude .git \
	--exclude .idea \
	--exclude .DS_Store \
	--exclude .github \
	--exclude $(RELEASE_SRC) \
	--exclude query-protocol/schema.graphqls \
	--exclude *.jar \
	.
	mv $(RELEASE_SRC)/$(RELEASE_SRC).tgz $(RELEASE_SRC).tgz
	-rm -rf "$(RELEASE_SRC)"

release-bin: build
	-mkdir $(RELEASE_BIN)
	-cp -R bin $(RELEASE_BIN)
	-cp -R dist/* $(RELEASE_BIN)
	-cp -R CHANGES.md $(RELEASE_BIN)
	-cp -R README.md $(RELEASE_BIN)
	-cp -R NOTICE $(RELEASE_BIN)
	-tar -zcvf $(RELEASE_BIN).tgz $(RELEASE_BIN)
	-rm -rf "$(RELEASE_BIN)"

release: verify release-src release-bin
	gpg --batch --yes --armor --detach-sig $(RELEASE_SRC).tgz
	shasum -a 512 $(RELEASE_SRC).tgz > $(RELEASE_SRC).tgz.sha512
	gpg --batch --yes --armor --detach-sig $(RELEASE_BIN).tgz
	shasum -a 512 $(RELEASE_BIN).tgz > $(RELEASE_BIN).tgz.sha512

.PHONY: docs-gen
docs-gen:
	-if [ ! -f "plantuml.jar" ]; then curl -sL -o plantuml.jar https://repo1.maven.org/maven2/net/sourceforge/plantuml/plantuml/$(PLANTUML_VERSION)/plantuml-$(PLANTUML_VERSION).jar; fi;
	-java -jar plantuml.jar -tsvg -nometadata "docs/*.plantuml"

.PHONY: verify-docs
verify-docs: docs-gen
	@if [ ! -z "`git status -s docs`" ]; then \
		echo "Following diagram files are not consistent with CI:"; \
		git status -s docs; \
		git diff --color --word-diff --exit-code docs; \
		exit 1; \
	fi

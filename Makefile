.DEFAULT_GOAL := build

export NAME ?= $(shell basename $(shell pwd))
export VERSION := v$(shell cat VERSION)
export OUT_PATH ?= out
export CGO_ENABLED ?= 0

local := $(PWD)/.local
localBin := $(local)/bin

export PATH := $(localBin):$(PATH)

export CHECKSUMS_FILE_NAME := release.sha256
export CHECKSUMS_FILE := $(OUT_PATH)/$(CHECKSUMS_FILE_NAME)

REVISION := $(shell git rev-parse --short=8 HEAD || echo unknown)
REFERENCE := $(shell git show-ref | grep "$(REVISION)" | grep -v HEAD | awk '{print $$2}' | sed 's|refs/remotes/origin/||' | sed 's|refs/heads/||' | sort | head -n 1)
BUILT := $(shell date -u +%Y-%m-%dT%H:%M:%S%z)
PKG = $(shell go list .)

OS_ARCHS ?= darwin/arm64 \
            linux/amd64
GO_LDFLAGS ?= -X $(PKG).NAME=$(NAME) -X $(PKG).VERSION=$(VERSION) \
              -X $(PKG).REVISION=$(REVISION) -X $(PKG).BUILT=$(BUILT) \
              -X $(PKG).REFERENCE=$(REFERENCE) \
              -w -extldflags '-static'

PROTOC := $(localBin)/protoc
PROTOC_VERSION := 22.2

PROTOC_GEN_GO := protoc-gen-go
PROTOC_GEN_GO_VERSION := v1.29.1

PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc
PROTOC_GEN_GO_GRPC_VERSION := v1.3.0

MOCKERY := mockery
MOCKERY_VERSION := 2.16.0

build:
	@mkdir -p $(OUT_PATH)
	go build -a -ldflags "$(GO_LDFLAGS)" -o $(OUT_PATH)/$(NAME) ./cmd/$(NAME)

.PHONY: .mods
.mods:
	go mod download

TARGETS = $(foreach OSARCH,$(OS_ARCHS),${OUT_PATH}/$(NAME)-$(subst /,-,$(OSARCH)))

$(TARGETS): .mods
	@mkdir -p $(OUT_PATH)
	GOOS=$(firstword $(subst -, ,$(subst $(OUT_PATH)/$(NAME)-,,$@))) \
			 GOARCH=$(lastword $(subst .exe,,$(subst -, ,$(subst $(OUT_PATH)/$(NAME)-,,$@)))) \
			 go build -a -ldflags "$(GO_LDFLAGS)" -o $@ ./cmd/$(NAME)

MAKEFLAGS += -j$(shell nproc)
all:$(TARGETS)

.PHONY: test
test: .mods
	go test ./...

.PHONY: test-race
test-race: export CGO_ENABLED=1
test-race: .mods
	go test -race ./...

.PHONY: shellcheck
shellcheck:
	shellcheck $(shell find ci -name "*.sh")

.PHONY: clean
clean:
	rm -fr $(OUT_PATH)

.PHONY: upload-release
upload-release: sign-checksums-file
upload-release:
	ci/upload-release.sh

.PHONY: sign-checksums-file
sign-checksums-file: generate-checksums-file
	ci/sign-checksums-file.sh

.PHONY: generate-checksums-file
generate-checksums-file:
	ci/generate-checksums-file.sh

.PHONY: release
release:
	ci/release.sh

.PHONY: do-release
do-release:
	git tag -s $(VERSION) -m "Version $(VERSION)"
	git push origin $(VERSION)

.PHONY: codegen
codegen: dependencies
	go generate ./...

.PHONY: dependencies
dependencies: $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) $(MOCKERY)

$(PROTOC): OS_TYPE ?= $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/osx/')
$(PROTOC): ARCH_TYPE ?= $(shell uname -m | sed 's/arm64/aarch_64/')
$(PROTOC): DOWNLOAD_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(OS_TYPE)-$(ARCH).zip
$(PROTOC):
	# Installing $(DOWNLOAD_URL) as $(PROTOC)
	@mkdir -p "$(localBin)"
	@curl -sL "$(DOWNLOAD_URL)" -o "$(local)/protoc.zip"
	@unzip "$(local)/protoc.zip" -d "$(local)/"
	@chmod +x "$(PROTOC)"
	@rm "$(local)/protoc.zip"

.PHONY: $(PROTOC_GEN_GO)
$(PROTOC_GEN_GO):
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)

.PHONY: $(PROTOC_GEN_GO_GRPC)
$(PROTOC_GEN_GO_GRPC):
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

.PHONY: $(MOCKERY)
$(MOCKERY):
	go install github.com/vektra/mockery/v2@v$(MOCKERY_VERSION)

#!/usr/bin/make -f

COMMIT := $(shell git log -1 --format='%H')
VERSION ?= $(shell git describe --tags --always)

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
TM_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build

# Dependencies version
DEPS_COSMOS_SDK_VERSION := $(shell cat go.sum | grep 'github.com/cosmos/cosmos-sdk' | grep -v -e 'go.mod' | tail -n 1 | awk '{ print $$2; }')
DEPS_ETHERMINT_VERSION := $(shell cat go.sum | grep 'github.com/dymensionxyz/ethermint' | grep -v -e 'go.mod' | tail -n 1 | awk '{ print $$2; }')
DEPS_OSMOSIS_VERSION := $(shell cat go.sum | grep 'github.com/dymensionxyz/osmosis' | grep -v -e 'go.mod' | tail -n 1 | awk '{ print $$2; }')
DEPS_IBC_GO_VERSION := $(shell cat go.sum | grep 'github.com/cosmos/ibc-go' | grep -v -e 'go.mod' | tail -n 1 | awk '{ print $$2; }')
DEPS_COSMOS_PROTO_VERSION := $(shell cat go.sum | grep 'github.com/cosmos/cosmos-proto' | grep -v -e 'go.mod' | tail -n 1 | awk '{ print $$2; }')
DEPS_COSMOS_GOGOPROTO_VERSION := $(shell cat go.sum | grep 'github.com/cosmos/gogoproto' | grep -v -e 'go.mod' | tail -n 1 | awk '{ print $$2; }')
DEPS_CONFIO_ICS23_VERSION := go/$(shell cat go.sum | grep 'github.com/confio/ics23/go' | grep -v -e 'go.mod' | tail -n 1 | awk '{ print $$2; }')

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(DYMENSION_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace := $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=dymension \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=dymd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
	      -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(TM_VERSION)

ifeq (cleveldb,$(findstring cleveldb,$(DYMENSION_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ifeq (,$(findstring nostrip,$(DYMENSION_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(DYMENSION_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

all: install

.PHONY: install
install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/dymd

.PHONY: build build-debug

build: go.sum
	go build $(BUILD_FLAGS) -o $(BUILDDIR)/dymd ./cmd/dymd

build-debug: go.sum
	$(eval temp_ldflags := $(filter-out -w -s,$(ldflags)))
	go build -tags "$(build_tags)" -ldflags '$(temp_ldflags)' -gcflags "all=-N -l" -o $(BUILDDIR)/dymd ./cmd/dymd

docker-build-e2e:
	@DOCKER_BUILDKIT=1 docker build -t ghcr.io/dymensionxyz/dymension:e2e -f Dockerfile .

docker-build-e2e-debug:
	@DOCKER_BUILDKIT=1 CGO_ENABLED=0 docker build -t ghcr.io/dymensionxyz/dymension:e2e-debug -f Dockerfile.debug .

docker-run-debug:
	@DOCKER_BUILDKIT=1 docker-compose -f docker-compose.debug.yml up

###############################################################################
###                                Releasing                                ###
###############################################################################

PACKAGE_NAME:=github.com/dymensionxyz/dymension
GOLANG_CROSS_VERSION  = v1.22
GOPATH ?= '$(HOME)/go'
release-dry-run:
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v ${GOPATH}/pkg:/go/pkg \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--clean --skip=validate --skip=publish --snapshot

release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean --skip=validate

.PHONY: release-dry-run release

###############################################################################
###                                Proto                                    ###
###############################################################################

# ------
# NOTE: Link to the tendermintdev/sdk-proto-gen docker images:
#       https://hub.docker.com/r/tendermintdev/sdk-proto-gen/tags
#
protoVer=v0.7
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(protoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(protoVer)
# ------
# NOTE: cosmos/proto-builder image is needed because clang-format is not installed
#       on the tendermintdev/sdk-proto-gen docker image.
#		Link to the cosmos/proto-builder docker images:
#       https://github.com/cosmos/cosmos-sdk/pkgs/container/proto-builder
#
protoCosmosVer=0.11.2
protoCosmosName=ghcr.io/cosmos/proto-builder:$(protoCosmosVer)
protoCosmosImage=$(DOCKER) run --network host --rm -v $(CURDIR):/workspace --workdir /workspace $(protoCosmosName)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi
	@go mod tidy

proto-swagger-gen:
	@echo "Downloading Protobuf dependencies"
	@make proto-download-deps
	@echo "Generating Protobuf Swagger"
	$(protoCosmosImage) sh ./scripts/protoc-swagger-gen.sh

SWAGGER_DIR=./swagger-proto
THIRD_PARTY_DIR=$(SWAGGER_DIR)/third_party

proto-download-deps:
	mkdir -p "$(THIRD_PARTY_DIR)/cosmos_tmp" && \
	cd "$(THIRD_PARTY_DIR)/cosmos_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/cosmos-sdk.git" && \
	git config core.sparseCheckout true && \
	printf "proto\nthird_party\n" > .git/info/sparse-checkout && \
	git fetch --depth=1 origin "$(DEPS_COSMOS_SDK_VERSION)" && \
	git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/cosmos_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/ethermint_tmp" && \
	cd "$(THIRD_PARTY_DIR)/ethermint_tmp" && \
	git init && \
	git remote add origin "https://github.com/dymensionxyz/ethermint.git" && \
	git config core.sparseCheckout true && \
	printf "proto\nthird_party\n" > .git/info/sparse-checkout && \
	git fetch --depth=1 origin "$(DEPS_ETHERMINT_VERSION)" && \
	git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/ethermint_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/osmosis_tmp" && \
	cd "$(THIRD_PARTY_DIR)/osmosis_tmp" && \
	git init && \
	git remote add origin "https://github.com/dymensionxyz/osmosis.git" && \
	git config core.sparseCheckout true && \
	printf "proto\nthird_party\n" > .git/info/sparse-checkout && \
	git fetch --depth=1 origin "$(DEPS_OSMOSIS_VERSION)" && \
	git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/osmosis_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/ibc_tmp" && \
	cd "$(THIRD_PARTY_DIR)/ibc_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/ibc-go.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git fetch --depth=1 origin "$(DEPS_IBC_GO_VERSION)" && \
	git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/ibc_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/cosmos_proto_tmp" && \
	cd "$(THIRD_PARTY_DIR)/cosmos_proto_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/cosmos-proto.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git fetch --depth=1 origin "$(DEPS_COSMOS_PROTO_VERSION)" && \
	git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/cosmos_proto_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/gogoproto" && \
	curl -SSL https://raw.githubusercontent.com/cosmos/gogoproto/$(DEPS_COSMOS_GOGOPROTO_VERSION)/gogoproto/gogo.proto > "$(THIRD_PARTY_DIR)/gogoproto/gogo.proto"

	mkdir -p "$(THIRD_PARTY_DIR)/google/api" && \
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > "$(THIRD_PARTY_DIR)/google/api/annotations.proto"
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > "$(THIRD_PARTY_DIR)/google/api/http.proto"

	mkdir -p "$(THIRD_PARTY_DIR)/confio/ics23" && \
	curl -sSL https://raw.githubusercontent.com/confio/ics23/$(DEPS_CONFIO_ICS23_VERSION)/proofs.proto > "$(THIRD_PARTY_DIR)/proofs.proto"

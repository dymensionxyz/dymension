#!/usr/bin/make -f

COMMIT := $(shell git log -1 --format='%H')
VERSION ?= $(shell git describe --tags --always)

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
TM_VERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::')
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build

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
	      -X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TM_VERSION)

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

.PHONY: build
build: go.sum
	go build $(BUILD_FLAGS) -o build/dymd ./cmd/dymd

build-debug: go.sum
	go build $(BUILD_FLAGS) -gcflags="all=-N -l" -o build/dymd ./cmd/dymd

docker-build-e2e:
	@DOCKER_BUILDKIT=1 docker build -t ghcr.io/dymensionxyz/dymension:e2e -f Dockerfile .

docker-build-e2e-debug:
	@DOCKER_BUILDKIT=1 docker build -t ghcr.io/dymensionxyz/dymension:e2e-debug -f Dockerfile.debug .

docker-run-debug:
	@DOCKER_BUILDKIT=1 docker run -d -p 36657:36657 -p 36656:36656 -p 8090:8090 -p 8091:8091 -p 1318:1318 -p 9545:9545 -p 9546:9546 -p 4000:4000 --name dymension-debug ghcr.io/dymensionxyz/dymension:e2e-debug

###############################################################################
###                                Proto                                    ###
###############################################################################

protoVer=v0.7
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(protoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(protoVer)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi
	@go mod tidy


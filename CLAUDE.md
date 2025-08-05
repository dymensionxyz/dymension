# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the Dymension Hub repository (`d-dymension`), which serves as the settlement layer of the Dymension protocol. It's a Cosmos SDK-based blockchain written in Go.

## Common Development Commands

### Building and Installation
```bash
# Install dymd binary to $GOPATH/bin
make install

# Build dymd binary to ./build/dymd
make build

# Build debug version (with debug symbols, no optimization)
make build-debug

# Generate protobuf files (requires Docker)
make proto-gen

# Generate swagger documentation
make proto-swagger-gen

# Format proto files
make proto-format

# Lint proto files
make proto-lint
```

### Local Development Setup
```bash
# Setup a local node with default configuration
./scripts/setup_local.sh

# Start local node
dymd start

# Bootstrap liquidity pools
sh scripts/pools/pools_bootstrap.sh

# Setup incentives streams
sh scripts/incentives/fund_incentives.sh
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with race detection and coverage
go test -race -coverprofile=coverage.txt ./...

# Run specific module tests
go test ./x/rollapp/...

# Run integration tests (in ibctesting directory)
go test ./ibctesting/...
```

### Linting and Code Formatting
```bash
# Run golangci-lint (configured in .golangci.yml)
golangci-lint run

# Linters enabled: errcheck, gocyclo, gosec, govet, ineffassign, misspell, revive, staticcheck, unconvert, unused, errorlint
# Uses golangci-lint v2.1 (as per CI)

# Format Go code with gofumpt
gofumpt -w .

# Format proto files
make proto-format
```

## Architecture Overview

### Core Modules (`x/`)

#### Settlement Layer Modules
- **rollapp**: Manages RollApp registration, state updates, and lifecycle
  - Handles state info submissions from sequencers
  - Manages finalization queue and fraud proofs
  - Supports genesis bridges for RollApp initialization
  
- **sequencer**: Manages sequencer registration, bonding, and rotation
  - Handles sequencer lifecycle (bonding, unbonding)
  - Manages proposer selection and rotation
  - Tracks sequencer rewards and slashing

- **lightclient**: Manages canonical IBC light clients for RollApps
  - Ensures safe IBC client creation and operation
  - Validates light client parameters against expected values
  - Critical for EIBC functionality

#### IBC and Bridging Modules
- **delayedack**: Handles delayed acknowledgments for RollApp packets
  - Implements custom IBC middleware for RollApp-specific packet handling
  - Manages packet finalization based on RollApp state updates
  - Handles fraud detection and rollback

- **eibc**: Event-driven IBC for fast finality
  - Allows market makers to fulfill orders before finalization
  - Manages demand orders and liquidity providers
  - Implements custom authorization for order fulfillment

- **bridgingfee**: Manages fees for bridging between Hub and RollApps

- **forward**: Handles token forwarding between Hyperlane and IBC/EIBC
  - Supports cross-protocol token transfers
  - Manages routing between different bridge types

- **kas**: Bookkeeping for Kaspa<->Dymension Hyperlane bridge

#### Economic Modules
- **iro**: Initial RollApp Offering
  - Manages token launches for new RollApps
  - Implements bonding curve mechanics
  - Handles vesting and liquidity provisioning

- **incentives**: Liquidity and staking incentives
  - Manages gauge-based reward distribution
  - Supports asset-specific and RollApp-specific gauges
  - Integrates with sponsorship for community-driven distribution

- **sponsorship**: Community-driven incentive distribution
  - Allows validators to vote on reward distribution
  - Updates streamer distributions based on community votes

- **streamer**: Time-based token distribution
  - Manages scheduled token releases
  - Supports multiple distribution records per stream

- **lockup**: Token locking mechanism
  - Required for earning incentive rewards
  - Manages lock periods and unlocking

#### Utility Modules
- **denommetadata**: Token denomination metadata management
  - Handles IBC denom metadata registration
  - Supports Hyperlane token metadata

- **dymns**: Dymension Name Service
  - Manages .dym domain registration and trading
  - Supports aliases and reverse resolution
  - Implements marketplace for name trading

- **gamm**: Generalized Automated Market Maker (from Osmosis)
  - Provides liquidity pool functionality
  - Required for incentives distribution

### Key Architectural Patterns

1. **Module Structure**: Each module follows standard Cosmos SDK patterns:
   - `keeper/`: Core business logic and state management
   - `types/`: Protobuf types, messages, and interfaces
   - `client/cli/`: CLI commands for the module
   - Tests alongside implementation files

2. **IBC Integration**: Heavy use of IBC middleware pattern for custom packet handling

3. **Hook System**: Extensive use of hooks for cross-module communication

4. **Testing Patterns**:
   - Unit tests: Standard Go testing with table-driven tests
   - Keeper tests: Often use test suites with setup/teardown
   - Integration tests: Located in `ibctesting/` using IBC testing framework

## Module Dependencies

Key inter-module dependencies:
- `rollapp` → `sequencer`: For sequencer validation
- `delayedack` → `rollapp`: For state finalization checks
- `eibc` → `delayedack`: For packet fulfillment
- `incentives` → `lockup`, `sponsorship`: For reward distribution
- `lightclient` → `rollapp`, `sequencer`: For canonical client validation

## GitHub Actions and CI/CD

### Workflows
The repository uses GitHub Actions for continuous integration. Key workflows include:

1. **Build and Test** (`test.yml`): Runs on all pushes and PRs
   - Builds the project
   - Runs tests with race detection and coverage
   - Uploads coverage to Codecov

2. **Linting** (`golangci_lint.yml`): Runs golangci-lint on all code
   - Uses golangci-lint v2.1
   - Configuration in `.golangci.yml`

3. **Protocol Buffers** (`proto.yaml`): Validates protobuf files
   - Runs proto-format and proto-gen
   - Checks for uncommitted changes
   - Breaking change detection with buf

4. **E2E Tests** (`e2e_test.yml`, `e2e_test_upgrade.yml`): Integration tests
   - Tests various scenarios including upgrades
   - Nightly runs for comprehensive testing

### Verifying CI Checks Locally

Before pushing changes, verify CI checks locally:

```bash
# 1. Run tests with race detection and coverage (mimics CI)
go install github.com/ory/go-acc@v0.2.6
go-acc -o coverage.txt ./... -- -v --race

# 2. Run golangci-lint (ensure version matches CI)
golangci-lint run

# 3. Format Go code with gofumpt
gofumpt -w .

# 4. For proto changes, run format and generation
make proto-format
make proto-gen

# 5. Check for any uncommitted changes
git status --porcelain
```

### Required Tools
- `golangci-lint` v2.1 (install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1`)
- `gofumpt` (install: `go install mvdan.cc/gofumpt@latest`)
- `go-acc` for coverage (install: `go install github.com/ory/go-acc@v0.2.6`)
- Docker (for proto generation)

## Environment Configuration

Default ports (can be configured via environment variables):
- RPC: 36657 (`SETTLEMENT_ADDR`)
- P2P: 36656 (`P2P_ADDRESS`)
- gRPC: 8090 (`GRPC_ADDRESS`)
- gRPC Web: 8091 (`GRPC_WEB_ADDRESS`)
- REST API: 1318 (`API_ADDRESS`)
- JSON-RPC: 9545 (`JSONRPC_ADDRESS`)
- WebSocket: 9546 (`JSONRPC_WS_ADDRESS`)

## Important Notes

- Go version: 1.23.0 (with toolchain 1.24.2)
- Uses custom forks of cosmos-sdk, cometbft, ibc-go, and osmosis
- Proto generation requires Docker with specific proto-builder images
- Default chain ID for local development: `dymension_100-1`
- Default denom: `udym` (1 dym = 1,000,000 udym)
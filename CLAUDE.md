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

#### Core Modules
- **rollapp**: RollApp registration, state updates, finalization queue, fraud proofs
  - Handles state info submissions from sequencers
  - Supports genesis bridges for RollApp initialization
  
- **sequencer**: Sequencer lifecycle (bonding/unbonding), rotation, rewards
  - Manages proposer selection and slashing
  
- **lightclient**: Canonical IBC light clients for RollApps (critical for EIBC)
  - Validates light client parameters against expected values
  
- **delayedack**: Custom IBC middleware for delayed packet acknowledgments
  - Manages packet finalization based on RollApp state updates
  - Handles fraud detection and rollback
  
- **eibc**: Fast finality via market makers fulfilling orders before finalization
  - Manages demand orders and liquidity providers
  
- **bridgingfee**: Fee management for Hub<->RollApp bridging
- **forward**: Token routing between Hyperlane and IBC/EIBC
- **kas**: Kaspa<->Dymension Hyperlane bridge bookkeeping

#### Economic Modules
- **iro**: Initial RollApp Offering with bonding curves
  - Handles vesting and liquidity provisioning
  
- **incentives**: Gauge-based liquidity rewards
  - Supports asset-specific and RollApp-specific gauges
  
- **sponsorship**: Validator-voted reward distribution
- **streamer**: Scheduled token releases
- **lockup**: Token locking for incentive rewards

#### Utility Modules
- **denommetadata**: IBC and Hyperlane token metadata
- **dymns**: .dym domain name service and marketplace
  - Supports aliases and reverse resolution
  
- **gamm**: AMM pools (from Osmosis)

### Key Architectural Patterns

- **Module Structure**: Standard Cosmos SDK with `keeper/`, `types/`, `client/cli/`
- **IBC Integration**: Custom middleware for RollApp packet handling
- **Hook System**: Cross-module communication via hooks
- **Testing**: Unit tests, keeper test suites, integration tests in `ibctesting/`

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

```bash
# Tests with coverage (mimics CI)
go install github.com/ory/go-acc@v0.2.6
go-acc -o coverage.txt ./... -- -v --race

# Linting and formatting
golangci-lint run      # v2.1 - install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1
gofumpt -w .          # install: go install mvdan.cc/gofumpt@latest

# Proto changes only
make proto-format
make proto-gen

# Check uncommitted changes
git status --porcelain
```

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

## Code Style Guidelines

### Comments
- **NEVER** comment the WHAT of code (the code itself shows what it does)
- **ONLY** add comments to explain WHY for unusual, unclear, or non-obvious implementations
- Examples:
  ```go
  // BAD: Increment counter by 1
  counter++
  
  // GOOD: We need to retry 3 times due to intermittent network issues with the external API
  for i := 0; i < 3; i++ {
      // ...
  }
  ```

### CI Tool Usage
When verifying code locally, run tools **only as needed** to avoid unnecessary work:
- `golangci-lint run` - Only if you modified Go code
- `gofumpt -w .` - Only if you modified Go code  
- `make proto-format` and `make proto-gen` - Only if you modified .proto files
- `go test` - Only for packages you modified
- Full test suite with `go-acc` - Only before final push or if requested

Remember: These tools can take significant time to run, especially proto generation and full test suites.
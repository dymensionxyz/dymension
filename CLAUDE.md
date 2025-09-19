# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the Dymension Hub repository (`d-dymension`), which serves as the settlement layer of the Dymension protocol. It's a Cosmos SDK-based blockchain written in Go.

## FACTS

- Cosmos SDK code all runs in a single thread, there is no multithreading
- Transactions rollback if there is any error, in this sense they are atomic: all or nothing. If they panic, nothing happens.
- BeginBlocker and Endblocker functions in module should never panic, it would cause a chain halt
- All cosmos validators must get the same result so the code should be deterministic

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
- Default denom: `adym` (1 dym = 1 * 10^18 adym)

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

## dymd CLI User Guide

The `dymd` command-line interface is the primary tool for interacting with the Dymension blockchain. It provides commands for node operations, key management, querying blockchain state, and broadcasting transactions.

### Core Concepts

- **Home Directory**: Default `~/.dymension/` - stores config, data, and keys
- **Chain ID**: Network identifier (e.g., `dymension_1100-1` for mainnet)
- **Node Connection**: Default `tcp://localhost:26657`, configurable via `--node`
- **Output Formats**: JSON (`-o json`) or text (`-o text`)
- **Key Backend**: Storage for private keys (test, file, os, kwallet, pass)

### Configuration

#### View and Modify Client Config
```bash
# View current client configuration
dymd config view client

# Set specific config values
dymd config set client chain-id dymension_1100-1
dymd config set client node https://dymension-rpc.polkachu.com:443
dymd config set client keyring-backend test
dymd config set client output json

# View differences from defaults
dymd config diff

# Get specific config value
dymd config get client.chain-id
```

### Node Management

#### Initialize a New Node
```bash
# Initialize node with moniker
dymd init my-node --chain-id dymension_1100-1

# Initialize with specific denom
dymd init my-node --chain-id dymension_1100-1 --default-denom adym

# Recover from existing mnemonic
dymd init my-node --recover
```

#### Start the Node
```bash
# Start full node
dymd start

# Start with custom home directory
dymd start --home /custom/path

# Start with specific ports
dymd start --p2p.laddr tcp://0.0.0.0:26656 --rpc.laddr tcp://0.0.0.0:26657
```

#### Node Information
```bash
# Get node status
dymd status

# Get node ID
dymd comet show-node-id

# Get validator consensus address
dymd comet show-address

# Get validator info
dymd comet show-validator

# Get CometBFT version
dymd comet version
```

### Key Management

#### Create and Manage Keys
```bash
# Add a new key
dymd keys add wallet1

# Recover key from mnemonic
dymd keys add wallet2 --recover

# Add key with specific HD path
dymd keys add wallet3 --hd-path "m/44'/60'/0'/0/0"

# List all keys
dymd keys list

# Show specific key details
dymd keys show wallet1
dymd keys show wallet1 --address  # Show address only
dymd keys show wallet1 --pubkey   # Show public key

# Delete a key
dymd keys delete wallet1

# Rename a key
dymd keys rename wallet1 main-wallet

# Export private key (encrypted)
dymd keys export wallet1

# Import private key
dymd keys import wallet2 wallet2.armor

# Parse address formats
dymd keys parse dym1abc...
```

#### Advanced Key Operations
```bash
# Generate mnemonic
dymd keys mnemonic

# Import Ethereum key (UNSAFE - for testing only)
dymd keys unsafe-import-eth-key eth-wallet 0xPRIVATE_KEY

# Export Ethereum key (UNSAFE)
dymd keys unsafe-export-eth-key eth-wallet
```

### Querying Blockchain State

#### Bank Module - Token Balances and Supply
```bash
# Query account balance
dymd query bank balances dym1address...

# Query specific denom balance
dymd query bank balance dym1address... adym

# Query total supply
dymd query bank total

# Query supply of specific denom
dymd query bank total adym

# Query denomination metadata
dymd query bank denom-metadata
```

#### Staking Module - Validators and Delegations
```bash
# List all validators
dymd query staking validators

# Query specific validator
dymd query staking validator dymvaloper1...

# Query delegations from an address
dymd query staking delegations dym1address...

# Query delegation to specific validator
dymd query staking delegation dym1address... dymvaloper1...

# Query unbonding delegations
dymd query staking unbonding-delegations dym1address...

# Query rewards
dymd query distribution rewards dym1address...

# Query commission
dymd query distribution commission dymvaloper1...
```

#### RollApp Module - RollApp Management
```bash
# List all RollApps
dymd query rollapp list

# Query specific RollApp
dymd query rollapp show rollapp_1234-1

# Query RollApp state
dymd query rollapp state rollapp_1234-1 --index 5

# Query latest state
dymd query rollapp state rollapp_1234-1

# Query latest height
dymd query rollapp latest-height rollapp_1234-1

# Query latest state index
dymd query rollapp latest-state-index rollapp_1234-1

# Query registered denoms
dymd query rollapp registered-denoms rollapp_1234-1
```

#### Sequencer Module - Sequencer Information
```bash
# List all sequencers
dymd query sequencer list-sequencer

# Query sequencer by address
dymd query sequencer show-sequencer dym1address...

# Query sequencers by RollApp
dymd query sequencer show-sequencers-by-rollapp rollapp_1234-1

# Query current proposer
dymd query sequencer proposer rollapp_1234-1

# Query next proposer
dymd query sequencer next-proposer rollapp_1234-1

# List all proposers
dymd query sequencer list-proposer
```

#### EIBC Module - Fast Finality Orders
```bash
# List demand orders by status
dymd query eibc list-demand-orders pending
dymd query eibc list-demand-orders finalized
dymd query eibc list-demand-orders reverted

# Query liquidity providers
dymd query eibc lps-demand

# Query LPs by address
dymd query eibc lps-demand-addr dym1address...
```

#### IRO Module - Initial RollApp Offerings
```bash
# List all IRO plans
dymd query iro plans

# Query specific IRO plan
dymd query iro plan 1

# Query IRO plan by RollApp
dymd query iro plan-by-rollapp rollapp_1234-1

# Query current price
dymd query iro price 1

# Query cost for buying/selling
dymd query iro cost 1 buy 1000000adym
dymd query iro cost 1 sell 1000000adym

# Query claimed amount
dymd query iro claimed 1 dym1address...
```

#### Governance Module
```bash
# List all proposals
dymd query gov proposals

# Query specific proposal
dymd query gov proposal 1

# Query votes on proposal
dymd query gov votes 1

# Query specific vote
dymd query gov vote 1 dym1address...

# Query deposit
dymd query gov deposit 1 dym1address...

# Query tally
dymd query gov tally 1
```

### Creating and Broadcasting Transactions

#### Common Transaction Flags
```bash
--from wallet1              # Signer's key name
--chain-id dymension_1100-1 # Network chain ID
--fees 20000adym           # Transaction fee
--gas auto                  # Auto-calculate gas
--gas-adjustment 1.5        # Gas multiplier for auto
--broadcast-mode sync       # sync, async, or block
--node https://rpc.url     # RPC endpoint
--dry-run                   # Simulate without broadcasting
--generate-only            # Create unsigned transaction
-y                         # Skip confirmation prompt
```

#### Bank Module - Token Transfers
```bash
# Send tokens
dymd tx bank send wallet1 dym1recipient... 1000000adym --from wallet1 --fees 20000adym

# Multi-send to multiple recipients
dymd tx bank multi-send wallet1 dym1addr1... 100adym dym1addr2... 200adym --from wallet1
```

#### Staking Module - Validator Operations
```bash
# Delegate to validator
dymd tx staking delegate dymvaloper1... 1000000adym --from wallet1

# Redelegate between validators
dymd tx staking redelegate dymvaloper1... dymvaloper2... 1000000adym --from wallet1

# Unbond from validator
dymd tx staking unbond dymvaloper1... 1000000adym --from wallet1

# Withdraw rewards
dymd tx distribution withdraw-rewards dymvaloper1... --from wallet1

# Withdraw all rewards
dymd tx distribution withdraw-all-rewards --from wallet1

# Withdraw validator commission
dymd tx distribution withdraw-validator-commission dymvaloper1... --from validator-key
```

#### Creating a Validator
```bash
# Create validator
dymd tx staking create-validator \
  --amount 1000000000adym \
  --pubkey $(dymd comet show-validator) \
  --moniker "My Validator" \
  --identity "keybase-id" \
  --website "https://example.com" \
  --details "Validator description" \
  --security-contact "security@example.com" \
  --commission-rate 0.10 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1 \
  --from validator-key
```

#### RollApp Module - RollApp Management
```bash
# Create new RollApp
dymd tx rollapp create-rollapp \
  rollapp_1234-1 \
  "eip155:1" \
  --description "My RollApp" \
  --logo "https://example.com/logo.png" \
  --website "https://example.com" \
  --from creator-wallet

# Update RollApp
dymd tx rollapp update-rollapp \
  rollapp_1234-1 \
  --description "Updated description" \
  --website "https://newsite.com" \
  --from owner-wallet

# Transfer ownership
dymd tx rollapp transfer-ownership rollapp_1234-1 dym1newowner... --from current-owner

# Add app
dymd tx rollapp add-app rollapp_1234-1 "App Name" "Description" "https://app.url" --from owner

# Update app
dymd tx rollapp update-app rollapp_1234-1 1 --name "New Name" --from owner

# Remove app
dymd tx rollapp remove-app rollapp_1234-1 1 --from owner
```

#### Sequencer Module - Sequencer Operations
```bash
# Create sequencer
dymd tx sequencer create-sequencer \
  dym1pubkey... \
  rollapp_1234-1 \
  "Sequencer Description" \
  --bond 1000000000adym \
  --from sequencer-wallet

# Update sequencer
dymd tx sequencer update-sequencer \
  --reward-address dym1reward... \
  --from sequencer-wallet

# Increase bond
dymd tx sequencer increase-bond 1000000000adym --from sequencer-wallet

# Decrease bond
dymd tx sequencer decrease-bond 500000000adym --from sequencer-wallet

# Unbond sequencer
dymd tx sequencer unbond --from sequencer-wallet

# Opt in/out as proposer
dymd tx sequencer opt-in --from sequencer-wallet
dymd tx sequencer opt-in --opt-out --from sequencer-wallet

# Update whitelisted relayers
dymd tx sequencer update-whitelisted-relayers \
  dym1relayer1...,dym1relayer2... \
  --from sequencer-wallet

# Kick current proposer (requires being next proposer)
dymd tx sequencer kick --from next-proposer-wallet
```

#### EIBC Module - Fast Finality Operations
```bash
# Fulfill EIBC order
dymd tx eibc fulfill-order \
  order-id \
  --fee 100000adym \
  --from fulfiller-wallet

# Create demand liquidity provider
dymd tx eibc create-demand-lp \
  rollapp_1234-1 \
  1000000000adym \
  --fee-rate 0.01 \
  --from lp-wallet

# Delete demand LP
dymd tx eibc delete-demand-lp lp-id --from lp-wallet

# Grant fulfillment authorization
dymd tx eibc grant dym1grantee... 1000000000adym --from granter-wallet
```

#### IRO Module - Initial RollApp Offering
```bash
# Buy IRO tokens
dymd tx iro buy plan-id 1000000adym --from buyer-wallet

# Sell IRO tokens
dymd tx iro sell plan-id 1000000tokens --from seller-wallet

# Claim vested tokens
dymd tx iro claim plan-id --from claimer-wallet
```

#### Governance Module - Proposals and Voting
```bash
# Submit text proposal
dymd tx gov submit-proposal \
  --title "Proposal Title" \
  --description "Detailed description" \
  --type text \
  --deposit 10000000adym \
  --from proposer-wallet

# Submit parameter change proposal
dymd tx gov submit-proposal param-change proposal.json --from proposer-wallet

# Deposit on proposal
dymd tx gov deposit 1 10000000adym --from depositor-wallet

# Vote on proposal
dymd tx gov vote 1 yes --from voter-wallet
dymd tx gov vote 1 no --from voter-wallet
dymd tx gov vote 1 abstain --from voter-wallet
dymd tx gov vote 1 no_with_veto --from voter-wallet
```

#### IBC Operations
```bash
# IBC transfer
dymd tx ibc-transfer transfer \
  transfer \
  channel-0 \
  cosmos1recipient... \
  1000000adym \
  --from sender-wallet

# Query IBC channels
dymd query ibc channel channels

# Query specific channel
dymd query ibc channel channel transfer channel-0

# Query channel client state
dymd query ibc channel client-state transfer channel-0
```

### Advanced Operations

#### Transaction Management
```bash
# Sign transaction offline
dymd tx bank send wallet1 dym1recipient... 1000000adym \
  --generate-only > unsigned.json
dymd tx sign unsigned.json --from wallet1 > signed.json

# Broadcast signed transaction
dymd tx broadcast signed.json

# Multisig operations
dymd tx multisign unsigned.json wallet1 sig1.json wallet2 sig2.json > multisig.json

# Simulate transaction
dymd tx simulate signed.json

# Decode transaction
dymd tx decode <base64-encoded-tx>

# Query transaction
dymd query tx <tx-hash>

# Wait for transaction
dymd query wait-tx <tx-hash>
```

#### Genesis Operations
```bash
# Add genesis account
dymd genesis add-genesis-account dym1address... 1000000000adym

# Generate genesis transaction
dymd genesis gentx validator-key 1000000000adym \
  --chain-id dymension_1100-1 \
  --moniker "My Validator"

# Collect genesis transactions
dymd genesis collect-gentxs

# Validate genesis file
dymd genesis validate

# Migrate genesis to new version
dymd genesis migrate v2 genesis.json
```

#### Snapshot Management
```bash
# List local snapshots
dymd snapshots list

# Export state to snapshot
dymd snapshots export

# Restore from snapshot
dymd snapshots restore <height> <format>

# Delete snapshot
dymd snapshots delete <height> <format>

# Dump snapshot to archive
dymd snapshots dump <height> <format> snapshot.tar.gz

# Load snapshot from archive
dymd snapshots load snapshot.tar.gz
```

#### State Management
```bash
# Export state to JSON
dymd export > state.json
dymd export --height 12345 > state-at-height.json
dymd export --for-zero-height > genesis-state.json

# Rollback state by one height
dymd rollback

# Prune old state
dymd prune

# Reset node state (UNSAFE - deletes all data)
dymd comet unsafe-reset-all
```

#### Debug Tools
```bash
# Convert address formats
dymd debug addr dym1abc...

# Decode public key
dymd debug pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"..."}'

# Decode raw public key
dymd debug pubkey-raw <hex/base64/bech32>

# Convert raw bytes to hex
dymd debug raw-bytes [10 21 13 255]

# List bech32 prefixes
dymd debug prefixes
```

### Network Endpoints

#### Mainnet (dymension_1100-1)
```bash
# Configure for mainnet
dymd config set client chain-id dymension_1100-1
dymd config set client node https://dymension-rpc.polkachu.com:443

# Alternative RPC endpoints
# https://rpc.dymension.nodestake.org
# https://dymension-rpc.publicnode.com:443
# https://rpc-dymension.imperator.co:443
```

#### Testnet/Playground (dymension_3405-1)
```bash
# Configure for testnet
dymd config set client chain-id dymension_3405-1
dymd config set client node https://rpc-dymension-playground35.mzonder.com:443
```

### Common Patterns and Tips

#### Using Different Key Backends
```bash
# Test backend (insecure, no password)
dymd keys add wallet --keyring-backend test

# File backend (encrypted, password required)
dymd keys add wallet --keyring-backend file

# OS backend (uses system keychain)
dymd keys add wallet --keyring-backend os
```

#### Batch Operations
```bash
# Query multiple accounts
for addr in addr1 addr2 addr3; do
  dymd query bank balances $addr
done

# Withdraw all rewards
dymd query distribution rewards <delegator> -o json | \
  jq -r '.rewards[].validator_address' | \
  xargs -I {} dymd tx distribution withdraw-rewards {} --from wallet -y
```

#### Working with JSON Output
```bash
# Parse balance with jq
dymd query bank balances dym1... -o json | jq '.balances[] | select(.denom=="adym")'

# Get all validator addresses
dymd query staking validators -o json | jq -r '.validators[].operator_address'

# Calculate total delegated
dymd query staking delegations dym1... -o json | \
  jq '[.delegation_responses[].balance.amount | tonumber] | add'
```

#### Environment Variables
```bash
# Set common environment variables
export DYMD_CHAIN_ID="dymension_1100-1"
export DYMD_NODE="https://dymension-rpc.polkachu.com:443"
export DYMD_FROM="wallet1"
export DYMD_KEYRING_BACKEND="test"

# Use in commands (automatically picked up)
dymd query bank balances dym1...
dymd tx bank send $DYMD_FROM dym1recipient... 1000000adym
```

### Troubleshooting

#### Common Issues and Solutions

1. **"account sequence mismatch"**
   - Query current sequence: `dymd query auth account dym1... | grep sequence`
   - Use correct sequence: `--sequence <number>`

2. **"insufficient fees"**
   - Increase fees: `--fees 50000adym`
   - Use gas auto: `--gas auto --gas-adjustment 1.5`

3. **"key not found"**
   - Check keyring backend: `--keyring-backend test`
   - List available keys: `dymd keys list`

4. **"connection refused"**
   - Check node is running: `dymd status`
   - Verify node URL: `--node https://correct-rpc-url`

5. **"unauthorized"**
   - Ensure correct chain-id: `--chain-id dymension_1100-1`
   - Verify account has funds: `dymd query bank balances <address>`

### Module Parameters

Query module parameters to understand system configuration:

```bash
# Query all module parameters
dymd query rollapp params
dymd query sequencer params
dymd query eibc params
dymd query iro params
dymd query incentives params
dymd query sponsorship params
dymd query staking params
dymd query distribution params
dymd query gov params
dymd query mint params
dymd query slashing params
```

### Security Best Practices

1. **Never share private keys or mnemonics**
2. **Use hardware wallets for mainnet operations**
3. **Always verify transaction details before signing**
4. **Use `--dry-run` to simulate transactions first**
5. **Keep your node software updated**
6. **Use secure RPC endpoints (HTTPS)**
7. **Backup keys securely (encrypted)**
8. **Use multisig for high-value accounts**
9. **Set appropriate gas limits to avoid overpaying**
10. **Verify validator identity before delegating**
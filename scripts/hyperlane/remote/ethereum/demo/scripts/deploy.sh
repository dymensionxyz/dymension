
# You will need:
# - Foundry (Forge + Anvil)
# - Hyperlane CLI
# - Etherenal block explorer

# Overview:
# 1. Create local Anvil ethereum instance
# 2. Configure and deploy Hyperlane contracts using the CLI


# STEP 1: Local Ethereum Instance

curl -L https://foundry.paradigm.xyz | bash
foundryup
# v 1.0.0

forge init demo

# TODO: also set chain id?
# Note: chain can be relaunched from genesis by just restarting the process
anvil --config-out /Users/danwt/Documents/dym/d-dymension/scripts/hyperlane/remote/ethereum/demo/scripts/temp/anvil.config.json --block-time 1

# Get priv key from Anvil (and note the corresponding address)
PRIV_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

# Demo
forge create --rpc-url http://127.0.0.1:8545 \
            --private-key $PRIV_KEY \
            --broadcast \
            src/MyContract.sol:MyContract

# Step 2: Deploy Hyperlane Core
# https://docs.hyperlane.xyz/docs/deploy-hyperlane

# v 9.1.0
npm install -g @hyperlane-xyz/cli

# Create hyperlane chain configs  (in ~/.hyperlane)
# follow the prompts. Chain ID from Anvil (31337). Choose something sensible as chain name and display name, for example: aaadym
hyperlane registry init
# TODO: explorer skippable?

export HYP_KEY=$PRIV_KEY
# creates a config file in cwd
# Choose testISM, then defaultHook = merkle, and requiredHook = protocolFee (with zeros)
hyperlane core init

# choose chain name from earlier from testnet list
# should succeed
# TODO: explorer for confirmation?
# Note: addresses saved in ~/.hyperlane
hyperlane core deploy 

# Step 2.5: Run Explorer
# 1. Register for Ethernal https://app.tryethernal.com/overview
# 2. Click the browser sync spinner and select 'Not Hardhat / Something else'
# 3. It will give you a cmd like ETHERNAL_API_TOKEN=xxx ethernal listen
# Note: you don't need to do anything if restarting Anvil instance


# Step 3: Deploy Hyperlane Warp Route
# https://docs.hyperlane.xyz/docs/guides/deploy-warp-route

# generate a warp route config file
hyperlane warp init

# Must adjust the config file:
    aaadym:
    type: synthetic
    mailbox: "0x8A791620dd6260079BF849Dc5567aDC3F2FdC318" # mailbox address route
    name: "SynthDym"
    symbol: "adym"
    totalSupply: 1000
    decimals: 1

# output:
# addressOrDenom: "0xc5a5C42992dECbae36851359345FE25997F5C42d"
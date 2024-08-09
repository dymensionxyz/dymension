#!/bin/sh

# Common commands
genesis_config_cmds="/app/scripts/genesis_config_commands.sh"
. "$genesis_config_cmds"

# Set parameters
DATA_DIRECTORY="$HOME/.dymension"
CONFIG_DIRECTORY="$DATA_DIRECTORY/config"
TENDERMINT_CONFIG_FILE="$CONFIG_DIRECTORY/config.toml"
CLIENT_CONFIG_FILE="$CONFIG_DIRECTORY/client.toml"
APP_CONFIG_FILE="$CONFIG_DIRECTORY/app.toml"
GENESIS_FILE="$CONFIG_DIRECTORY/genesis.json"
CHAIN_ID=${CHAIN_ID:-"dymension_100-1"}
MONIKER_NAME=${MONIKER_NAME:-"local"}
KEY_NAME=${KEY_NAME:-"local-user"}

# Setting non-default ports to avoid port conflicts when running local rollapp
SETTLEMENT_ADDR=${SETTLEMENT_ADDR:-"0.0.0.0:36657"}
P2P_ADDRESS=${P2P_ADDRESS:-"0.0.0.0:36656"}
GRPC_ADDRESS=${GRPC_ADDRESS:-"0.0.0.0:8090"}
GRPC_WEB_ADDRESS=${GRPC_WEB_ADDRESS:-"0.0.0.0:8091"}
API_ADDRESS=${API_ADDRESS:-"0.0.0.0:1318"}
JSONRPC_ADDRESS=${JSONRPC_ADDRESS:-"0.0.0.0:9545"}
JSONRPC_WS_ADDRESS=${JSONRPC_WS_ADDRESS:-"0.0.0.0:9546"}

TOKEN_AMOUNT=${TOKEN_AMOUNT:-"1000000000000000000000000adym"} #1M DYM
STAKING_AMOUNT=${STAKING_AMOUNT:-"670000000000000000000000adym"} #67% staked

# Validate dymension binary exists
export PATH=$PATH:$HOME/go/bin
if ! command -v dymd > /dev/null; then
  echo "dymension binary not found in $PATH"
  exit 1
fi

# Create and init dymension chain
dymd init "$MONIKER_NAME" --chain-id="$CHAIN_ID"

# Set chain parameters
set_consenus_params
set_gov_params
set_hub_params
set_misc_params
set_EVM_params
set_bank_denom_metadata
set_epochs_params
set_incentives_params
set_dymns_params

# Setup genesis account and transaction
dymd keys add "$KEY_NAME" --keyring-backend test
dymd add-genesis-account "$(dymd keys show "$KEY_NAME" -a --keyring-backend test)" "$TOKEN_AMOUNT"
dymd gentx "$KEY_NAME" "$STAKING_AMOUNT" --chain-id "$CHAIN_ID" --keyring-backend test
dymd collect-gentxs

#!/bin/sh

# Common commands
genesis_config_cmds="$(dirname "$0")/src/genesis_config_commands.sh"

if [ -f "$genesis_config_cmds" ]; then
  . "$genesis_config_cmds"
else
  echo "Error: header file not found" >&2
  exit 1
fi

# Set parameters
DATA_DIRECTORY="$HOME/.dymension"
CONFIG_DIRECTORY="$DATA_DIRECTORY/config"
TENDERMINT_CONFIG_FILE="$CONFIG_DIRECTORY/config.toml"
CLIENT_CONFIG_FILE="$CONFIG_DIRECTORY/client.toml"
APP_CONFIG_FILE="$CONFIG_DIRECTORY/app.toml"
GENESIS_FILE="$CONFIG_DIRECTORY/genesis.json"
CHAIN_ID=${CHAIN_ID:-"local-testnet"}
MONIKER_NAME=${MONIKER_NAME:-"local"}
KEY_NAME=${KEY_NAME:-"local-user"}
HUB_PEERS=${HUB_PEERS:-""}

SETTLEMENT_ADDR=${SETTLEMENT_ADDR:-"0.0.0.0:36657"}
P2P_ADDRESS=${P2P_ADDRESS:-"0.0.0.0:36656"}
GRPC_ADDRESS=${GRPC_ADDRESS:-"0.0.0.0:8090"}
GRPC_WEB_ADDRESS=${GRPC_WEB_ADDRESS:-"0.0.0.0:8091"}
API_ADDRESS=${API_ADDRESS:-"0.0.0.0:1317"}
UNSAFE_CORS=${UNSAFE_CORS:-""}

TOKEN_AMOUNT=${TOKEN_AMOUNT:-"1000000000000udym"} #1M DYM (1e6dym == 1e12udym)
STAKING_AMOUNT=${STAKING_AMOUNT:-"670000000000udym"} #67% is staked (inflation goal)

# Validate dymension binary exists
export PATH=$PATH:$HOME/go/bin
if ! command -v dymd > /dev/null; then
  make install

  if ! command -v dymd; then
    echo "dymension binary not found in $PATH"
    exit 1
  fi
fi

# Verify that a genesis file doesn't exists for the dymension chain
if [ -f "$GENESIS_FILE" ]; then
  printf "\n======================================================================================================\n"
  echo "A genesis file already exists. building the chain will delete all previous chain data. continue? (y/n)"
  read -r answer
  if [ "$answer" != "${answer#[Yy]}" ]; then
    rm -rf "$DATA_DIRECTORY"
  else
    exit 1
  fi
fi

# Create and init dymension chain
dymd init "$MONIKER_NAME" --chain-id="$CHAIN_ID"
dymd tendermint unsafe-reset-all

# ---------------------------------------------------------------------------- #
#                              Set configurations                              #
# ---------------------------------------------------------------------------- #
sed -i'' -e "/\[rpc\]/,+3 s/laddr *= .*/laddr = \"tcp:\/\/$SETTLEMENT_ADDR\"/" "$TENDERMINT_CONFIG_FILE"
sed -i'' -e "/\[p2p\]/,+3 s/laddr *= .*/laddr = \"tcp:\/\/$P2P_ADDRESS\"/" "$TENDERMINT_CONFIG_FILE"
sed  -i '' -e "s/^persistent_peers =.*/persistent_peers = \"$HUB_PEERS\"/" "$TENDERMINT_CONFIG_FILE"

sed -i'' -e "/\[grpc\]/,+6 s/address *= .*/address = \"$GRPC_ADDRESS\"/" "$APP_CONFIG_FILE"
sed -i'' -e "/\[grpc-web\]/,+7 s/address *= .*/address = \"$GRPC_WEB_ADDRESS\"/" "$APP_CONFIG_FILE"
sed -i'' -e "s/^chain-id *= .*/chain-id = \"$CHAIN_ID\"/" "$CLIENT_CONFIG_FILE"

sed -i'' -e "s/^node *= .*/node = \"tcp:\/\/$SETTLEMENT_ADDR\"/" "$CLIENT_CONFIG_FILE"
sed -i'' -e 's/bond_denom": ".*"/bond_denom": "udym"/' "$GENESIS_FILE"
sed -i'' -e 's/mint_denom": ".*"/mint_denom": "udym"/' "$GENESIS_FILE"

sed -i'' -e 's/^minimum-gas-prices *= .*/minimum-gas-prices = "0udym"/' "$APP_CONFIG_FILE"
sed -i'' -e '/\[api\]/,+3 s/enable *= .*/enable = true/' "$APP_CONFIG_FILE"
sed -i'' -e "/\[api\]/,+9 s/address *= .*/address = \"tcp:\/\/$API_ADDRESS\"/" "$APP_CONFIG_FILE"

set_distribution_params
set_gov_params
set_minting_params
set_staking_slashing_params
set_ibc_params
set_hub_params


if [ -n "$UNSAFE_CORS" ]; then
  echo "Setting CORS"
  sed -ie 's/enabled-unsafe-cors.*$/enabled-unsafe-cors = true/' "$APP_CONFIG_FILE"
  sed -ie 's/enable-unsafe-cors.*$/enabled-unsafe-cors = true/' "$APP_CONFIG_FILE"
  sed -ie 's/cors_allowed_origins.*$/cors_allowed_origins = ["*"]/' "$TENDERMINT_CONFIG_FILE"
fi

dymd keys add "$KEY_NAME" --keyring-backend test

if [ "$HUB_PEERS" != "" ]; then
  printf "\n======================================================================================================\n"
  echo "To join existing chain, copy the genesis file to $GENESIS_FILE"
  echo "To run a validator, run set_validator.sh after the node synced"
fi

echo "Do you want to initialize genesis accounts? (Y/n) "
read -r answer
if [ ! "$answer" != "${answer#[Nn]}" ] ;then
  dymd add-genesis-account "$(dymd keys show "$KEY_NAME" -a --keyring-backend test)" "$TOKEN_AMOUNT"
  dymd gentx "$KEY_NAME" "$STAKING_AMOUNT" --chain-id "$CHAIN_ID" --keyring-backend test
  dymd collect-gentxs
fi


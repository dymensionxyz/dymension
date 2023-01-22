#!/bin/sh

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

SETTLEMENT_RPC=${SETTLEMENT_RPC:-"0.0.0.0:36657"}
P2P_ADDRESS=${P2P_ADDRESS:-"0.0.0.0:36656"}
GRPC_ADDRESS=${GRPC_ADDRESS:-"0.0.0.0:8090"}
GRPC_WEB_ADDRESS=${GRPC_WEB_ADDRESS:-"0.0.0.0:8091"}

TOKEN_AMOUNT=${TOKEN_AMOUNT:-"1000000000000udym"} #1M DYM (1e6dym == 1e12udym)
STAKING_AMOUNT=${STAKING_AMOUNT:-"670000000000udym"} #67% is staked (inflation goal)

# Validate dymension binary exists
export PATH=$PATH:$HOME/go/bin
if ! command -v dymd; then
  make install

  if ! command -v dymd; then
    echo "dymension binary not found in $PATH"
    exit 1
  fi
fi

# Verify that a genesis file doesn't exists for the dymension chain
if [ -f "$GENESIS_FILE" ]; then
  printf "\n======================================================================================================"
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

sed -i'' -e "/\[rpc\]/,+3 s/laddr *= .*/laddr = \"tcp:\/\/$SETTLEMENT_RPC\"/" "$TENDERMINT_CONFIG_FILE"
sed -i'' -e "/\[p2p\]/,+3 s/laddr *= .*/laddr = \"tcp:\/\/$P2P_ADDRESS\"/" "$TENDERMINT_CONFIG_FILE"
sed -i'' -e "/\[grpc\]/,+6 s/address *= .*/address = \"$GRPC_ADDRESS\"/" "$APP_CONFIG_FILE"
sed -i'' -e "/\[grpc-web\]/,+7 s/address *= .*/address = \"$GRPC_WEB_ADDRESS\"/" "$APP_CONFIG_FILE"
sed -i'' -e "s/^chain-id *= .*/chain-id = \"$CHAIN_ID\"/" "$CLIENT_CONFIG_FILE"
sed -i'' -e "s/^node *= .*/node = \"tcp:\/\/$SETTLEMENT_RPC\"/" "$CLIENT_CONFIG_FILE"
sed -i'' -e 's/bond_denom": ".*"/bond_denom": "udym"/' "$GENESIS_FILE"
sed -i'' -e 's/mint_denom": ".*"/mint_denom": "udym"/' "$GENESIS_FILE"

dymd keys add "$KEY_NAME" --keyring-backend test
dymd add-genesis-account "$(dymd keys show "$KEY_NAME" -a --keyring-backend test)" "$TOKEN_AMOUNT"
dymd gentx "$KEY_NAME" "$STAKING_AMOUNT" --chain-id "$CHAIN_ID" --keyring-backend test
dymd collect-gentxs

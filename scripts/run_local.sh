#!/bin/sh

# .................................................... #
# ...........Set and validate parameters.............. #
# .................................................... #
DATA_DIRECTORY="$HOME/.dymension"
CONFIG_DIRECTORY="$DATA_DIRECTORY/config"
CLIENT_CONFIG_FILE="$CONFIG_DIRECTORY/client.toml"
GENESIS_FILE="$CONFIG_DIRECTORY/genesis.json"
CHAIN_ID=${CHAIN_ID:-"local-testnet"}
MONIKER_NAME=${MONIKER_NAME:-"local"}
KEY_NAME=${KEY_NAME:-"local-user"}

export PATH=$PATH:$HOME/go/bin
if ! command -v dymd; then
  echo "dYmension binary not found, call \"make install\"."
  exit 1
fi

# ...................................................... #
# ............... Init dymension chain ................. #
# ...................................................... #
if [ ! -f "$GENESIS_FILE" ]; then
  dymd init "$MONIKER_NAME" --chain-id="$CHAIN_ID"
  dymd tendermint unsafe-reset-all

  sed -i'' -e "s/^chain-id *= .*/chain-id = \"$CHAIN_ID\"/" "$CLIENT_CONFIG_FILE"
  sed -i'' -e 's/bond_denom": ".*"/bond_denom": "dym"/' "$GENESIS_FILE"
  sed -i'' -e 's/mint_denom": ".*"/mint_denom": "dym"/' "$GENESIS_FILE"

  dymd keys add "$KEY_NAME"
  dymd add-genesis-account "$(dymd keys show "$KEY_NAME" -a)" 100000000000dym
  dymd gentx "$KEY_NAME" 100000000dym --chain-id "$CHAIN_ID"
  dymd collect-gentxs
fi

# ...................................................... #
# ................... Run the chain .................... #
# ...................................................... #
dymd start




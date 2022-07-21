#!/bin/sh

# .................................................... #
# ...........Set and validate parameters.............. #
# .................................................... #

DATA_DIRECTORY="$HOME/.dymension/config"
TENDERMINT_CONFIG_FILE="$DATA_DIRECTORY/config.toml"
CLIENT_CONFIG_FILE="$DATA_DIRECTORY/client.toml"
CHAIN_ID=${CHAIN_ID:-""}
BLOCK_HEIGHT=${BLOCK_HEIGHT:-""}
BLOCK_HASH=${BLOCK_HASH:-""}
MONIKER_NAME=${MONIKER_NAME:-""}
TOKEN=${TOKEN:-""}

if [ -z "$CHAIN_ID" ]; then
  echo "Missing CHAIN_ID."
  exit 1
fi

if [ -z "$MONIKER_NAME" ]; then
  echo "Missing MONIKER_NAME."
  exit 1
fi

currentGenesisFile="$DATA_DIRECTORY/genesis.json"
if [ -f "$currentGenesisFile" ]; then
  echo "Genesis file exists ($currentGenesisFile), you should remove it before continue"
  exit 1
fi

# ...................................................... #
# .................. Init dYmension .................... #
# ...................................................... #

if ! command -v dymd; then
  echo "dYmension binary not found, call \"make install\"."
  exit 1
fi

dymd init "$MONIKER_NAME" --chain-id="$CHAIN_ID"
dymd tendermint unsafe-reset-all

# ...................................................... #
# .................. Fetch genesis ..................... #
# ...................................................... #

genesisPath="raw.githubusercontent.com/dymensionxyz/networks/main/$CHAIN_ID/genesis.json"
if [ "$TOKEN" ]; then
  genesisPath="$TOKEN@$genesisPath"
fi
curl -s "https://$genesisPath" >genesis.json
if grep -q "404: Not Found" "genesis.json"; then
  rm genesis.json
  echo "Can't download genesis file"
  exit 1
fi
mv genesis.json "$DATA_DIRECTORY/"

# ...................................................... #
# ............... Update configurations ................ #
# ...................................................... #

PEERS='81d3d1cc389ac41a36c89c449bfc4282f7b494ef@44.209.89.17:26656'
tsed -i'' -e "s/^chain-id *= .*/chain-id = \"$CHAIN_ID\"/" "$CLIENT_CONFIG_FILE"

if [ "$BLOCK_HEIGHT" ] && [ "$BLOCK_HASH" ]; then
  sed -i'' -e 's/^enable *= false/enable = true/' "$TENDERMINT_CONFIG_FILE"
  sed -i'' -e "s/^trust_height *= .*/trust_height = $BLOCK_HEIGHT/" "$TENDERMINT_CONFIG_FILE"
  sed -i'' -e "s/^trust_hash *= .*/trust_hash = \"$BLOCK_HASH\"/" "$TENDERMINT_CONFIG_FILE"
  sed -i'' -e 's/^rpc_servers *= .*/rpc_servers = "https:\/\/rpc.cosmos.network:443"/' "$TENDERMINT_CONFIG_FILE"
fi


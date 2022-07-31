#!/bin/sh

# .................................................... #
# ...........Set and validate parameters.............. #
# .................................................... #

DATA_DIRECTORY="$HOME/.dymension"
CONFIG_DIRECTORY="$DATA_DIRECTORY/config"
TENDERMINT_CONFIG_FILE="$CONFIG_DIRECTORY/config.toml"
CLIENT_CONFIG_FILE="$CONFIG_DIRECTORY/client.toml"
CHAIN_ID=${CHAIN_ID:-""}
MONIKER_NAME=${MONIKER_NAME:-""}
TOKEN=${TOKEN:-""}
STATE_SYNC=${STATE_SYNC:-false}
CHAIN_REPO=https://$([ -z "$TOKEN" ] && echo "" || echo "$TOKEN@")raw.githubusercontent.com/dymensionxyz/networks/main

echo "$CHAIN_ID"
if [ -z "$CHAIN_ID" ]; then
  echo "Missing CHAIN_ID."
  exit 1
fi

if [ -z "$MONIKER_NAME" ]; then
  echo "Missing MONIKER_NAME."
  exit 1
fi

genesisFilePath="$CONFIG_DIRECTORY/genesis.json"
if [ -f "$genesisFilePath" ]; then
  echo "Genesis file exists ($genesisFilePath), you should remove it before continue"
  exit 1
fi

# ...................................................... #
# .................. Init dYmension .................... #
# ...................................................... #

export PATH=$PATH:$HOME/go/bin
if ! command -v dymd; then
  echo "dYmension binary not found, call \"make install\"."
  exit 1
fi

dymd init "$MONIKER_NAME" --chain-id="$CHAIN_ID"
dymd tendermint unsafe-reset-all

# ...................................................... #
# .................. Fetch genesis ..................... #
# ...................................................... #

curl -s "$CHAIN_REPO/$CHAIN_ID/genesis.json" >genesis.json
if grep -q "404: Not Found" "genesis.json"; then
  rm genesis.json
  echo "Can't download genesis file"
  rm -rf "$DATA_DIRECTORY"
  exit 1
fi
mv genesis.json "$CONFIG_DIRECTORY/"

# ...................................................... #
# ............... Update configurations ................ #
# ...................................................... #

PEERS="$(curl -s "$CHAIN_REPO/$CHAIN_ID/persistent_peers.txt")"
sed -i'' -e "s/^persistent_peers *= .*/persistent_peers = \"$PEERS\"/" "$TENDERMINT_CONFIG_FILE"
sed -i'' -e 's/^allow_duplicate_ip = false/allow_duplicate_ip = true/' "$TENDERMINT_CONFIG_FILE"
sed -i'' -e "s/^chain-id *= .*/chain-id = \"$CHAIN_ID\"/" "$CLIENT_CONFIG_FILE"

if [ "$STATE_SYNC" -eq 1 ]; then
  RPC_SERVER="rpc.dymension.xyz:26657"
  TRUSTED_BLOCK=$(curl "$RPC_SERVER/commit" | jq -r "{height: .result.signed_header.header.height, hash: .result.signed_header.commit.block_id.hash}")
  BLOCK_HEIGHT=$(echo "$TRUSTED_BLOCK" | jq -r ".height")
  BLOCK_HASH=$(echo "$TRUSTED_BLOCK" | jq -r ".hash")
  sed -i'' -e 's/^enable *= false/enable = true/' "$TENDERMINT_CONFIG_FILE"
  sed -i'' -e "s/^trust_height *= .*/trust_height = $BLOCK_HEIGHT/" "$TENDERMINT_CONFIG_FILE"
  sed -i'' -e "s/^trust_hash *= .*/trust_hash = \"$BLOCK_HASH\"/" "$TENDERMINT_CONFIG_FILE"
  sed -i'' -e "s/^rpc_servers *= .*/rpc_servers = \"http:\/\/$RPC_SERVER,http:\/\/$RPC_SERVER\"/" "$TENDERMINT_CONFIG_FILE"
fi

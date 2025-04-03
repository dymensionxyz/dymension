# For a local explorer with anvil

ANVIL_RPC_URL=http://127.0.0.1:8545

docker run \
  --rm \
  -p 5100:80 \
  --name otterscan \
  -d \
  --env OTTERSCAN_CONFIG='{
    "erigonURL": "'$ANVIL_RPC_URL'",
    "assetsURLPrefix": "http://127.0.0.1:5175",
    "branding": {
        "siteName": "My Otterscan",
        "networkTitle": "Dev Network"
    },
}' \
otterscan/otterscan:latest

# Then go to http://localhost:5100/

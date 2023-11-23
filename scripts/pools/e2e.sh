#!/bin/bash

create_asset_pool() {
    dymd tx gamm create-pool --pool-file=$1 --from pools  --keyring-backend=test -b block --gas auto --yes
}

# create pools and pool gauges
echo "====================="
echo "Creating pools"
echo "Creating udym/uatom pool"
create_asset_pool "$(dirname "$0")/nativeDenomPoolA.json"
echo "Creating udym/uusd pool"
create_asset_pool "$(dirname "$0")/nativeDenomPoolB.json"

# fund streamer
echo "====================="
streamer_addr=$(dymd q auth module-account streamer -o json | jq '.account.base_account.address' | tr -d '"')
echo "Sending 300000dym to $streamer_addr"
dymd tx bank send local-user $streamer_addr 300000dym --keyring-backend test -b block -y

# lock LP tokens
echo "====================="
echo "Locking LP tokens"
echo "locking LP1 tokens for 1 day"
dymd tx lockup lock-tokens 50000000000000000000gamm/pool/1 --duration="24h" --from pools --keyring-backend=test -b block -y
echo "locking LP2 tokens for 1 minute"
dymd tx lockup lock-tokens 50000000000000000000gamm/pool/2 --duration="1m" --from pools --keyring-backend=test -b block -y


# create new stream
echo "====================="
echo "Gov proposal for creating new stream with LP1 and LP2 as incentives targets"
dymd tx gov submit-legacy-proposal create-stream-proposal 1,2 40,60 10000dym --epoch-identifier minute --from local-user -b block --title sfasfas --description ddasda --deposit 1dym -y
last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')
dymd tx gov vote "$last_proposal_id" yes --from local-user -b block -y
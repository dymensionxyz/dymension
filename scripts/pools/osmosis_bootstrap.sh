#!/bin/sh

# Define the functions using the new function
create_asset_pool() {
    dymd tx gamm create-pool --pool-file=$1 --from pools  --keyring-backend=test -b block --gas auto --yes
}



create_asset_pool "$(dirname "$0")/nativeDenomPoolA.json"
create_asset_pool "$(dirname "$0")/nativeDenomPoolB.json"
create_asset_pool "$(dirname "$0")/nativeDenomThreeAssetPool.json"
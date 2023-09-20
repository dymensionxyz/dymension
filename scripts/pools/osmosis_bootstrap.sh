#!/bin/sh

# Define the functions using the new function
create_asset_pool() {
    dymd tx gamm create-pool --pool-file=$1 --from pools  --keyring-backend=test -b block --gas auto --yes
}


join_to_pool() {
    dymd tx gamm join-pool --pool-id 1 --share-amount-out 40227549469722224220 --from local-user --keyring-backend=test
}

exit_pool() {
    dymd tx gamm exit-pool --pool-id=$1 --shares=$2 --from pools  --keyring-backend=test -b block --gas auto --yes
}

swap_tokens() {
    # dymd tx gamm swap --exact-amount-in=$1 --exact-amount-out=$2 --from pools  --keyring-backend=test -b block --gas auto --yes
    dymd tx gamm swap-exact-amount-in 5000uatom 5000 --swap-route-pool-ids 1 --swap-route-denoms uusd --from local-user --keyring-backend test
}


create_asset_pool "$(dirname "$0")/nativeDenomPoolA.json"
create_asset_pool "$(dirname "$0")/nativeDenomPoolB.json"
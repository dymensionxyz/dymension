#!/bin/sh

# Define the functions using the new function
create_asset_pool() {
    dymd tx gamm create-pool --pool-file=$1 --from pools  --keyring-backend=test -b block --gas auto --yes
}


join_to_pool() {
    dymd tx gamm join-pool --pool-id 1 --share-amount-out 20000000000000000000  --max-amounts-in "" --from user --keyring-backend test -b block
    dymd tx gamm join-swap-extern-amount-in 100dym 20000000000000000000 --pool-id 1  --from user --keyring-backend test  -b block
}

exit_pool() {
    dymd tx gamm exit-pool --pool-id=$1 --shares=$2 --from pools  --keyring-backend=test -b block --gas auto --yes
}

swap_tokens() {
    # dymd tx gamm swap --exact-amount-in=$1 --exact-amount-out=$2 --from pools  --keyring-backend=test -b block --gas auto --yes
    dymd tx gamm swap-exact-amount-in 50udym 50000000 --swap-route-pool-ids 1 --swap-route-denoms uatom --from user --keyring-backend test -b block
}

multi_hop_swap() {
    dymd tx gamm swap-exact-amount-in 50000000uatom 20000000 --swap-route-pool-ids 1,2 --swap-route-denoms udym,uusd --from user --keyring-backend test -b block
}

echo "Creating pools"
echo "Creating udym/uatom pool"
create_asset_pool "$(dirname "$0")/nativeDenomPoolA.json"
echo "Creating udym/uusd pool"
create_asset_pool "$(dirname "$0")/nativeDenomPoolB.json"
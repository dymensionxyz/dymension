#!/bin/sh
BINARY=dymd

# An example of creating liquidity pool 1
$BINARY tx liquidity create-pool 1 1000000000uatom,50000000000uusd --from provider1 --keyring-backend test -b block --gas auto -y

# An example of creating liquidity pool 2
$BINARY tx liquidity create-pool 1 10000000000000000000udym,10000000uusd --from provider2 --keyring-backend test -b block --gas auto -y

# An example of requesting swap
$BINARY tx liquidity swap 1 1 50000000uusd uatom 0.019 0.003 --from provider2 --keyring-backend test -b block --gas auto -y

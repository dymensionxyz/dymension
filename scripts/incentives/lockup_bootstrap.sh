#!/bin/sh

echo "locking LP1 tokens for 2 weeks"
dymd tx lockup lock-tokens 50000000000000000000gamm/pool/1 --duration="60s" --from pools --keyring-backend=test -b sync -y --gas-prices 100000000adym
sleep 7

echo "locking uatom tokens for 1h"
dymd tx lockup lock-tokens 500000000uatom --duration="3600s" --from user --keyring-backend=test -b sync -y --gas-prices 100000000adym
sleep 7

echo "locking dym tokens for 1 day"
dymd tx lockup lock-tokens 100dym --duration="24h" --from user --keyring-backend=test -b sync -y --gas-prices 100000000adym
sleep 7

echo "unlocking dym tokens"
dymd tx lockup begin-unlock-by-id 2 --from user --keyring-backend=test -b sync -y --gas-prices 100000000adym

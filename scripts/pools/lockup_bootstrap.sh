#!/bin/sh

LPDENOM=gamm/pool/1

# Define the functions using the new function
dymd tx lockup lock-tokens 50000000000000000000gamm/pool/1 --duration="1h" --from pools --keyring-backend=test -b block -y


dymd tx lockup lock-tokens 100dym --duration="24h" --from user --keyring-backend=test -b block -y
dymd tx lockup begin-unlock-by-id 2 --from user --keyring-backend=test -b block -y



# Add external incentives to the pool gauge
dymd tx incentives add-to-gauge 1 100dym --from user -b block -y
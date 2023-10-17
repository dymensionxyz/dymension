#!/bin/sh

# Define the functions using the new function
dymd tx lockup lock-tokens 100dym --duration="1h" --from user --keyring-backend=test -b block -y
dymd tx lockup lock-tokens 200dym --duration="24h" --from user --keyring-backend=test -b block -y
dymd tx lockup lock-tokens 300dym --duration="336h" --from user --keyring-backend=test -b block -y


dymd tx lockup begin-unlock-by-id 1 --from=user --keyring-backend=test -b block -y

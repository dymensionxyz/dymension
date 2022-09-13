#!/bin/sh

# Validate necessary data is exists
export PATH=$PATH:$HOME/go/bin
if ! command -v dymd || [ ! -f "$HOME/.dymension/config/genesis.json" ]; then
  echo "dYmension binary or genesis file are not exists, run 'setup_local.sh' before"
  exit 1
fi

# Run the dymension chain
dymd start

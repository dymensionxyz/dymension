#!/bin/sh

if ! ./scripts/setup_dymension.sh; then
  exit 1
fi
./scripts/run_dymension.sh


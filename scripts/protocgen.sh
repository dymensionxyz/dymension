#!/usr/bin/env bash

set -eo pipefail

# get protoc executions
# go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos 2>/dev/null


echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ../proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
    for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
      if grep go_package $file &>/dev/null; then
        echo "Generating gogo proto code for $file"
        buf generate --template buf.gen.gogo.yaml $file
      fi
    done
done

cd ..

# move proto files to the right places
# Find the first versioned directory and copy its contents
if [ -d "github.com/dymensionxyz/dymension" ]; then
    VERSION_DIR=$(ls -d github.com/dymensionxyz/dymension/v* 2>/dev/null | head -n 1)
    if [ -n "$VERSION_DIR" ] && [ -d "$VERSION_DIR" ]; then
        cp -r "$VERSION_DIR"/* ./
    fi
fi

rm -rf github.com

# TODO: Uncomment once ORM/Pulsar support is needed.
#
# Ref: https://github.com/osmosis-labs/osmosis/pull/1589
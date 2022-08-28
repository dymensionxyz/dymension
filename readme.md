# dYmension

1. [Install Go](https://go.dev/doc/install) 1.18+
1. Clone this repo
1. Install the dymd CLI

    ```shell
    make install
    ```

## Usage

```sh
# Print help message
dymd --help

# Create your own single node devnet
dymd init test --chain-id test
dymd keys add user1
dymd add-genesis-account <address from above command> 10000000stake,1000token
dymd gentx user1 1000000stake --chain-id test
dymd collect-gentxs
dymd start

See <https://docs.dymension.xyz/node-runners/dymension-hub> for more information

## Contributing

### Tools

1. Install [golangci-lint](https://golangci-lint.run/usage/install/)

### Helpful Commands

```sh
# Build a new dymd binary and output to build/dymd
make build

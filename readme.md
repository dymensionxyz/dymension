# Dymension Hub

![image](./docs/dymension.png)

![license](https://img.shields.io/github/license/dymensionxyz/dymension)
![Go](https://img.shields.io/badge/go-1.18-blue.svg)
![issues](https://img.shields.io/github/issues/dymensionxyz/dymension)
![tests](https://github.com/dymensionxyz/dymint/actions/workflows/test.yml/badge.svg?branch=main)
![lint](https://github.com/dymensionxyz/dymint/actions/workflows/lint.yml/badge.svg?branch=main)

## Overview

Welcome to the Dymension Hub, the **Settlement Layer of the Dymension protocol**.

This guide will walk you through the steps required to set up and run a Dymension Hub full node.

## Table of Contents

- [Dymension Hub](#dymension-hub)
  - [Overview](#overview)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Initializing `dymd`](#initializing-dymd)
  - [Running the Chain](#running-the-chain)
  - [Adding liquidity](#adding-liquidity)
  - [Adding incentives](#adding-incentives)
    - [Setting incentivised pools and weights](#setting-incentivised-pools-and-weights)
    - [Locking tokens](#locking-tokens)
    - [Fund the incentives](#fund-the-incentives)

## Prerequisites

- [Go (v1.18 or above)](https://go.dev/doc/install)

## Installation

Clone `dymension`:

```sh
git clone https://github.com/dymensionxyz/dymension.git
cd dymension
make install
```

Check that the dymd binaries have been successfully installed:

```sh
dymd version
```

If the dymd command is not found an error message is returned,
confirm that your [GOPATH](https://go.dev/doc/gopath_code#GOPATH) is correctly configured by running the following command:

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

## Initializing `dymd`

- Using the setup script:

    This method is preffered as it preconfigured to support [running rollapps locally](https://github.com/dymensionxyz/roller)

    ```sh
    bash scripts/setup_local.sh
    ```

- Manually:

    First, set the following environment variables:

    ```sh
    export CHAIN_ID="dymension_100-1"
    export KEY_NAME="local-user"
    export MONIKER_NAME="local"
    ```

    Then, initialize a chain with a user:

    ```sh
    dymd init "$MONIKER_NAME" --chain-id "$CHAIN_ID"
    dymd keys add "$KEY_NAME" --keyring-backend test
    dymd add-genesis-account "$(dymd keys show "$KEY_NAME" -a --keyring-backend test)" 100000000000udym
    dymd gentx "$KEY_NAME" 1000000dym --chain-id "$CHAIN_ID" --keyring-backend test
    dymd collect-gentxs
    ```

## Running the Chain

Now start the chain!

```sh
dymd start
```

You should have a running local node!

## Adding liquidity

To bootstrap the `GAMM` module with pools:

```sh
sh scripts/pools/pools_bootstrap.sh
```

## Adding incentives

### Setting incentivised pools and weights

After creating the pools above, set the incentives weights through gov:

```sh
sh scripts/incentives/incentive_pools_bootstrap.sh
```

wait for the gov proposal to pass, and valitate with:

```sh
dymd q poolincentives distr-info
```

### Locking tokens

To get incentives, we need to lock the LP tokens:

```sh
sh scripts/incentives/lockup_bootstrap.sh
```

Valitate with:

```sh
dymd q lockup module-balance
```

### Fund the incentives

Now we fund the pool incentives.
The funds will be distributed between the incentivised pools according to weights.
The following script funds the pool incentives both by external funds (direct funds transfer from some user)
and using the community pool (using gov proposal)

```sh
sh scripts/incentives/fund_incentives.sh
```

validate with:

```sh
dymd q incentives gauges
```

If you have any issues please contact us on [discord](http://discord.gg/dymension) in the Developer section. We are here for you!

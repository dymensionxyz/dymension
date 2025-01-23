package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/evmos/ethermint/types"
)

// Set additional config
// prefix and denoms registered on app init
func initSDKConfig() {
	config := sdk.GetConfig()
	SetBip44CoinType(config)
	config.Seal()
}

// SetBip44CoinType sets the global coin type to be used in hierarchical deterministic wallets.
func SetBip44CoinType(config *sdk.Config) {
	config.SetCoinType(ethermint.Bip44CoinType)
	config.SetPurpose(sdk.Purpose) // Shared
}

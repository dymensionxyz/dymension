package cmd

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"

	ethermint "github.com/evmos/ethermint/types"
)

// Set additional config
// prefix and denoms registered on app init
func initSDKConfig() {
	config := sdk.GetConfig()

	appparams.SetAddressPrefixes(config)
	SetBip44CoinType(config)
	config.Seal()

	RegisterDenoms()
}

// RegisterDenoms registers the base and display denominations to the SDK.
func RegisterDenoms() {
	if err := sdk.RegisterDenom(appparams.DisplayDenom, math.LegacyOneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(appparams.BaseDenom, math.LegacyNewDecWithPrec(1, appparams.BaseDenomUnit)); err != nil {
		panic(err)
	}
}

// SetBip44CoinType sets the global coin type to be used in hierarchical deterministic wallets.
func SetBip44CoinType(config *sdk.Config) {
	config.SetCoinType(ethermint.Bip44CoinType)
	config.SetPurpose(sdk.Purpose) // Shared
}

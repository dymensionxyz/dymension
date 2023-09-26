package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/dymensionxyz/dymension/app/params"

	ethcmd "github.com/evmos/ethermint/cmd/config"
)

func initSDKConfig() {

	appparams.SetAddressPrefixes()

	// Set config
	config := sdk.GetConfig()
	ethcmd.SetBip44CoinType(config)
	registerDenoms()

	config.Seal()
}

// RegisterDenoms registers the base and display denominations to the SDK.
func registerDenoms() {
	if err := sdk.RegisterDenom(appparams.DisplayDenom, sdk.OneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(appparams.BaseDenom, sdk.NewDecWithPrec(1, appparams.BaseDenomUnit)); err != nil {
		panic(err)
	}
}

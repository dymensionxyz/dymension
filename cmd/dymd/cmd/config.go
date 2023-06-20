package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/app"

	ethcmd "github.com/evmos/ethermint/cmd/config"
)

const (
	// DisplayDenom is the denomination used to display the amount of tokens held
	DisplayDenom = "dym"
	// BaseDenomUnit is the base denom unit for the Hub
	BaseDenom = "udym"

	// BaseDenomUnit defines the base denomination unit for Photons.
	// 1 DYM = 1x10^{BaseDenomUnit} udym
	BaseDenomUnit = 18
)

func initSDKConfig() {
	// Set prefixes
	accountPubKeyPrefix := app.AccountAddressPrefix + "pub"
	validatorAddressPrefix := app.AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := app.AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := app.AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := app.AccountAddressPrefix + "valconspub"

	// Set config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)

	ethcmd.SetBip44CoinType(config)

	config.Seal()

	registerDenoms()
}

// RegisterDenoms registers the base and display denominations to the SDK.
func registerDenoms() {
	if err := sdk.RegisterDenom(DisplayDenom, sdk.OneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(BaseDenom, sdk.NewDecWithPrec(1, BaseDenomUnit)); err != nil {
		panic(err)
	}
}

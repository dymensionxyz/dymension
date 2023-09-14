package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/dymensionxyz/dymension/app/params"

	ethcmd "github.com/evmos/ethermint/cmd/config"
)

func initSDKConfig() {
	// Set prefixes
	accountPubKeyPrefix := appparams.AccountAddressPrefix + "pub"
	validatorAddressPrefix := appparams.AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := appparams.AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := appparams.AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := appparams.AccountAddressPrefix + "valconspub"

	// Set config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(appparams.AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)

	ethcmd.SetBip44CoinType(config)

	config.Seal()

	registerDenoms()
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

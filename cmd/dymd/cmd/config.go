package cmd

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethermint "github.com/evmos/ethermint/types"
)

// Set additional config
// prefix and denoms registered on app init
func initSDKConfig() {
	config := sdk.GetConfig()

	SetAddressPrefixes(config)
	SetBip44CoinType(config)
	config.Seal()

	RegisterDenoms()
}

// RegisterDenoms registers the base and display denominations to the SDK.
func RegisterDenoms() {
	if err := sdk.RegisterDenom(appparams.DisplayDenom, math.LegacyNewDecWithPrec(1, 1)); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(appparams.BaseDenom, math.LegacyNewDecWithPrec(1, appparams.BaseDenomUnit)); err != nil {
		panic(err)
	}
}

func SetAddressPrefixes(config *sdk.Config) {
	// Set prefixes
	accountPubKeyPrefix := appparams.AccountAddressPrefix + "pub"
	validatorAddressPrefix := appparams.AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := appparams.AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := appparams.AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := appparams.AccountAddressPrefix + "valconspub"

	// Set config
	config.SetBech32PrefixForAccount(appparams.AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)

	config.SetAddressVerifier(func(bytes []byte) error {
		if len(bytes) == 0 {
			return errorsmod.Wrap(sdkerrors.ErrUnknownAddress, "addresses cannot be empty")
		}

		if len(bytes) > address.MaxAddrLen {
			return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", address.MaxAddrLen, len(bytes))
		}

		if len(bytes) != 20 && len(bytes) != 32 {
			return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address length must be 20 or 32 bytes, got %d", len(bytes))
		}

		return nil
	})
}

// SetBip44CoinType sets the global coin type to be used in hierarchical deterministic wallets.
func SetBip44CoinType(config *sdk.Config) {
	config.SetCoinType(ethermint.Bip44CoinType)
	config.SetPurpose(sdk.Purpose) // Shared
}

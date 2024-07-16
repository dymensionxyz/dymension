package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func NewRollapp(
	creator,
	rollappId,
	initSequencerAddress,
	bech32Prefix,
	genesisChecksum,
	website,
	description,
	logoDataUri,
	alias string,
	transfersEnabled bool,
) Rollapp {
	ret := Rollapp{
		RollappId:               rollappId,
		Creator:                 creator,
		InitialSequencerAddress: initSequencerAddress,
		GenesisChecksum:         genesisChecksum,
		Bech32Prefix:            bech32Prefix,
		Website:                 website,
		Description:             description,
		LogoDataUri:             logoDataUri,
		Alias:                   alias,
	}
	ret.GenesisState.TransfersEnabled = transfersEnabled
	return ret
}

func (r Rollapp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(r.Creator)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidCreatorAddress, err.Error())
	}

	// validate rollappId
	_, err = NewChainID(r.RollappId)
	if err != nil {
		return err
	}

	_, err = sdk.AccAddressFromBech32(r.InitialSequencerAddress)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidInitialSequencerAddress, err.Error())
	}

	if err = validateBech32Prefix(r.Bech32Prefix); err != nil {
		return errorsmod.Wrap(ErrInvalidBech32Prefix, err.Error())
	}

	if r.GenesisChecksum == "" {
		return errorsmod.Wrap(ErrEmptyGenesisChecksum, "GenesisChecksum")
	}

	return nil
}

func validateBech32Prefix(prefix string) error {
	bechAddr, err := sdk.Bech32ifyAddressBytes(prefix, sample.Acc())
	if err != nil {
		return err
	}

	bAddr, err := sdk.GetFromBech32(bechAddr, prefix)
	if err != nil {
		return err
	}

	if err = sdk.VerifyAddressFormat(bAddr); err != nil {
		return err
	}
	return nil
}

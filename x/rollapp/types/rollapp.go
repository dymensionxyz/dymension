package types

import (
	"net/url"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func NewRollapp(
	creator,
	rollappId,
	initSequencerAddress,
	bech32Prefix string,
	genesisInfo GenesisInfo,
	transfersEnabled bool,
) Rollapp {
	ret := Rollapp{
		RollappId:               rollappId,
		Creator:                 creator,
		InitialSequencerAddress: initSequencerAddress,
		GenesisInfo:             genesisInfo,
		Bech32Prefix:            bech32Prefix,
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

	// validate GenesisChecksum
	if err = r.ValidateGenesisInfo(); err != nil {
		return errorsmod.Wrap(ErrEmptyGenesisChecksum, err.Error())
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

func (r Rollapp) ValidateGenesisInfo() error {
	// validate GenesisChecksum
	if r.GenesisInfo.GenesisChecksum == "" {
		return errorsmod.Wrap(ErrEmptyGenesisChecksum, "GenesisChecksum")
	}

	// validate GenesisURLs
	if len(r.GenesisInfo.GenesisUrls) == 0 {
		return errorsmod.Wrap(ErrEmptyGenesisURLs, "GenesisURLs")
	}

	for _, u := range r.GenesisInfo.GenesisUrls {
		// validate url
		_, err := url.Parse(u)
		if err != nil {
			return errorsmod.Wrap(err, "GenesisURL")
		}
	}

	return nil
}

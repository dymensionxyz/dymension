package types

import (
	"net/url"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewRollapp(
	creator,
	rollappId,
	initSequencerAddress,
	bech32Prefix string,
	genesisInfo *GenesisInfo,
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
		return errorsmod.Wrap(err, ErrInvalidCreatorAddress.Error())
	}

	// validate rollappId
	_, err = NewChainID(r.RollappId)
	if err != nil {
		return err
	}

	if r.InitialSequencerAddress == "" {
		return errorsmod.Wrap(ErrEmptyInitialSequencerAddress, "InitialSequencerAddress")
	}

	// validate Bech32Prefix
	if _, err := sdk.AccAddressFromBech32(r.Bech32Prefix); err != nil {
		return errorsmod.Wrap(err, ErrInvalidBech32Prefix.Error())
	}

	// validate GenesisInfo
	if r.GenesisInfo == nil {
		return errorsmod.Wrap(ErrNilGenesisInfo, "GenesisInfo")
	}

	// validate GenesisChecksum
	if r.GenesisInfo.GenesisChecksum == "" {
		return errorsmod.Wrap(ErrEmptyGenesisChecksum, "GenesisChecksum")
	}

	// validate GenesisURLs
	if len(r.GenesisInfo.GenesisUrls) == 0 {
		return errorsmod.Wrap(ErrEmptyGenesisURLs, "GenesisURLs")
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

package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewUpdateRollappInformation(
	creator,
	rollappId,
	initSequencerAddress,
	genesisChecksum,
	alias string,
	metadata *RollappMetadata,
) *UpdateRollappInformation {
	return &UpdateRollappInformation{
		RollappId:               rollappId,
		Creator:                 creator,
		InitialSequencerAddress: initSequencerAddress,
		GenesisChecksum:         genesisChecksum,
		Alias:                   alias,
		Metadata:                metadata,
	}
}

func (r UpdateRollappInformation) ValidateBasic() error {
	if r.InitialSequencerAddress != "" {
		_, err := sdk.AccAddressFromBech32(r.InitialSequencerAddress)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidInitialSequencerAddress, err.Error())
		}
	}

	if len(r.Alias) > maxAliasLength {
		return ErrInvalidAlias
	}

	if len(r.GenesisChecksum) > maxGenesisChecksumLength {
		return ErrInvalidGenesisChecksum
	}

	if err := validateMetadata(r.Metadata); err != nil {
		return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
	}

	return nil
}

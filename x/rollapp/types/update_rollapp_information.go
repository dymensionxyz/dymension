package types

import (
	"unicode"

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

	if err := validateAlias(r.Alias); err != nil {
		return err
	}

	if len(r.GenesisChecksum) > maxGenesisChecksumLength {
		return ErrInvalidGenesisChecksum
	}

	if err := validateMetadata(r.Metadata); err != nil {
		return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
	}

	return nil
}

func validateAlias(alias string) error {
	if l := len(alias); l == 0 || l > maxAliasLength {
		return ErrInvalidAlias
	}

	// only allow alphanumeric characters and underscores
	for _, c := range alias {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_' {
			return ErrInvalidAlias
		}
	}

	return nil
}

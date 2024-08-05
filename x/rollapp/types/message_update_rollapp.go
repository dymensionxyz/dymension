package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgUpdateRollappInformation = "update_rollapp"

var _ sdk.Msg = &MsgUpdateRollappInformation{}

func NewMsgUpdateRollappInformation(
	creator,
	rollappId,
	initSequencerAddress,
	genesisChecksum,
	alias string,
	metadata *RollappMetadata,
) *MsgUpdateRollappInformation {
	return &MsgUpdateRollappInformation{
		Owner:                   creator,
		RollappId:               rollappId,
		InitialSequencerAddress: initSequencerAddress,
		GenesisChecksum:         genesisChecksum,
		Alias:                   alias,
		Metadata:                metadata,
	}
}

func (msg *MsgUpdateRollappInformation) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRollappInformation) Type() string {
	return TypeMsgUpdateRollappInformation
}

func (msg *MsgUpdateRollappInformation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateRollappInformation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRollappInformation) ValidateBasic() error {
	if msg.InitialSequencerAddress != "" {
		_, err := sdk.AccAddressFromBech32(msg.InitialSequencerAddress)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidInitialSequencerAddress, err.Error())
		}
	}

	if err := validateAlias(msg.Alias); err != nil {
		return err
	}

	if len(msg.GenesisChecksum) > maxGenesisChecksumLength {
		return ErrInvalidGenesisChecksum
	}

	if err := validateMetadata(msg.Metadata); err != nil {
		return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
	}

	return nil
}

func (msg *MsgUpdateRollappInformation) UpdatingImmutableValues() bool {
	return msg.InitialSequencerAddress != "" || msg.GenesisChecksum != "" || msg.Alias != ""
}

package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const TypeMsgUpdateRollappInformation = "update_rollapp"

var _ sdk.Msg = &MsgUpdateRollappInformation{}

func NewMsgUpdateRollappInformation(
	creator,
	rollappId,
	initSequencer string,
	metadata *RollappMetadata,
	genesisInfo *GenesisInfo,
) *MsgUpdateRollappInformation {
	return &MsgUpdateRollappInformation{
		Owner:            creator,
		RollappId:        rollappId,
		InitialSequencer: initSequencer,
		Metadata:         metadata,
		GenesisInfo:      genesisInfo,
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
	if msg.InitialSequencer != "" && msg.InitialSequencer != "*" {
		_, err := sdk.AccAddressFromBech32(msg.InitialSequencer)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidInitialSequencer, err.Error())
		}
	}

	if msg.GenesisInfo != nil {
		if len(msg.GenesisInfo.GenesisChecksum) > maxGenesisChecksumLength {
			return ErrInvalidGenesisChecksum
		}

		if msg.GenesisInfo.Bech32Prefix != "" {
			if err := validateBech32Prefix(msg.GenesisInfo.Bech32Prefix); err != nil {
				return errorsmod.Wrap(errors.Join(err, gerrc.ErrInvalidArgument), "bech32 prefix")
			}
		}
	}

	if msg.Metadata != nil {
		if err := msg.Metadata.Validate(); err != nil {
			return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
		}
	}

	return nil
}

func (msg *MsgUpdateRollappInformation) UpdatingImmutableValues() bool {
	return msg.InitialSequencer != ""
}

func (msg *MsgUpdateRollappInformation) UpdatingGenesisInfo() bool {
	if msg.GenesisInfo == nil {
		return false
	}
	return msg.GenesisInfo.GenesisChecksum != "" ||
		msg.GenesisInfo.Bech32Prefix != "" ||
		msg.GenesisInfo.NativeDenom != nil ||
		!msg.GenesisInfo.InitialSupply.IsNil()
}

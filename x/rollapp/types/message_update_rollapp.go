package types

import (
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
		Update: &UpdateRollappInformation{
			Creator:                 creator,
			RollappId:               rollappId,
			InitialSequencerAddress: initSequencerAddress,
			GenesisChecksum:         genesisChecksum,
			Alias:                   alias,
			Metadata:                metadata,
		},
	}
}

func (msg *MsgUpdateRollappInformation) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRollappInformation) Type() string {
	return TypeMsgUpdateRollappInformation
}

func (msg *MsgUpdateRollappInformation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Update.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateRollappInformation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRollappInformation) GetUpdateRollappInformation() *UpdateRollappInformation {
	return NewUpdateRollappInformation(
		msg.Update.Creator,
		msg.Update.RollappId,
		msg.Update.InitialSequencerAddress,
		msg.Update.GenesisChecksum,
		msg.Update.Alias,
		msg.Update.Metadata,
	)
}

func (msg *MsgUpdateRollappInformation) ValidateBasic() error {
	rollapp := msg.GetUpdateRollappInformation()
	if err := rollapp.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

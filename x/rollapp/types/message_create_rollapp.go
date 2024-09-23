package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgCreateRollapp = "create_rollapp"

var _ sdk.Msg = &MsgCreateRollapp{}

func NewMsgCreateRollapp(
	creator,
	rollappId,
	initSequencer,
	alias string,
	vmType Rollapp_VMType,
	metadata *RollappMetadata,
	genesisInfo *GenesisInfo,
) *MsgCreateRollapp {
	return &MsgCreateRollapp{
		Creator:          creator,
		RollappId:        rollappId,
		InitialSequencer: initSequencer,
		Alias:            alias,
		VmType:           vmType,
		Metadata:         metadata,
		GenesisInfo:      genesisInfo,
	}
}

func (msg *MsgCreateRollapp) Route() string {
	return RouterKey
}

func (msg *MsgCreateRollapp) Type() string {
	return TypeMsgCreateRollapp
}

func (msg *MsgCreateRollapp) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateRollapp) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateRollapp) GetRollapp() Rollapp {
	return NewRollapp(
		msg.Creator,
		msg.RollappId,
		msg.InitialSequencer,
		msg.VmType,
		msg.Metadata,
		msg.GenesisInfo,
		false,
	)
}

func (msg *MsgCreateRollapp) ValidateBasic() error {
	if len(msg.Alias) == 0 {
		return ErrInvalidAlias
	}

	rollapp := msg.GetRollapp()
	if err := rollapp.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

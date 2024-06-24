package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgCreateRollapp = "create_rollapp"

var _ sdk.Msg = &MsgCreateRollapp{}

const MaxAllowedSequencers = 100

func NewMsgCreateRollapp(creator string, rollappId string, maxSequencers uint64, permissionedAddresses []string) *MsgCreateRollapp {
	return &MsgCreateRollapp{
		Creator:               creator,
		RollappId:             rollappId,
		MaxSequencers:         maxSequencers,
		PermissionedAddresses: permissionedAddresses,
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
	return NewRollapp(msg.Creator, msg.RollappId, msg.MaxSequencers, msg.PermissionedAddresses, false)
}

func (msg *MsgCreateRollapp) ValidateBasic() error {
	rollapp := msg.GetRollapp()
	if err := rollapp.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgTransferOwnership = "transfer_ownership"

var _ sdk.Msg = &MsgTransferOwnership{}

func NewMsgTransferOwnership(
	currentOwner,
	newOwner,
	rollappId string,
) *MsgTransferOwnership {
	return &MsgTransferOwnership{
		CurrentOwner: currentOwner,
		NewOwner:     newOwner,
		RollappId:    rollappId,
	}
}

func (msg *MsgTransferOwnership) Route() string {
	return RouterKey
}

func (msg *MsgTransferOwnership) Type() string {
	return TypeMsgTransferOwnership
}

func (msg *MsgTransferOwnership) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.CurrentOwner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferOwnership) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTransferOwnership) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.CurrentOwner); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.NewOwner); err != nil {
		return err
	}

	return nil
}

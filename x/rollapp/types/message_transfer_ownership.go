package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func (msg *MsgTransferOwnership) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.CurrentOwner); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.NewOwner); err != nil {
		return err
	}

	return nil
}

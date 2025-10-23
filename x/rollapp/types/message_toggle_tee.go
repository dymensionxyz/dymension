package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgToggleTEE{}

func NewMsgToggleTEE(
	owner,
	rollappId string,
	enable bool,
) *MsgToggleTEE {
	return &MsgToggleTEE{
		Owner:     owner,
		RollappId: rollappId,
		Enable:    enable,
	}
}

func (msg *MsgToggleTEE) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}

	return nil
}

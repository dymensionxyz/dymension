package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSubmitFraud = "submit_fraud"

var _ sdk.Msg = &MsgSubmitFraud{}

func NewMsgSubmitFraud(creator string, rollappID string) *MsgSubmitFraud {
  return &MsgSubmitFraud{
		Creator: creator,
    RollappID: rollappID,
	}
}

func (msg *MsgSubmitFraud) Route() string {
  return RouterKey
}

func (msg *MsgSubmitFraud) Type() string {
  return TypeMsgSubmitFraud
}

func (msg *MsgSubmitFraud) GetSigners() []sdk.AccAddress {
  creator, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{creator}
}

func (msg *MsgSubmitFraud) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitFraud) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Creator)
  	if err != nil {
  		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
  	}
  return nil
}


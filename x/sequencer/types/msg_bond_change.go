package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgIncreaseBond{}
	_ sdk.Msg = &MsgDecreaseBond{}
)

/* ---------------------------- MsgIncreaseBond ---------------------------- */
func NewMsgIncreaseBond(creator string, addAmount sdk.Coin) *MsgIncreaseBond {
	return &MsgIncreaseBond{
		Creator:   creator,
		AddAmount: addAmount,
	}
}

func (msg *MsgIncreaseBond) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if !(msg.AddAmount.IsValid() && msg.AddAmount.IsPositive()) {
		return errorsmod.Wrapf(ErrInvalidCoins, "invalid bond amount: %s", msg.AddAmount.String())
	}

	return nil
}

func (msg *MsgIncreaseBond) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

/* ---------------------------- MsgDecreaseBond ---------------------------- */
func NewMsgDecreaseBond(creator string, decreaseBond sdk.Coin) *MsgDecreaseBond {
	return &MsgDecreaseBond{
		Creator:        creator,
		DecreaseAmount: decreaseBond,
	}
}

func (msg *MsgDecreaseBond) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if !(msg.DecreaseAmount.IsValid() && msg.DecreaseAmount.IsPositive()) {
		return errorsmod.Wrapf(ErrInvalidCoins, "invalid bond amount: %s", msg.DecreaseAmount.String())
	}

	return nil
}

func (msg *MsgDecreaseBond) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

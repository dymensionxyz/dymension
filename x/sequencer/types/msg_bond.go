package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgIncreaseBond{}
	_ sdk.Msg = &MsgDecreaseBond{}
)

func NewMsgIncreaseBond(creator string, addAmount sdk.Coin) *MsgIncreaseBond {
	return &MsgIncreaseBond{
		Creator:   creator,
		AddAmount: addAmount,
	}
}

func (msg *MsgIncreaseBond) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddr, "invalid creator address (%s)", err)
	}

	if !msg.AddAmount.IsValid() || !msg.AddAmount.IsPositive() {
		return errorsmod.Wrapf(ErrInvalidCoins, "invalid bond amount: %s", msg.AddAmount.String())
	}

	return nil
}

func NewMsgDecreaseBond(creator string, decreaseBond sdk.Coin) *MsgDecreaseBond {
	return &MsgDecreaseBond{
		Creator:        creator,
		DecreaseAmount: decreaseBond,
	}
}

func (msg *MsgDecreaseBond) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddr, "invalid creator address (%s)", err)
	}

	if !msg.DecreaseAmount.IsValid() || !msg.DecreaseAmount.IsPositive() {
		return errorsmod.Wrapf(ErrInvalidCoins, "invalid bond amount: %s", msg.DecreaseAmount.String())
	}

	return nil
}

var _ sdk.Msg = &MsgUnbond{}

func NewMsgUnbond(creator string) *MsgUnbond {
	return &MsgUnbond{
		Creator: creator,
	}
}

func (msg *MsgUnbond) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddr, "invalid creator address (%s)", err)
	}

	return nil
}

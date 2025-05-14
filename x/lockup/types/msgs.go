package types

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgLockTokens{}
	_ sdk.Msg = &MsgBeginUnlocking{}
	_ sdk.Msg = &MsgExtendLockup{}
	_ sdk.Msg = &MsgForceUnlock{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgLockTokens creates a message to lock tokens.
func NewMsgLockTokens(owner sdk.AccAddress, duration time.Duration, coins sdk.Coins) *MsgLockTokens {
	return &MsgLockTokens{
		Owner:    owner.String(),
		Duration: duration,
		Coins:    coins,
	}
}

// NewMsgBeginUnlocking creates a message to begin unlocking the tokens of a specific lock.
func NewMsgBeginUnlocking(owner sdk.AccAddress, id uint64, coins sdk.Coins) *MsgBeginUnlocking {
	return &MsgBeginUnlocking{
		Owner: owner.String(),
		ID:    id,
		Coins: coins,
	}
}

// NewMsgExtendLockup creates a message to edit the properties of existing locks
func NewMsgExtendLockup(owner sdk.AccAddress, id uint64, duration time.Duration) *MsgExtendLockup {
	return &MsgExtendLockup{
		Owner:    owner.String(),
		ID:       id,
		Duration: duration,
	}
}

func (m MsgExtendLockup) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid owner address (%s)", err)
	}
	if m.ID == 0 {
		return fmt.Errorf("id is empty")
	}
	if m.Duration <= 0 {
		return fmt.Errorf("duration should be positive: %d < 0", m.Duration)
	}
	return nil
}

// NewMsgForceUnlock creates a message to begin unlocking tokens.
func NewMsgForceUnlock(owner sdk.AccAddress, id uint64, coins sdk.Coins) *MsgForceUnlock {
	return &MsgForceUnlock{
		Owner: owner.String(),
		ID:    id,
		Coins: coins,
	}
}

func (m MsgForceUnlock) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid owner address (%s)", err)
	}

	if m.ID <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "lock id should be bigger than 1 (%s)", err)
	}

	if !m.Coins.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Coins.String())
	}
	return nil
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "authority must be a valid bech32 address: %s", m.Authority),
		)
	}

	err = m.Params.ValidateBasic()
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidRequest,
			errorsmod.Wrapf(err, "failed to validate params"),
		)
	}

	return nil
}

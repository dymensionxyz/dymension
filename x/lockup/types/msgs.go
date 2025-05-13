package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgLockTokens     = "lock_tokens"
	TypeMsgBeginUnlocking = "begin_unlocking"
	TypeMsgExtendLockup   = "edit_lockup"
	TypeForceUnlock       = "force_unlock"
)

var (
	_ sdk.Msg = &MsgLockTokens{}
	_ sdk.Msg = &MsgBeginUnlocking{}
	_ sdk.Msg = &MsgExtendLockup{}
	_ sdk.Msg = &MsgForceUnlock{}
)

// NewMsgLockTokens creates a message to lock tokens.
func NewMsgLockTokens(owner sdk.AccAddress, duration time.Duration, coins sdk.Coins) *MsgLockTokens {
	return &MsgLockTokens{
		Owner:    owner.String(),
		Duration: duration,
		Coins:    coins,
	}
}

func (m MsgLockTokens) Route() string { return RouterKey }
func (m MsgLockTokens) Type() string  { return TypeMsgLockTokens }

func (m MsgLockTokens) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgBeginUnlocking{}

// NewMsgBeginUnlocking creates a message to begin unlocking the tokens of a specific lock.
func NewMsgBeginUnlocking(owner sdk.AccAddress, id uint64, coins sdk.Coins) *MsgBeginUnlocking {
	return &MsgBeginUnlocking{
		Owner: owner.String(),
		ID:    id,
		Coins: coins,
	}
}

func (m MsgBeginUnlocking) Route() string { return RouterKey }
func (m MsgBeginUnlocking) Type() string  { return TypeMsgBeginUnlocking }

func (m MsgBeginUnlocking) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

// NewMsgExtendLockup creates a message to edit the properties of existing locks
func NewMsgExtendLockup(owner sdk.AccAddress, id uint64, duration time.Duration) *MsgExtendLockup {
	return &MsgExtendLockup{
		Owner:    owner.String(),
		ID:       id,
		Duration: duration,
	}
}

func (m MsgExtendLockup) Route() string { return RouterKey }
func (m MsgExtendLockup) Type() string  { return TypeMsgExtendLockup }

func (m MsgExtendLockup) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgForceUnlock{}

// NewMsgForceUnlock creates a message to begin unlocking tokens.
func NewMsgForceUnlock(owner sdk.AccAddress, id uint64, coins sdk.Coins) *MsgForceUnlock {
	return &MsgForceUnlock{
		Owner: owner.String(),
		ID:    id,
		Coins: coins,
	}
}

func (m MsgForceUnlock) Route() string { return RouterKey }
func (m MsgForceUnlock) Type() string  { return TypeForceUnlock }
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

func (m MsgForceUnlock) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

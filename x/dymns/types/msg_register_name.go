package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgRegisterName{}

func (m *MsgRegisterName) ValidateBasic() error {
	if len(m.Name) > MaxDymNameLength {
		return ErrDymNameTooLong
	}

	if !dymnsutils.IsValidDymName(m.Name) {
		return ErrValidationFailed.Wrap("name is not a valid dym name")
	}

	if m.Duration < 1 {
		return ErrValidationFailed.Wrap("duration must be at least 1 year")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return ErrValidationFailed.Wrapf("owner is not a valid bech32 account address: %s", m.Owner)
	}

	if m.ConfirmPayment.IsNil() || m.ConfirmPayment.IsZero() {
		return ErrValidationFailed.Wrap("confirm payment is not set")
	} else if err := m.ConfirmPayment.Validate(); err != nil {
		return ErrValidationFailed.Wrapf("invalid confirm payment: %v", err)
	}

	if len(m.Contact) > MaxDymNameContactLength {
		return ErrValidationFailed.Wrapf("invalid contact length; got: %d, max: %d", len(m.Contact), MaxDymNameContactLength)
	}

	return nil
}

func (m *MsgRegisterName) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

func (m *MsgRegisterName) Route() string {
	return RouterKey
}

func (m *MsgRegisterName) Type() string {
	return TypeMsgRegisterName
}

func (m *MsgRegisterName) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

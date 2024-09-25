package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgCreatePlan{}
	_ sdk.Msg = &MsgBuy{}
	_ sdk.Msg = &MsgSell{}
	_ sdk.Msg = &MsgClaim{}
	_ sdk.Msg = &MsgUpdateParams{}
)

func (m *MsgCreatePlan) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", err)
	}

	if err := m.BondingCurve.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidBondingCurve, err)
	}

	allocationDec := ScaleXFromBase(m.AllocatedAmount, m.BondingCurve.SupplyDecimals())
	if !allocationDec.GT(MinTokenAllocation) {
		return ErrInvalidAllocation
	}

	if m.PreLaunchTime.Before(m.StartTime) {
		return ErrInvalidEndTime
	}

	if err := m.IncentivePlanParams.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidIncentivePlanParams, err)
	}

	return nil
}

func (m *MsgCreatePlan) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{addr}
}

func (m *MsgBuy) ValidateBasic() error {
	// buyer bech32
	_, err := sdk.AccAddressFromBech32(m.Buyer)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid buyer address: %s", err)
	}

	// coin exist and valid
	if m.Amount.IsNil() || !m.Amount.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrapf("amount %v must be positive", m.Amount)
	}

	if m.MaxCostAmount.IsNil() || !m.MaxCostAmount.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrapf("expected out amount %v must be positive", m.MaxCostAmount)
	}

	return nil
}

func (m *MsgBuy) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Buyer)
	return []sdk.AccAddress{addr}
}

func (m *MsgSell) ValidateBasic() error {
	// seller bech32
	_, err := sdk.AccAddressFromBech32(m.Seller)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid seller address: %s", err)
	}

	// coin exist and valid
	if m.Amount.IsNil() || !m.Amount.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrapf("amount %v must be positive", m.Amount)
	}

	if m.MinIncomeAmount.IsNil() || !m.MinIncomeAmount.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrapf("expected out amount %v must be positive", m.MinIncomeAmount)
	}

	return nil
}

func (m *MsgSell) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Seller)
	return []sdk.AccAddress{addr}
}

func (m *MsgClaim) ValidateBasic() error {
	// claimer bech32
	_, err := sdk.AccAddressFromBech32(m.Claimer)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid claimer address: %s", err)
	}

	return nil
}

func (m *MsgClaim) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Claimer)
	return []sdk.AccAddress{addr}
}

func (m *MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(
			"authority '%s' must be a valid bech32 address: %s",
			m.Authority, err.Error(),
		)
	}

	err = m.NewParams.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

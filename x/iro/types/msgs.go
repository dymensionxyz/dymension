package types

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreatePlan = "create_plan"
	TypeMsgBuy        = "buy"
	TypeMsgExactSpend = "buy_exact_spend"
	TypeMsgSell       = "sell"
	TypeMsgClaim      = "claim"
	TypeUpdateParams  = "update_params"
)

var (
	_ sdk.Msg = &MsgCreatePlan{}
	_ sdk.Msg = &MsgBuy{}
	_ sdk.Msg = &MsgBuyExactSpend{}
	_ sdk.Msg = &MsgSell{}
	_ sdk.Msg = &MsgClaim{}
	_ sdk.Msg = &MsgClaimVested{}
	_ sdk.Msg = &MsgEnableTrading{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// ValidateBasic performs basic validation checks on the MsgCreatePlan message.
// It ensures that the owner address is valid, the bonding curve is valid, the allocated amount
// is greater than the minimum token allocation, the pre-launch time is before the start time,
// and the incentive plan parameters are valid.
func (m *MsgCreatePlan) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", err)
	}

	if err := m.BondingCurve.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidBondingCurve, err)
	}

	allocationDec := ScaleFromBase(m.AllocatedAmount, m.BondingCurve.SupplyDecimals())
	if !allocationDec.GT(MinTokenAllocation) {
		return ErrInvalidAllocation
	}

	if m.IroPlanDuration < 0 {
		return ErrInvalidEndTime
	}

	// if start time set, trading must be enabled
	if m.StartTime.Unix() > 0 && !m.TradingEnabled {
		return errors.New("trading must be enabled to set start time")
	}

	if err := m.IncentivePlanParams.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidIncentivePlanParams, err)
	}

	if m.LiquidityPart.IsNegative() || m.LiquidityPart.GT(math.LegacyOneDec()) {
		return fmt.Errorf("liquidity part must be positive: %s", m.LiquidityPart)
	}

	if m.VestingDuration < 0 {
		return fmt.Errorf("vesting duration must be non-negative: %v", m.VestingDuration)
	}

	if m.VestingStartTimeAfterSettlement < 0 {
		return fmt.Errorf("vesting start time after settlement must be non-negative: %v", m.VestingStartTimeAfterSettlement)
	}

	if sdk.ValidateDenom(m.LiquidityDenom) != nil {
		return fmt.Errorf("invalid liquidity denom: %s", m.LiquidityDenom)
	}
	return nil
}

func (m *MsgCreatePlan) Route() string {
	return RouterKey
}

func (m *MsgBuy) Route() string {
	return RouterKey
}

func (m *MsgSell) Route() string {
	return RouterKey
}

func (m *MsgClaim) Route() string {
	return RouterKey
}

func (m *MsgBuyExactSpend) Route() string {
	return RouterKey
}

func (m *MsgUpdateParams) Route() string {
	return RouterKey
}

func (m *MsgCreatePlan) Type() string {
	return TypeMsgCreatePlan
}

func (m *MsgBuy) Type() string {
	return TypeMsgBuy
}

func (m *MsgSell) Type() string {
	return TypeMsgSell
}

func (m *MsgClaim) Type() string {
	return TypeMsgClaim
}

func (m *MsgBuyExactSpend) Type() string {
	return TypeMsgExactSpend
}

func (m *MsgUpdateParams) Type() string {
	return TypeUpdateParams
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

func (m *MsgClaimVested) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Claimer)
	return []sdk.AccAddress{addr}
}

// ValidateBasic implements types.Msg.
func (m *MsgClaimVested) ValidateBasic() error {
	// claimer bech32
	_, err := sdk.AccAddressFromBech32(m.Claimer)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid claimer address: %s", err)
	}

	return nil
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

// ValidateBasic implements types.Msg.
func (m *MsgBuyExactSpend) ValidateBasic() error {
	// buyer bech32
	_, err := sdk.AccAddressFromBech32(m.Buyer)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid buyer address: %s", err)
	}

	// coin exist and valid
	if m.Spend.IsNil() || !m.Spend.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrapf("amount %v must be positive", m.Spend)
	}

	if m.MinOutTokensAmount.IsNil() || !m.MinOutTokensAmount.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrapf("expected out amount %v must be positive", m.MinOutTokensAmount)
	}

	return nil
}

// GetSigners implements types.Msg.
func (m *MsgBuyExactSpend) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Buyer)
	return []sdk.AccAddress{addr}
}

// GetSigners implements types.Msg.
func (m *MsgEnableTrading) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{addr}
}

func (m *MsgEnableTrading) ValidateBasic() error {
	// owner bech32
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", err)
	}

	return nil
}

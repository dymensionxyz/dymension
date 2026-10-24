package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
)

func (m *MsgBuy) ValidateBasic() error {
	if m.MaxCostOrZero().IsNegative() {
		return errorsmod.Wrap(ErrInvalidPurchaseAmount, "max cost must be non-negative")
	}
	return nil
}

func (m *MsgBuyExactSpend) ValidateBasic() error {
	if m.MinTokensOutOrZero().IsNegative() {
		return errorsmod.Wrap(ErrInvalidPurchaseAmount, "min tokens out must be non-negative")
	}
	return nil
}

// MaxCostOrZero normalizes an unset bound from clients predating the field; zero disables the check.
func (m *MsgBuy) MaxCostOrZero() math.Int {
	return zeroIfNil(m.MaxCost)
}

// MinTokensOutOrZero normalizes an unset bound from clients predating the field; zero disables the check.
func (m *MsgBuyExactSpend) MinTokensOutOrZero() math.Int {
	return zeroIfNil(m.MinTokensOut)
}

func zeroIfNil(i math.Int) math.Int {
	if i.IsNil() {
		return math.ZeroInt()
	}
	return i
}

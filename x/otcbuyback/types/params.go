package types

import (
	"cosmossdk.io/math"
)

// DefaultParams returns the default parameters for the Otcbuyback module
func DefaultParams() Params {
	return Params{
		MovingAverageSmoothingFactor: math.LegacyNewDecWithPrec(1, 1), // 0.1
	}
}

// ValidateBasic performs basic validation on the Params
func (p Params) ValidateBasic() error {
	if p.MovingAverageSmoothingFactor.IsNegative() || p.MovingAverageSmoothingFactor.GT(math.LegacyOneDec()) {
		return ErrInvalidParams
	}
	return nil
}

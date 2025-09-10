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

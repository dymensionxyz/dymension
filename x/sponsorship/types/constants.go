package types

import (
	"cosmossdk.io/math"
)

var (
	// DYM represents 1 DYM. Equals to 10^18 aDYM.
	DYM = math.NewIntWithDecimal(1, 18)

	// MinAllocationWeight is a min weight one can allocate for the gauge.
	// Equals 1 and is measured in percentages.
	// 1 unit is 10^-18%.
	MinAllocationWeight = math.NewInt(1)
	// MaxAllocationWeight is a max weight one can allocate for the gauge.
	// Equals 100 * 10^18 and is measured in percentages.
	// 1 unit is 10^-18%. 100 * 10^18 is 100%.
	MaxAllocationWeight = DYM.MulRaw(100)

	DefaultMinAllocationWeight = MinAllocationWeight // 10^-18%
	DefaultMinVotingPower      = DYM                 // 1 DYM
)

package types

import (
	"cosmossdk.io/math"
)

var (
	// ADYM represents 1 ADYM.
	ADYM = math.NewInt(1)
	// DYM represents 1 DYM. Equals to 10^18 ADYM.
	DYM = math.NewIntWithDecimal(1, 18)

	// MinAllocationWeight is a min weight one can allocate for the gauge.
	// Equals 1 and is measured in percentages.
	// 1 unit is 10^-18%.
	MinAllocationWeight = ADYM
	// MaxAllocationWeight is a max weight one can allocate for the gauge.
	// Equals 100 * 10^18 and is measured in percentages.
	// 1 unit is 10^-18%. 100 * 10^18 is 100%.
	MaxAllocationWeight = DYM.MulRaw(100)

	DefaultMinAllocationWeight = DYM // 1%
	DefaultMinVotingPower      = DYM // 1 DYM
)

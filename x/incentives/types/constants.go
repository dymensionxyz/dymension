package types

import "cosmossdk.io/math"

var (
	// DYM represents 1 DYM
	DYM = math.NewIntWithDecimal(1, 18)

	DefaultCreateGaugeFee = DYM.MulRaw(10) // 10 DYM
	DefaultAddToGaugeFee  = math.ZeroInt() // 0 DYM
	DefaultAddDenomFee    = DYM            // 1 DYM
)

const DefaultDistrEpochIdentifier = "week"

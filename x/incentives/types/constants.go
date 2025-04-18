package types

import "cosmossdk.io/math"

var (
	// DYM represents 1 DYM
	DYM = math.NewIntWithDecimal(1, 18)

	DefaultCreateGaugeFee    = DYM.MulRaw(1)  // 1 DYM
	DefaultAddToGaugeFee     = math.ZeroInt() // 0 DYM
	DefaultAddDenomFee       = DYM            // 1 DYM
	DefaultRollappGaugesMode = Params_ActiveOnly
)

const DefaultDistrEpochIdentifier = "week"

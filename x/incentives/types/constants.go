package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
)

var (
	// DYM represents 1 DYM
	DYM = math.NewIntWithDecimal(1, 18)

	DefaultCreateGaugeFee    = DYM.MulRaw(1)                                                // 1 DYM
	DefaultAddToGaugeFee     = math.ZeroInt()                                               // 0 DYM
	DefaultAddDenomFee       = DYM                                                          // 1 DYM
	DefaultMinValueForDistr  = sdk.NewCoin(params.BaseDenom, math.NewIntWithDecimal(1, 16)) // 0.01 DYM
	DefaultRollappGaugesMode = Params_ActiveOnly
)

const DefaultDistrEpochIdentifier = "week"

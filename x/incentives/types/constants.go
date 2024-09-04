package types

import (
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// DYM represents 1 DYM
	DYM = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	DefaultDistrEpochIdentifier = "week"
	DefaultCreateGaugeFee       = DYM.Mul(sdk.NewInt(10))
	DefaultAddToGaugeFee        = DYM
)

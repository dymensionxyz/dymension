package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
)

var (
	// ADYM represents 1 ADYM.
	ADYM = math.NewInt(1)
	// DYM represents 1 DYM. Equals to 10^18 ADYM.
	DYM = math.NewIntWithDecimal(1, 18)
	// DYMCoin is 1 DYM coin.
	DYMCoin = sdk.NewCoin(params.BaseDenom, DYM)
)

// return DYM
func Dym(nDym math.Int) sdk.Coin {
	ret := DYMCoin
	ret.Amount = ret.Amount.Mul(nDym)
	return ret
}

// return ADYM
func ADym(nAdym math.Int) sdk.Coin {
	return sdk.NewCoin(params.BaseDenom, nAdym)
}

package common

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func StringToSdkInt(value string) sdk.Int {
	bigInt := big.NewInt(0)
	bigInt.SetString(value, 10)
	return sdk.NewIntFromBigInt(bigInt)
}

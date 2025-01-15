package ucoin

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MulDec multiplies each coin by dec, truncating down to the nearest integer.
func MulDec(dec sdk.Dec, coins ...sdk.Coin) sdk.Coins {
	ret := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		ret[i].Denom = coin.Denom
		ret[i].Amount = dec.MulInt(coin.Amount).TruncateInt()
	}
	return ret
}

func SimpleMul(c sdk.Coin, x int64) sdk.Coin {
	c.Amount = c.Amount.Mul(sdk.NewInt(x))
	return c
}

// SimpleMin returns the coin whos amt is less, a if equal
func SimpleMin(a sdk.Coin, b sdk.Coin) sdk.Coin {
	if a.Amount.LTE(b.Amount) {
		return a
	}
	return b
}

// SimpleMax returns the coin whos amt is greater, a if equal
func SimpleMax(a sdk.Coin, b sdk.Coin) sdk.Coin {
	if a.Amount.GTE(b.Amount) {
		return a
	}
	return b
}

// NonNegative makes the amt zero if negative
func NonNegative(c sdk.Coin) sdk.Coin {
	c.Amount = math.MaxInt(math.ZeroInt(), c.Amount)
	return c
}

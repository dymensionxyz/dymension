package ucoin

import sdk "github.com/cosmos/cosmos-sdk/types"

/*
TODO: migrate this package to dymension/sdk-utils
*/

// MulDec multiplies each coin by dec, truncating down to the nearest integer.
func MulDec(dec sdk.Dec, coins ...sdk.Coin) sdk.Coins {
	ret := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		ret[i].Denom = coin.Denom
		ret[i].Amount = dec.MulInt(coin.Amount).TruncateInt()
	}
	return ret
}

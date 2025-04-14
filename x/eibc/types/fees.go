package types

import (
	"math/big"

	"cosmossdk.io/math"
)

// calculate the new price: transferTotal - fee - bridgingFee. Ensures fulfiller does not lose due to bridge fee
func CalcPriceWithBridgingFee(amt math.Int, eibcFee math.Int, bridgeFeeMul math.LegacyDec) (math.Int, error) {
	bridgingFee := bridgeFeeMul.MulInt(amt).TruncateInt()
	price := amt.Sub(eibcFee).Sub(bridgingFee)
	// Check that the price is positive
	if !price.IsPositive() {
		return math.ZeroInt(), ErrFeeTooHigh
	}
	return price, nil
}

// returns an ibc-transfer amount sufficient to have a order price of target after fees (bridge + eibc)
// note that in the finalize without fulfillment case, the eibc fee is not applied, so the recipient will get approx target + eibcFee
// WARNING: not intended for on-chain code
func CalcTargetPriceAmt(target math.Int, eibcFee math.Int, bridgeFeeMul math.LegacyDec) math.Int {
	var ret math.Int

	l := target
	r := maxMathInt()

	for l.LTE(r) {
		delta := r.Sub(l).Quo(math.NewInt(2))
		mid := l.Add(delta)

		price, err := CalcPriceWithBridgingFee(mid, eibcFee, bridgeFeeMul)

		if err == nil && price.GTE(target) {
			ret = mid
			r = mid.Sub(math.OneInt())
		} else {
			l = mid.Add(math.OneInt())
		}
	}

	return ret
}

// (2^255 - 1)
func maxMathInt() math.Int {
	maxIntBig := new(big.Int)
	maxIntBig.Lsh(big.NewInt(1), 255)
	maxIntBig.Sub(maxIntBig, big.NewInt(1))
	return math.NewIntFromBigInt(maxIntBig)
}

package types

import (
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
//
// equation is:
// price = amt - eibcFee - floor(bridgeFeeMul*amt)
// solve for amt
func CalcTargetPriceAmt(target math.Int, eibcFee math.Int, bridgeFeeMul math.LegacyDec) math.Int {
	div := math.LegacyNewDec(1).Sub(bridgeFeeMul)

	mul := math.LegacyNewDec(1).Quo(div)

	amt := mul.MulInt(target.Add(eibcFee)).Ceil().TruncateInt()

	price, _ := CalcPriceWithBridgingFee(amt, eibcFee, bridgeFeeMul)

	if price.LT(target) {
		return amt.Add(math.OneInt())
	}

	return amt
}

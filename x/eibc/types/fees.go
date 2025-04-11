package types

import (
	"cosmossdk.io/math"
)

// calculate the new price: transferTotal - fee - bridgingFee. Ensures fulfiller does not lose due to bridge fee
func CalcPriceWithBridgingFee(amt math.Int, feeInt math.Int, bridgingFeeMultiplier math.LegacyDec) (math.Int, error) {
	bridgingFee := bridgingFeeMultiplier.MulInt(amt).TruncateInt()
	price := amt.Sub(feeInt).Sub(bridgingFee)
	// Check that the price is positive
	if !price.IsPositive() {
		return math.ZeroInt(), ErrFeeTooHigh
	}
	return price, nil
}

// returns an amount sufficient to have a price of target after fees
func SolvePrice(target math.Int, feeInt math.Int, bridgingFeeMultiplier math.LegacyDec) math.Int {
	div := math.LegacyNewDec(1).Sub(bridgingFeeMultiplier)

	mul := math.LegacyNewDec(1).Quo(div)

	return mul.MulInt(target.Add(feeInt)).TruncateInt()
}

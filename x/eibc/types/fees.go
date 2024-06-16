package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// calculate the new price: transferTotal - fee - bridgingFee
func CalcPriceWithBridgingFee(amt sdk.Int, feeInt sdk.Int, bridgingFeeMultiplier sdk.Dec) (sdk.Int, error) {
	bridgingFee := bridgingFeeMultiplier.MulInt(amt).TruncateInt()
	price := amt.Sub(feeInt).Sub(bridgingFee)
	// Check that the price is positive
	if !price.IsPositive() {
		return sdk.ZeroInt(), ErrFeeTooHigh
	}
	return price, nil
}

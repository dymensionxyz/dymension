package types

import (
	"time"

	"cosmossdk.io/math"
)

// NewPurchase creates a new user vesting plan
func NewPurchase(
	amount math.Int,
) Purchase {
	return Purchase{
		Amount:  amount,
		Claimed: math.ZeroInt(),
	}
}

// ClaimTokens updates the purchase data after a successful claim
func (v *Purchase) ClaimTokens(amount math.Int) {
	v.Claimed = v.Claimed.Add(amount)
}

// UnclaimedAmount returns the amount still vesting (not yet claimed)
func (v Purchase) UnclaimedAmount() math.Int {
	return v.Amount.Sub(v.Claimed)
}

// ClaimableAmount calculates the amount of tokens that have vested and are claimable
func (v Purchase) ClaimableAmount(currTime time.Time, startTime, endTime time.Time) math.Int {
	unclaimed := v.UnclaimedAmount()

	// no tokens to claim
	if !unclaimed.IsPositive() {
		return math.ZeroInt()
	}

	// not started
	if currTime.Before(startTime) {
		return math.ZeroInt()
	}

	// ended - all remaining tokens are claimable
	if currTime.After(endTime) {
		return unclaimed
	}

	// calculate the vesting scalar using linear vesting
	x := currTime.Sub(startTime)
	y := endTime.Sub(startTime)

	vestedAmt := v.Amount.MulRaw(x.Nanoseconds()).QuoRaw(y.Nanoseconds())
	claimable := vestedAmt.Sub(v.Claimed)

	// Ensure claimable is not negative
	if claimable.IsNegative() {
		return math.ZeroInt()
	}

	return claimable
}

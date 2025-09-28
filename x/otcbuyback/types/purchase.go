package types

import (
	"time"

	"cosmossdk.io/math"
)

// NewPurchase creates a new user vesting plan
func NewPurchase(
	amount math.Int,
	startTime, endTime time.Time,
) Purchase {
	return Purchase{
		Amount:    amount,
		Claimed:   math.ZeroInt(),
		StartTime: startTime,
		EndTime:   endTime,
	}
}

// ClaimTokens updates the purchase data after a successful claim
func (v *Purchase) ClaimTokens(amount math.Int) {
	v.Claimed = v.Claimed.Add(amount)
}

// GetRemainingVesting returns the amount still vesting (not yet claimed)
func (v Purchase) GetRemainingVesting() math.Int {
	return v.Amount.Sub(v.Claimed)
}

// VestedAmount calculates the amount of tokens that have vested and are claimable
func (v Purchase) VestedAmount(currTime time.Time) math.Int {
	unclaimed := v.GetRemainingVesting()

	// no tokens to claim
	if !unclaimed.IsPositive() {
		return math.ZeroInt()
	}

	// not started
	if currTime.Before(v.StartTime) {
		return math.ZeroInt()
	}

	// ended - all remaining tokens are claimable
	if currTime.After(v.EndTime) {
		return unclaimed
	}

	// calculate the vesting scalar using linear vesting
	x := currTime.Sub(v.StartTime)
	y := v.EndTime.Sub(v.StartTime)

	vestedAmt := v.Amount.MulRaw(x.Nanoseconds()).QuoRaw(y.Nanoseconds())
	claimable := vestedAmt.Sub(v.Claimed)

	// Ensure claimable is not negative
	if claimable.IsNegative() {
		return math.ZeroInt()
	}

	return claimable
}

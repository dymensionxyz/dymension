package types

import (
	"time"

	"cosmossdk.io/math"
)

// NewVestingPlan creates a new user vesting plan
func NewVestingPlan(
	amount math.Int,
	startTime time.Time,
	endTime time.Time,
) VestingPlan {
	return VestingPlan{
		Amount:    amount,
		Claimed:   math.ZeroInt(),
		StartTime: startTime,
		EndTime:   endTime,
	}
}

// ClaimTokens updates the vesting plan after a successful claim
func (v *VestingPlan) ClaimTokens(amount math.Int) {
	v.Claimed = v.Claimed.Add(amount)
}

// GetRemainingVesting returns the amount still vesting (not yet claimed)
func (v VestingPlan) GetRemainingVesting() math.Int {
	return v.Amount.Sub(v.Claimed)
}

// VestedAmount calculates the amount of tokens that have vested and are claimable
// This uses the same linear vesting calculation as the IRO module
func (v VestingPlan) VestedAmount(currTime time.Time) math.Int {
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
	s := math.LegacyNewDec(x.Nanoseconds()).Quo(math.LegacyNewDec(y.Nanoseconds()))

	vestedAmt := s.Mul(math.LegacyNewDecFromInt(v.Amount)).TruncateInt()
	claimable := vestedAmt.Sub(v.Claimed)

	// Ensure claimable is not negative
	if claimable.IsNegative() {
		return math.ZeroInt()
	}

	return claimable
}

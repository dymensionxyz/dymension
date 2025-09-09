package types

import (
	"errors"
	"time"

	"cosmossdk.io/math"
)

// ValidateBasic performs basic validation on the user vesting plan
func (v UserVestingPlan) ValidateBasic() error {
	if !v.Amount.IsPositive() {
		return errors.New("vesting amount must be positive")
	}

	if v.Amount.LT(v.Claimed) {
		return errors.New("amount cannot be less than claimed")
	}

	if !v.StartTime.IsZero() && !v.EndTime.IsZero() && v.StartTime.After(v.EndTime) {
		return errors.New("start time cannot be after end time")
	}

	return nil
}

// FIXME: review
// VestedAmount calculates the amount of tokens that have vested and are claimable
// This uses the same linear vesting calculation as the IRO module
func (v UserVestingPlan) VestedAmount(currTime time.Time) math.Int {
	unclaimed := v.Amount.Sub(v.Claimed)

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

// IsFullyVested returns true if all tokens have been vested (STUB)
func (v UserVestingPlan) IsFullyVested(currTime time.Time) bool {
	// TODO: Implement vesting logic with auction end time context
	return false
}

// IsVestingStarted returns true if vesting has started (STUB)
func (v UserVestingPlan) IsVestingStarted(currTime time.Time) bool {
	// TODO: Implement vesting logic with auction end time context
	return false
}

// GetVestingProgress returns the vesting progress as a decimal between 0 and 1 (STUB)
func (v UserVestingPlan) GetVestingProgress(currTime time.Time) math.LegacyDec {
	// TODO: Implement vesting progress calculation
	return math.LegacyZeroDec()
}

// NewUserVestingPlan creates a new user vesting plan
func NewUserVestingPlan(
	amount math.Int,
	startTime time.Time,
	endTime time.Time,
) UserVestingPlan {
	return UserVestingPlan{
		Amount:    amount,
		Claimed:   math.ZeroInt(),
		StartTime: startTime,
		EndTime:   endTime,
	}
}

// ClaimTokens updates the vesting plan after a successful claim
func (v *UserVestingPlan) ClaimTokens(amount math.Int) {
	v.Claimed = v.Claimed.Add(amount)
}

// GetRemainingVesting returns the amount still vesting (not yet claimed)
func (v UserVestingPlan) GetRemainingVesting() math.Int {
	return v.Amount.Sub(v.Claimed)
}

// ValidateBasic performs basic validation on the user vesting plan
func (v UserVestingPlan) ValidateBasic() error {
	if v.VestingDuration <= 0 {
		return errors.New("vesting duration must be positive")
	}

	if v.StartTime < 0 {
		return errors.New("start time cannot be negative")
	}

	return nil
}

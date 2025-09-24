package types

import (
	"errors"
	"time"

	math "cosmossdk.io/math"
)

func (v IROVestingPlan) ValidateBasic() error {
	if v.VestingDuration < 0 {
		return errors.New("vesting duration cannot be negative")
	}

	if v.StartTimeAfterSettlement < 0 {
		return errors.New("start time after settlement cannot be negative")
	}

	if v.Amount.IsNegative() {
		return errors.New("amount cannot be negative")
	}

	if v.Claimed.IsNegative() {
		return errors.New("claimed amount cannot be negative")
	}

	if v.Amount.LT(v.Claimed) {
		return errors.New("amount cannot be less than claimed")
	}

	// endtime must be greater than starttime
	if !v.Amount.IsZero() && v.EndTime.Compare(v.StartTime) <= 0 {
		return errors.New("end time must be greater than start time")
	}

	return nil
}

func (v IROVestingPlan) VestedAmt(currTime time.Time) math.Int {
	unclaimed := v.Amount.Sub(v.Claimed)

	// no tokens to claim
	if !unclaimed.IsPositive() {
		return math.ZeroInt()
	}

	// not started
	if currTime.Before(v.StartTime) {
		return math.ZeroInt()
	}

	// ended
	if currTime.After(v.EndTime) {
		return unclaimed
	}

	// calculate the vesting scalar
	x := currTime.Sub(v.StartTime)
	y := v.EndTime.Sub(v.StartTime)

	// handle edge case where vesting has zero duration
	if y == 0 {
		// if no vesting period, all unclaimed tokens are immediately available
		return unclaimed
	}

	s := math.LegacyNewDec(x.Nanoseconds()).Quo(math.LegacyNewDec(y.Nanoseconds()))

	vestedAmt := s.Mul(math.LegacyNewDecFromInt(v.Amount)).TruncateInt()
	claimable := vestedAmt.Sub(v.Claimed)
	return claimable
}

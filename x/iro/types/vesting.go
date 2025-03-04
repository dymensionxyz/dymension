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

	if v.StartTime.After(v.EndTime) {
		return errors.New("start time cannot be after end time")
	}

	if v.Amount.LT(v.Claimed) {
		return errors.New("amount cannot be less than claimed")
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
	s := math.LegacyNewDec(x.Nanoseconds()).Quo(math.LegacyNewDec(y.Nanoseconds()))

	vestedAmt := s.Mul(math.LegacyNewDecFromInt(v.Amount)).TruncateInt()
	claimable := vestedAmt.Sub(v.Claimed)
	return claimable
}

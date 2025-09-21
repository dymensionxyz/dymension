package types_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func TestGetVestedCoinsContVestingAcc(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)
	amount := math.NewInt(1_000_000).MulRaw(1e18)

	var v types.IROVestingPlan

	cases := []struct {
		name     string
		time     time.Time
		mallete  func()
		expected math.Int
	}{
		{
			"not started",
			startTime.Add(-1 * time.Hour),
			nil,
			math.ZeroInt(),
		},
		{
			"fully vested",
			endTime,
			nil,
			amount,
		},
		{
			"partially vested",
			startTime.Add(12 * time.Hour),
			nil,
			amount.QuoRaw(2),
		},
		{
			"empty",
			endTime,
			func() {
				v.Amount = math.ZeroInt()
			},
			math.ZeroInt(),
		},
		{
			"some claimed",
			startTime.Add(12 * time.Hour),
			func() {
				v.Claimed = math.NewInt(500)
			},
			amount.QuoRaw(2).SubRaw(500),
		},
		{
			"zero duration vesting - all immediately vested",
			startTime, // same as both start and end time
			func() {
				v.EndTime = v.StartTime // zero duration
			},
			amount, // should return full unclaimed amount
		},
	}

	for _, tc := range cases {
		v = types.IROVestingPlan{
			Amount:    amount,
			Claimed:   math.ZeroInt(),
			StartTime: startTime,
			EndTime:   endTime,
		}

		t.Run(tc.name, func(t *testing.T) {
			if tc.mallete != nil {
				tc.mallete()
			}
			require.Equal(t, tc.expected.String(), v.VestedAmt(tc.time).String())
		})
	}
}

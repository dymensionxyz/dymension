package types

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestAuction_GetDiscount_Linear(t *testing.T) {
	// Test setup - create base auction with known parameters
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	endTime := startTime.Add(10 * time.Hour)            // 10 hour auction
	initialDiscount := math.LegacyNewDecWithPrec(10, 2) // 0.10 = 10%
	maxDiscount := math.LegacyNewDecWithPrec(50, 2)     // 0.50 = 50%

	discountType := NewLinearDiscountType(
		initialDiscount,
		maxDiscount,
		24*time.Hour,
	)

	auction := NewAuction(
		1,                    // id
		math.NewInt(1000000), // allocation
		startTime,            // start time
		endTime,              // end time
		discountType,
		Auction_VestingParams{ // vesting params
			VestingDelay: time.Hour,
		},
		Auction_PumpParams{}, // pump params
	)

	tests := []struct {
		name             string
		currentTime      time.Time
		expectedDiscount math.LegacyDec
		description      string
	}{
		{
			name:             "before auction starts",
			currentTime:      startTime.Add(-1 * time.Hour),
			expectedDiscount: initialDiscount,
			description:      "Should return initial discount when auction hasn't started",
		},
		{
			name:             "exactly at start time",
			currentTime:      startTime,
			expectedDiscount: initialDiscount,
			description:      "Should return initial discount at exact start time",
		},
		{
			name:             "25% through auction (2.5 hours)",
			currentTime:      startTime.Add(2*time.Hour + 30*time.Minute),
			expectedDiscount: math.LegacyNewDecWithPrec(20, 2), // 0.20 = 20% (10% + 25% * 40%)
			description:      "Should calculate correct discount at 25% progress",
		},
		{
			name:             "50% through auction (5 hours)",
			currentTime:      startTime.Add(5 * time.Hour),
			expectedDiscount: math.LegacyNewDecWithPrec(30, 2), // 0.30 = 30% (10% + 50% * 40%)
			description:      "Should calculate correct discount at 50% progress",
		},
		{
			name:             "75% through auction (7.5 hours)",
			currentTime:      startTime.Add(7*time.Hour + 30*time.Minute),
			expectedDiscount: math.LegacyNewDecWithPrec(40, 2), // 0.40 = 40% (10% + 75% * 40%)
			description:      "Should calculate correct discount at 75% progress",
		},
		{
			name:             "exactly at end time",
			currentTime:      endTime,
			expectedDiscount: maxDiscount,
			description:      "Should return max discount at exact end time",
		},
		{
			name:             "after auction ends",
			currentTime:      endTime.Add(1 * time.Hour),
			expectedDiscount: maxDiscount,
			description:      "Should return max discount when auction has ended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualDiscount, _, err := auction.GetDiscount(tt.currentTime, 0)
			require.NoError(t, err, "unexpected error getting discount")
			require.True(t, tt.expectedDiscount.Equal(actualDiscount),
				"%s: expected discount %s, got %s", tt.description, tt.expectedDiscount.String(), actualDiscount.String())
		})
	}
}

func TestAuction_GetDiscount_Linear_ZeroDiscountRange(t *testing.T) {
	// Test edge case where initial and max discount are the same
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	startTime := baseTime
	endTime := baseTime.Add(5 * time.Hour)
	discount := math.LegacyNewDecWithPrec(25, 2) // 0.25 = 25%

	discountType := NewLinearDiscountType(
		discount, // same initial and max discount
		discount, // same initial and max discount
		24*time.Hour,
	)

	auction := NewAuction(
		1,
		math.NewInt(1000000),
		startTime,
		endTime,
		discountType,
		Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		Auction_PumpParams{},
	)

	// Test at different time points - should always return the same discount
	testTimes := []time.Time{
		startTime.Add(-1 * time.Hour), // before
		startTime,                     // start
		startTime.Add(2 * time.Hour),  // middle
		endTime,                       // end
		endTime.Add(1 * time.Hour),    // after
	}

	for _, testTime := range testTimes {
		actualDiscount, _, err := auction.GetDiscount(testTime, 0)
		require.NoError(t, err, "unexpected error getting discount")
		require.True(t, discount.Equal(actualDiscount),
			"Expected constant discount %s, got %s at time %s",
			discount.String(), actualDiscount.String(), testTime.String())
	}
}

func TestAuction_GetDiscount_Linear_ZeroDuration(t *testing.T) {
	// Test edge case where start time equals end time (zero duration)
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	startTime := baseTime
	endTime := baseTime                                 // Same as start time - zero duration
	initialDiscount := math.LegacyNewDecWithPrec(10, 2) // 0.10 = 10%
	maxDiscount := math.LegacyNewDecWithPrec(50, 2)     // 0.50 = 50%

	discountType := NewLinearDiscountType(
		initialDiscount,
		maxDiscount,
		24*time.Hour,
	)

	auction := NewAuction(
		1,
		math.NewInt(1000000),
		startTime,
		endTime, // Same as startTime
		discountType,
		Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		Auction_PumpParams{},
	)

	// Test at the exact time when start equals end
	// Should return max discount immediately since there's no time for progression
	actualDiscount, _, err := auction.GetDiscount(startTime, 0)
	require.NoError(t, err, "unexpected error getting discount")
	require.True(t, maxDiscount.Equal(actualDiscount),
		"Expected max discount %s for zero-duration auction, got %s",
		maxDiscount.String(), actualDiscount.String())
}

package types

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestPurchase_VestedAmount(t *testing.T) {
	// Test setup - create base purchase with known parameters
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	startTime := baseTime
	endTime := baseTime.Add(10 * time.Hour) // 10 hour vesting period
	totalAmount := math.NewInt(1000000)     // 1,000,000 tokens

	tests := []struct {
		name           string
		purchase       Purchase
		currentTime    time.Time
		expectedVested math.Int
		description    string
	}{
		{
			name: "before vesting starts",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.ZeroInt(),
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime.Add(-1 * time.Hour),
			expectedVested: math.ZeroInt(),
			description:    "Should return 0 when vesting hasn't started",
		},
		{
			name: "exactly at start time",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.ZeroInt(),
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime,
			expectedVested: math.ZeroInt(),
			description:    "Should return 0 at exact start time",
		},
		{
			name: "25% through vesting (2.5 hours)",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.ZeroInt(),
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime.Add(2*time.Hour + 30*time.Minute),
			expectedVested: math.NewInt(250000), // 25% of 1,000,000
			description:    "Should calculate correct vested amount at 25% progress",
		},
		{
			name: "50% through vesting (5 hours)",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.ZeroInt(),
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime.Add(5 * time.Hour),
			expectedVested: math.NewInt(500000), // 50% of 1,000,000
			description:    "Should calculate correct vested amount at 50% progress",
		},
		{
			name: "75% through vesting (7.5 hours)",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.ZeroInt(),
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime.Add(7*time.Hour + 30*time.Minute),
			expectedVested: math.NewInt(750000), // 75% of 1,000,000
			description:    "Should calculate correct vested amount at 75% progress",
		},
		{
			name: "exactly at end time",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.ZeroInt(),
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    endTime,
			expectedVested: totalAmount,
			description:    "Should return all unclaimed tokens at exact end time",
		},
		{
			name: "after vesting ends",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.ZeroInt(),
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    endTime.Add(1 * time.Hour),
			expectedVested: totalAmount,
			description:    "Should return all unclaimed tokens when vesting has ended",
		},
		{
			name: "50% through vesting with some tokens already claimed",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.NewInt(200000), // 200,000 already claimed
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime.Add(5 * time.Hour), // 50% through
			expectedVested: math.NewInt(300000),          // 500,000 total vested - 200,000 claimed = 300,000
			description:    "Should account for already claimed tokens",
		},
		{
			name: "all tokens already claimed",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   totalAmount, // All tokens claimed
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime.Add(5 * time.Hour),
			expectedVested: math.ZeroInt(),
			description:    "Should return 0 when all tokens are already claimed",
		},
		{
			name: "over-claimed scenario (should return 0)",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   totalAmount.Add(math.NewInt(100000)), // More than total amount
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime.Add(5 * time.Hour),
			expectedVested: math.ZeroInt(),
			description:    "Should return 0 when claimed exceeds total amount",
		},
		{
			name: "1 nanosecond after start",
			purchase: Purchase{
				Amount:    totalAmount,
				Claimed:   math.ZeroInt(),
				StartTime: startTime,
				EndTime:   endTime,
			},
			currentTime:    startTime.Add(1 * time.Nanosecond),
			expectedVested: math.ZeroInt(), // Should be extremely close to 0
			description:    "Should handle nanosecond precision after start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualVested := tt.purchase.VestedAmount(tt.currentTime)
			require.True(t, tt.expectedVested.Equal(actualVested),
				"%s: expected vested %s, got %s", tt.description, tt.expectedVested.String(), actualVested.String())
		})
	}
}

func TestPurchase_GetRemainingVesting(t *testing.T) {
	// Test the GetRemainingVesting helper function used by VestedAmount
	tests := []struct {
		name     string
		purchase Purchase
		expected math.Int
	}{
		{
			name: "no tokens claimed",
			purchase: Purchase{
				Amount:  math.NewInt(1000000),
				Claimed: math.ZeroInt(),
			},
			expected: math.NewInt(1000000),
		},
		{
			name: "some tokens claimed",
			purchase: Purchase{
				Amount:  math.NewInt(1000000),
				Claimed: math.NewInt(300000),
			},
			expected: math.NewInt(700000),
		},
		{
			name: "all tokens claimed",
			purchase: Purchase{
				Amount:  math.NewInt(1000000),
				Claimed: math.NewInt(1000000),
			},
			expected: math.ZeroInt(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.purchase.GetRemainingVesting()
			require.True(t, tt.expected.Equal(actual),
				"Expected %s, got %s", tt.expected.String(), actual.String())
		})
	}
}

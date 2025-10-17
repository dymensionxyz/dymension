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
	vestingPeriod := endTime.Sub(startTime)
	totalAmount := math.NewInt(1000000) // 1,000,000 tokens

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
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.ZeroInt(),
			},
			currentTime:    startTime.Add(-1 * time.Hour),
			expectedVested: math.ZeroInt(),
			description:    "Should return 0 when vesting hasn't started",
		},
		{
			name: "exactly at start time",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.ZeroInt(),
			},
			currentTime:    startTime,
			expectedVested: math.ZeroInt(),
			description:    "Should return 0 at exact start time",
		},
		{
			name: "25% through vesting (2.5 hours)",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.ZeroInt(),
			},
			currentTime:    startTime.Add(2*time.Hour + 30*time.Minute),
			expectedVested: math.NewInt(250000), // 25% of 1,000,000
			description:    "Should calculate correct vested amount at 25% progress",
		},
		{
			name: "50% through vesting (5 hours)",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.ZeroInt(),
			},
			currentTime:    startTime.Add(5 * time.Hour),
			expectedVested: math.NewInt(500000), // 50% of 1,000,000
			description:    "Should calculate correct vested amount at 50% progress",
		},
		{
			name: "75% through vesting (7.5 hours)",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.ZeroInt(),
			},
			currentTime:    startTime.Add(7*time.Hour + 30*time.Minute),
			expectedVested: math.NewInt(750000), // 75% of 1,000,000
			description:    "Should calculate correct vested amount at 75% progress",
		},
		{
			name: "exactly at end time",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.ZeroInt(),
			},
			currentTime:    endTime,
			expectedVested: totalAmount,
			description:    "Should return all unclaimed tokens at exact end time",
		},
		{
			name: "after vesting ends",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.ZeroInt(),
			},
			currentTime:    endTime.Add(1 * time.Hour),
			expectedVested: totalAmount,
			description:    "Should return all unclaimed tokens when vesting has ended",
		},
		{
			name: "50% through vesting with some tokens already claimed",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.NewInt(200000), // 200,000 already claimed

			},
			currentTime:    startTime.Add(5 * time.Hour), // 50% through
			expectedVested: math.NewInt(300000),          // 500,000 total vested - 200,000 claimed = 300,000
			description:    "Should account for already claimed tokens",
		},
		{
			name: "all tokens already claimed",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: totalAmount, // All tokens claimed

			},
			currentTime:    startTime.Add(5 * time.Hour),
			expectedVested: math.ZeroInt(),
			description:    "Should return 0 when all tokens are already claimed",
		},
		{
			name: "over-claimed scenario (should return 0)",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: totalAmount.Add(math.NewInt(100000)), // More than total amount

			},
			currentTime:    startTime.Add(5 * time.Hour),
			expectedVested: math.ZeroInt(),
			description:    "Should return 0 when claimed exceeds total amount",
		},
		{
			name: "1 nanosecond after start",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           totalAmount,
					VestingStartTime: startTime,
					VestingDuration:  vestingPeriod,
				}},
				Claimed: math.ZeroInt(),
			},
			currentTime:    startTime.Add(1 * time.Nanosecond),
			expectedVested: math.ZeroInt(), // Should be extremely close to 0
			description:    "Should handle nanosecond precision after start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualVested := tt.purchase.ClaimableAmount(tt.currentTime)
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
				Entries: []PurchaseEntry{{
					Amount:           math.NewInt(1000000),
					VestingStartTime: time.Time{},
					VestingDuration:  0,
				}},
				Claimed: math.ZeroInt(),
			},
			expected: math.NewInt(1000000),
		},
		{
			name: "some tokens claimed",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           math.NewInt(1000000),
					VestingStartTime: time.Time{},
					VestingDuration:  0,
				}},
				Claimed: math.NewInt(300000),
			},
			expected: math.NewInt(700000),
		},
		{
			name: "all tokens claimed",
			purchase: Purchase{
				Entries: []PurchaseEntry{{
					Amount:           math.NewInt(1000000),
					VestingStartTime: time.Time{},
					VestingDuration:  0,
				}},
				Claimed: math.NewInt(1000000),
			},
			expected: math.ZeroInt(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.purchase.UnclaimedAmount()
			require.True(t, tt.expected.Equal(actual),
				"Expected %s, got %s", tt.expected.String(), actual.String())
		})
	}
}

func TestPurchase_OverlappingVesting(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		entries           []PurchaseEntry
		claimed           math.Int
		testTime          time.Time
		expectedVested    math.Int
		expectedClaimable math.Int
	}{
		{
			name: "two entries, both not started",
			entries: []PurchaseEntry{
				{Amount: math.NewInt(1000), VestingStartTime: baseTime, VestingDuration: 10 * time.Hour},
				{Amount: math.NewInt(2000), VestingStartTime: baseTime.Add(2 * time.Hour), VestingDuration: 10 * time.Hour},
			},
			claimed:           math.ZeroInt(),
			testTime:          baseTime.Add(-1 * time.Hour),
			expectedVested:    math.ZeroInt(),
			expectedClaimable: math.ZeroInt(),
		},
		{
			name: "two entries, first partially vested, second not started",
			entries: []PurchaseEntry{
				{Amount: math.NewInt(1000), VestingStartTime: baseTime, VestingDuration: 10 * time.Hour},
				{Amount: math.NewInt(2000), VestingStartTime: baseTime.Add(2 * time.Hour), VestingDuration: 10 * time.Hour},
			},
			claimed:           math.ZeroInt(),
			testTime:          baseTime.Add(1 * time.Hour),
			expectedVested:    math.NewInt(100), // 10% of 1000 + 0% of 2000
			expectedClaimable: math.NewInt(100),
		},
		{
			name: "two entries, both partially vested",
			entries: []PurchaseEntry{
				{Amount: math.NewInt(1000), VestingStartTime: baseTime, VestingDuration: 10 * time.Hour},
				{Amount: math.NewInt(2000), VestingStartTime: baseTime.Add(2 * time.Hour), VestingDuration: 10 * time.Hour},
			},
			claimed:           math.ZeroInt(),
			testTime:          baseTime.Add(7 * time.Hour),
			expectedVested:    math.NewInt(1700), // 70% of 1000 + 50% of 2000
			expectedClaimable: math.NewInt(1700),
		},
		{
			name: "two entries, first fully vested, second partially",
			entries: []PurchaseEntry{
				{Amount: math.NewInt(1000), VestingStartTime: baseTime, VestingDuration: 10 * time.Hour},
				{Amount: math.NewInt(2000), VestingStartTime: baseTime.Add(2 * time.Hour), VestingDuration: 10 * time.Hour},
			},
			claimed:           math.ZeroInt(),
			testTime:          baseTime.Add(10 * time.Hour),
			expectedVested:    math.NewInt(2600), // 100% of 1000 + 80% of 2000
			expectedClaimable: math.NewInt(2600),
		},
		{
			name: "two entries, both fully vested",
			entries: []PurchaseEntry{
				{Amount: math.NewInt(1000), VestingStartTime: baseTime, VestingDuration: 10 * time.Hour},
				{Amount: math.NewInt(2000), VestingStartTime: baseTime.Add(2 * time.Hour), VestingDuration: 10 * time.Hour},
			},
			claimed:           math.ZeroInt(),
			testTime:          baseTime.Add(12 * time.Hour),
			expectedVested:    math.NewInt(3000), // 100% of 1000 + 100% of 2000
			expectedClaimable: math.NewInt(3000),
		},
		{
			name: "three entries with different start times and durations",
			entries: []PurchaseEntry{
				{Amount: math.NewInt(100), VestingStartTime: baseTime, VestingDuration: 4 * time.Hour},
				{Amount: math.NewInt(200), VestingStartTime: baseTime.Add(1 * time.Hour), VestingDuration: 5 * time.Hour},
				{Amount: math.NewInt(300), VestingStartTime: baseTime.Add(2 * time.Hour), VestingDuration: 10 * time.Hour},
			},
			claimed:           math.ZeroInt(),
			testTime:          baseTime.Add(5 * time.Hour),
			expectedVested:    math.NewInt(350), // 100% of 100 + 80% of 200 + 30% of 300
			expectedClaimable: math.NewInt(350),
		},
		{
			name: "overlapping with partial claims",
			entries: []PurchaseEntry{
				{Amount: math.NewInt(1000), VestingStartTime: baseTime, VestingDuration: 10 * time.Hour},
				{Amount: math.NewInt(2000), VestingStartTime: baseTime.Add(2 * time.Hour), VestingDuration: 10 * time.Hour},
			},
			claimed:           math.NewInt(500),
			testTime:          baseTime.Add(7 * time.Hour), // 70% of 1000 + 50% of 2000
			expectedVested:    math.NewInt(1700),
			expectedClaimable: math.NewInt(1200), // 1700 - 500 claimed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			purchase := Purchase{
				Entries: tt.entries,
				Claimed: tt.claimed,
			}

			unlocked := purchase.UnlockedAmount(tt.testTime)
			require.True(t, tt.expectedVested.Equal(unlocked),
				"Expected vested %s, got %s", tt.expectedVested.String(), unlocked.String())

			claimable := purchase.ClaimableAmount(tt.testTime)
			require.True(t, tt.expectedClaimable.Equal(claimable),
				"Expected claimable %s, got %s", tt.expectedClaimable.String(), claimable.String())
		})
	}
}

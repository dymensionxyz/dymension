package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

func TestGauge_IsActiveGauge(t *testing.T) {
	now := time.Now()
	oneHour := time.Hour
	zeroCoins := sdk.NewCoins()
	someCoins := sdk.NewCoins(sdk.NewCoin("testdenom", sdk.NewInt(100)))

	testCases := []struct {
		name        string
		gauge       Gauge
		curTime     time.Time
		expected    bool
		description string
	}{
		// --- EndorsementGauge Cases ---
		{
			name: "EndorsementGauge: Non-perpetual, active period, has coins",
			gauge: NewEndorsementGauge(1, false, "rollapp1", someCoins, now.Add(-oneHour), 10),
			curTime:     now,
			expected:    true,
			description: "Non-perpetual endorsement gauge, started, within duration (by coins), should be active",
		},
		{
			name: "EndorsementGauge: Non-perpetual, active period, zero coins",
			gauge: NewEndorsementGauge(1, false, "rollapp1", zeroCoins, now.Add(-oneHour), 10),
			curTime:     now,
			expected:    false,
			description: "Non-perpetual endorsement gauge, started, zero coins, should be inactive",
		},
		{
			name: "EndorsementGauge: Perpetual, active period, has coins",
			gauge: NewEndorsementGauge(1, true, "rollapp1", someCoins, now.Add(-oneHour), 0),
			curTime:     now,
			expected:    true,
			description: "Perpetual endorsement gauge, started, has coins, should be active",
		},
		{
			name: "EndorsementGauge: Perpetual, active period, zero coins",
			gauge: NewEndorsementGauge(1, true, "rollapp1", zeroCoins, now.Add(-oneHour), 0),
			curTime:     now,
			expected:    false,
			description: "Perpetual endorsement gauge, started, zero coins, should be inactive",
		},
		{
			name: "EndorsementGauge: Upcoming",
			gauge: NewEndorsementGauge(1, false, "rollapp1", someCoins, now.Add(oneHour), 10),
			curTime:     now,
			expected:    false,
			description: "Endorsement gauge, not yet started, should be inactive",
		},
		// --- Non-EndorsementGauge Cases (AssetGauge example) ---
		{
			name: "AssetGauge: Non-perpetual, active period, within epochs",
			gauge: NewAssetGauge(2, false, lockuptypes.QueryCondition{}, someCoins, now.Add(-oneHour), 10), // FilledEpochs = 0
			curTime:     now,
			expected:    true,
			description: "Non-perpetual asset gauge, started, epochs not filled, should be active",
		},
		{
			name: "AssetGauge: Non-perpetual, active period, epochs filled",
			gauge: func() Gauge {
				g := NewAssetGauge(2, false, lockuptypes.QueryCondition{}, someCoins, now.Add(-oneHour), 10)
				g.FilledEpochs = 10 // Manually set to filled
				return g
			}(),
			curTime:     now,
			expected:    false,
			description: "Non-perpetual asset gauge, started, epochs filled, should be inactive",
		},
		{
			name: "AssetGauge: Perpetual, active period",
			gauge: NewAssetGauge(2, true, lockuptypes.QueryCondition{}, someCoins, now.Add(-oneHour), 0),
			curTime:     now,
			expected:    true,
			description: "Perpetual asset gauge, started, should be active",
		},
		{
			name: "AssetGauge: Perpetual, zero coins (should still be active by current logic for non-endorsement)",
			gauge: NewAssetGauge(2, true, lockuptypes.QueryCondition{}, zeroCoins, now.Add(-oneHour), 0),
			curTime:     now,
			expected:    true,
			description: "Perpetual asset gauge, zero coins, should still be active as per logic for non-endorsement perpetuals",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.gauge.IsActiveGauge(tc.curTime), tc.description)
		})
	}
}

func TestGauge_IsFinishedGauge(t *testing.T) {
	now := time.Now()
	oneHour := time.Hour
	zeroCoins := sdk.NewCoins()
	someCoins := sdk.NewCoins(sdk.NewCoin("testdenom", sdk.NewInt(100)))

	testCases := []struct {
		name        string
		gauge       Gauge
		curTime     time.Time
		expected    bool
		description string
	}{
		// --- EndorsementGauge Cases ---
		{
			name: "EndorsementGauge: Non-perpetual, active period, has coins",
			gauge: NewEndorsementGauge(1, false, "rollapp1", someCoins, now.Add(-oneHour), 10),
			curTime:     now,
			expected:    false,
			description: "Non-perpetual endorsement gauge, started, has coins, should not be finished",
		},
		{
			name: "EndorsementGauge: Non-perpetual, active period, zero coins",
			gauge: NewEndorsementGauge(1, false, "rollapp1", zeroCoins, now.Add(-oneHour), 10),
			curTime:     now,
			expected:    true,
			description: "Non-perpetual endorsement gauge, started, zero coins, should be finished",
		},
		{
			name: "EndorsementGauge: Perpetual, active period, has coins",
			gauge: NewEndorsementGauge(1, true, "rollapp1", someCoins, now.Add(-oneHour), 0),
			curTime:     now,
			expected:    false, // Perpetual gauges are never finished by this method's criteria
			description: "Perpetual endorsement gauge, started, has coins, should not be finished",
		},
		{
			name: "EndorsementGauge: Perpetual, active period, zero coins",
			gauge: NewEndorsementGauge(1, true, "rollapp1", zeroCoins, now.Add(-oneHour), 0),
			curTime:     now,
			expected:    false, // Perpetual gauges are never finished by this method's criteria
			description: "Perpetual endorsement gauge, started, zero coins, should not be finished",
		},
		{
			name: "EndorsementGauge: Upcoming",
			gauge: NewEndorsementGauge(1, false, "rollapp1", someCoins, now.Add(oneHour), 10),
			curTime:     now,
			expected:    false,
			description: "Endorsement gauge, not yet started, should not be finished",
		},
		// --- Non-EndorsementGauge Cases (AssetGauge example) ---
		{
			name: "AssetGauge: Non-perpetual, active period, within epochs",
			gauge: NewAssetGauge(2, false, lockuptypes.QueryCondition{}, someCoins, now.Add(-oneHour), 10), // FilledEpochs = 0
			curTime:     now,
			expected:    false,
			description: "Non-perpetual asset gauge, started, epochs not filled, should not be finished",
		},
		{
			name: "AssetGauge: Non-perpetual, active period, epochs filled",
			gauge: func() Gauge {
				g := NewAssetGauge(2, false, lockuptypes.QueryCondition{}, someCoins, now.Add(-oneHour), 10)
				g.FilledEpochs = 10 // Manually set to filled
				return g
			}(),
			curTime:     now,
			expected:    true,
			description: "Non-perpetual asset gauge, started, epochs filled, should be finished",
		},
		{
			name: "AssetGauge: Perpetual, active period",
			gauge: NewAssetGauge(2, true, lockuptypes.QueryCondition{}, someCoins, now.Add(-oneHour), 0),
			curTime:     now,
			expected:    false, // Perpetual gauges are never finished
			description: "Perpetual asset gauge, started, should not be finished",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.gauge.IsFinishedGauge(tc.curTime), tc.description)
		})
	}
}

// TestGauge_NewEndorsementGauge_Defaults ensures that the EpochRewards (or accumulated_rewards if regenerated)
// field is properly initialized (e.g. as nil or empty sdk.Coins) since it's not explicitly set.
// This test is more about confirming default behavior.
func TestGauge_NewEndorsementGauge_Defaults(t *testing.T) {
	gauge := NewEndorsementGauge(1, false, "rollapp1", sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))), time.Now(), 10)

	endorsementGaugePart := gauge.GetEndorsement()
	require.NotNil(t, endorsementGaugePart, "Endorsement part of gauge should not be nil")

	// Due to stale gauge.pb.go, the field is EpochRewards. If it were regenerated, it would be AccumulatedRewards.
	// We are checking the default initialization, which should be empty or nil.
	// This test might need adjustment if the field name in Go struct changes after proto-gen.
	// For now, assuming direct access or a getter that reflects the current Go struct.
	// If EpochRewards is the current field name in the Go struct:
	if endorsementGaugePart.EpochRewards != nil {
		require.True(t, endorsementGaugePart.EpochRewards.Empty(), "EpochRewards should be empty by default")
	}
	// If AccumulatedRewards is the current field name in the Go struct (after a successful proto-gen):
	// if endorsementGaugePart.AccumulatedRewards != nil {
	//  require.True(t, endorsementGaugePart.AccumulatedRewards.Empty(), "AccumulatedRewards should be empty by default")
	// }

	// The actual field name is `EpochRewards` because `gauge.pb.go` is stale.
	// Accessing it directly:
	concreteEndorsementGauge, ok := gauge.DistributeTo.(*Gauge_Endorsement)
	require.True(t, ok)
	require.NotNil(t, concreteEndorsementGauge.Endorsement)
	// The field `EpochRewards` (or `AccumulatedRewards` post-regeneration) should be nil or empty.
	// Checking `concreteEndorsementGauge.Endorsement.EpochRewards`
	if concreteEndorsementGauge.Endorsement.EpochRewards != nil {
		// If it's not nil, it must be empty.
		require.True(t, concreteEndorsementGauge.Endorsement.EpochRewards.Empty(), "EndorsementGauge.EpochRewards should be empty by default if not nil")
	} else {
		// Or it can be nil, which is also acceptable as an empty list of coins.
		require.Nil(t, concreteEndorsementGauge.Endorsement.EpochRewards, "EndorsementGauge.EpochRewards should be nil or empty by default")
	}
}

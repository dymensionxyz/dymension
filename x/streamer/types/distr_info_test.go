package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func TestDistrInfoFromDistribution(t *testing.T) {
	testCases := []struct {
		name     string
		distr    sponsorshiptypes.Distribution
		expDistr *types.DistrInfo
	}{
		{
			name:  "Empty distribution",
			distr: sponsorshiptypes.NewDistribution(),
			expDistr: &types.DistrInfo{
				TotalWeight: sdk.NewInt(0),
				Records:     []types.DistrRecord{},
			},
		},
		{
			name: "Distribution with single gauge",
			distr: sponsorshiptypes.Distribution{
				VotingPower: sdk.NewInt(10),
				Gauges: []sponsorshiptypes.Gauge{
					{
						GaugeId: 1,
						Power:   sdk.NewInt(10),
					},
				},
			},
			expDistr: &types.DistrInfo{
				TotalWeight: sdk.NewInt(10),
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  sdk.NewInt(10),
					},
				},
			},
		},
		{
			name: "Distribution with multiple gauges",
			distr: sponsorshiptypes.Distribution{
				VotingPower: sdk.NewInt(30),
				Gauges: []sponsorshiptypes.Gauge{
					{
						GaugeId: 1,
						Power:   sdk.NewInt(10),
					},
					{
						GaugeId: 2,
						Power:   sdk.NewInt(20),
					},
				},
			},
			expDistr: &types.DistrInfo{
				TotalWeight: sdk.NewInt(30),
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  sdk.NewInt(10),
					},
					{
						GaugeId: 2,
						Weight:  sdk.NewInt(20),
					},
				},
			},
		},
		{
			name: "Distribution with abstained gauge",
			distr: sponsorshiptypes.Distribution{
				VotingPower: sdk.NewInt(100),
				Gauges: []sponsorshiptypes.Gauge{
					// 30 is abstained
					{
						GaugeId: 1,
						Power:   sdk.NewInt(50),
					},
					{
						GaugeId: 2,
						Power:   sdk.NewInt(20),
					},
				},
			},
			expDistr: &types.DistrInfo{
				TotalWeight: sdk.NewInt(100),
				Records: []types.DistrRecord{
					// 30 is abstained
					{
						GaugeId: 1,
						Weight:  sdk.NewInt(50),
					},
					{
						GaugeId: 2,
						Weight:  sdk.NewInt(20),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			distr := types.DistrInfoFromDistribution(tc.distr)
			assert.Equal(t, tc.expDistr, distr)
		})
	}
}

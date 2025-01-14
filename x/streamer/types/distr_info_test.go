package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func TestDistrInfoFromDistribution(t *testing.T) {
	testCases := []struct {
		name     string
		distr    sponsorshiptypes.Distribution
		expDistr types.DistrInfo
	}{
		{
			name:  "Empty distribution",
			distr: sponsorshiptypes.NewDistribution(),
			expDistr: types.DistrInfo{
				TotalWeight: math.NewInt(0),
				Records:     []types.DistrRecord{},
			},
		},
		{
			name: "Distribution with single gauge",
			distr: sponsorshiptypes.Distribution{
				VotingPower: math.NewInt(10),
				Gauges: []sponsorshiptypes.Gauge{
					{
						GaugeId: 1,
						Power:   math.NewInt(10),
					},
				},
			},
			expDistr: types.DistrInfo{
				TotalWeight: math.NewInt(10),
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  math.NewInt(10),
					},
				},
			},
		},
		{
			name: "Distribution with multiple gauges",
			distr: sponsorshiptypes.Distribution{
				VotingPower: math.NewInt(30),
				Gauges: []sponsorshiptypes.Gauge{
					{
						GaugeId: 1,
						Power:   math.NewInt(10),
					},
					{
						GaugeId: 2,
						Power:   math.NewInt(20),
					},
				},
			},
			expDistr: types.DistrInfo{
				TotalWeight: math.NewInt(30),
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  math.NewInt(10),
					},
					{
						GaugeId: 2,
						Weight:  math.NewInt(20),
					},
				},
			},
		},
		{
			name: "Distribution with empty gauges",
			distr: sponsorshiptypes.Distribution{
				VotingPower: math.NewInt(30),
				Gauges:      []sponsorshiptypes.Gauge{},
			},
			expDistr: types.DistrInfo{
				TotalWeight: sdk.ZeroInt(),
				Records:     []types.DistrRecord{},
			},
		},
		{
			name: "Distribution with abstained gauge",
			distr: sponsorshiptypes.Distribution{
				VotingPower: math.NewInt(100),
				Gauges: []sponsorshiptypes.Gauge{
					// 30 is abstained
					{
						GaugeId: 1,
						Power:   math.NewInt(50),
					},
					{
						GaugeId: 2,
						Power:   math.NewInt(20),
					},
				},
			},
			expDistr: types.DistrInfo{
				TotalWeight: math.NewInt(70),
				Records: []types.DistrRecord{
					// 30 is abstained
					{
						GaugeId: 1,
						Weight:  math.NewInt(50),
					},
					{
						GaugeId: 2,
						Weight:  math.NewInt(20),
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

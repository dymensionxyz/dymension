package keeper_test

import (
	"cosmossdk.io/math"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (s *KeeperTestSuite) TestProcessEpochPointer() {
	tests := []struct {
		name            string
		iterationsLimit int
		numGauges       int
		pointer         types.EpochPointer
		streams         []types.Stream
		expectedError   error
	}{
		{
			name:            "General case",
			iterationsLimit: 100,
			numGauges:       30,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				{
					Id: 1,
					DistributeTo: &types.DistrInfo{
						TotalWeight: math.NewInt(100),
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(3)},
							{GaugeId: 2, Weight: math.NewInt(3)},
							{GaugeId: 3, Weight: math.NewInt(3)},
						},
					},
					DistrEpochIdentifier: "day",
				},
				{
					Id: 2,
					DistributeTo: &types.DistrInfo{
						TotalWeight: math.NewInt(100),
						Records: []types.DistrRecord{
							{GaugeId: 2, Weight: math.NewInt(3)},
							{GaugeId: 3, Weight: math.NewInt(3)},
							{GaugeId: 4, Weight: math.NewInt(3)},
						},
					},
					DistrEpochIdentifier: "hour",
				},
				{
					Id: 3,
					DistributeTo: &types.DistrInfo{
						TotalWeight: math.NewInt(100),
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(3)},
							{GaugeId: 5, Weight: math.NewInt(3)},
							{GaugeId: 6, Weight: math.NewInt(3)},
						},
					},
					DistrEpochIdentifier: "hour",
				},
				{
					Id: 4,
					DistributeTo: &types.DistrInfo{
						TotalWeight: math.NewInt(100),
						Records: []types.DistrRecord{
							{GaugeId: 2, Weight: math.NewInt(3)},
							{GaugeId: 5, Weight: math.NewInt(3)},
							{GaugeId: 7, Weight: math.NewInt(3)},
						},
					},
					DistrEpochIdentifier: "day",
				},
			},
			expectedError: nil,
		},
	}

	// Run tests
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.CreateGauges(tc.numGauges)

			err := s.App.StreamerKeeper.EndBlock(s.Ctx)
			_ = err

			s.T().Log(err)
		})
	}
}

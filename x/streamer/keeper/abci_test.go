package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/internal/pagination"
	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func TestStreamIterator(t *testing.T) {
	// newStream is a helper function
	newStream := func(id uint64, epochID string, gaugeIDs ...uint64) types.Stream {
		g := make([]types.DistrRecord, 0, len(gaugeIDs))
		for _, gID := range gaugeIDs {
			g = append(g, types.DistrRecord{GaugeId: gID})
		}
		return types.Stream{
			Id:                   id,
			DistributeTo:         &types.DistrInfo{Records: g},
			DistrEpochIdentifier: epochID,
		}
	}

	tests := []struct {
		name              string
		maxIters          uint64
		pointer           types.EpochPointer
		streams           []types.Stream
		expectedIters     uint64
		expectedTraversal [][2]uint64 // holds an expected stream slice traversal. [2]uint64 is a pair of streamID and gaugeID.
		expectedPointer   types.EpochPointer
	}{
		{
			name:     "Pointer at the very beginning 1",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour", 1, 5, 6),
				newStream(4, "day", 2, 5, 7),
			},
			expectedIters: 6,
			expectedTraversal: [][2]uint64{
				{1, 1},
				{1, 2},
				{1, 3},
				{4, 2},
				{4, 5},
				{4, 7},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Pointer at the very beginning 2",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "hour",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour", 1, 5, 6),
				newStream(4, "day", 2, 5, 7),
			},
			expectedIters: 6,
			expectedTraversal: [][2]uint64{
				{2, 2},
				{2, 3},
				{2, 4},
				{3, 1},
				{3, 5},
				{3, 6},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "hour",
			},
		},
		{
			name:     "Pointer at the very end 1",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour", 1, 5, 6),
				newStream(4, "day", 2, 5, 7),
			},
			expectedIters:     0,
			expectedTraversal: nil,
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Pointer at the very end 2",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "hour",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour", 1, 5, 6),
				newStream(4, "day", 2, 5, 7),
			},
			expectedIters:     0,
			expectedTraversal: nil,
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "hour",
			},
		},
		{
			name:     "Empty stream 1",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "hour",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour"),
				newStream(4, "hour", 1, 5, 6),
				newStream(5, "day", 2, 5, 7),
			},
			expectedIters:     0,
			expectedTraversal: nil,
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "hour",
			},
		},
		{
			name:     "Empty stream 2",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "hour",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour"),
				newStream(4, "hour", 1, 5, 6),
				newStream(5, "day", 2, 5, 7),
			},
			expectedIters: 6,
			expectedTraversal: [][2]uint64{
				{2, 2},
				{2, 3},
				{2, 4},
				{4, 1},
				{4, 5},
				{4, 6},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "hour",
			},
		},
		{
			name:     "Empty stream 3: the last stream is empty",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour"),
				newStream(4, "hour", 1, 5, 6),
				newStream(5, "day"),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{1, 1},
				{1, 2},
				{1, 3},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Empty stream 4: the first stream is empty",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day"),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour"),
				newStream(4, "hour", 1, 5, 6),
				newStream(5, "day", 1, 2, 3),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{5, 1},
				{5, 2},
				{5, 3},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Pointer stops at the middle gauge the stream",
			maxIters: 4,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour", 1, 5, 6),
				newStream(4, "day", 2, 5, 7),
			},
			expectedIters: 4,
			expectedTraversal: [][2]uint64{
				{1, 1},
				{1, 2},
				{1, 3},
				{4, 2},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        4,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Pointer stops at the last gauge of the stream",
			maxIters: 3,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 2, 3, 4),
				newStream(3, "hour", 1, 5, 6),
				newStream(4, "day", 2, 5, 7),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{1, 1},
				{1, 2},
				{1, 3},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        4,
				GaugeId:         2,
				EpochIdentifier: "day",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var traversal [][2]uint64
			newPointer, iters := keeper.IterateEpochPointer(tc.pointer, tc.streams, tc.maxIters, func(v keeper.StreamGauge) pagination.Stop {
				// TODO: verify traversal log
				t.Logf("stream %d   gauge %d", v.Stream.Id, v.Gauge.GaugeId)
				traversal = append(traversal, [2]uint64{v.Stream.Id, v.Gauge.GaugeId})
				return pagination.Continue
			})

			require.Equal(t, tc.expectedIters, iters)
			require.Equal(t, tc.expectedTraversal, traversal)
			require.Equal(t, tc.expectedPointer, newPointer)
		})
	}
}

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
				EpochCoins:      sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100))),
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

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func TestStreamIterator(t *testing.T) {
	t.Parallel()

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
			name:     "No streams",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams:           []types.Stream{},
			expectedIters:     0,
			expectedTraversal: nil,
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
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
			name:     "One relevant empty stream",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day"),
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
			name:     "One irrelevant non-empty stream",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
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
			name:     "One irrelevant empty stream",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour"),
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
			name:     "Two streams: 1st is relevant",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour", 1, 2, 3),
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
			name:     "Two streams: 2nd is relevant",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
				newStream(2, "day", 1, 2, 3),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{2, 1},
				{2, 2},
				{2, 3},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Two streams: 1, 2 relevant",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "day", 1, 2, 3),
			},
			expectedIters: 6,
			expectedTraversal: [][2]uint64{
				{1, 1},
				{1, 2},
				{1, 3},
				{2, 1},
				{2, 2},
				{2, 3},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Two streams: none relevant",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
				newStream(2, "hour", 1, 2, 3),
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
			name:     "Two streams: 1st is relevant but empty",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day"),
				newStream(2, "hour", 1, 2, 3),
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
			name:     "Two streams: 2nd is relevant but empty",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
				newStream(2, "day"),
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
			name:     "Two streams: 1, 2 relevant but empty",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day"),
				newStream(2, "day"),
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
			name:     "Two streams: 1st is relevant, irrelevant is empty",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
				newStream(2, "hour"),
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
			name:     "Two streams: 2nd is relevant, irrelevant is empty",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour"),
				newStream(2, "day", 1, 2, 3),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{2, 1},
				{2, 2},
				{2, 3},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Two streams: both irrelevant are empty",
			maxIters: 100,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour"),
				newStream(2, "hour"),
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
		// All further cases intend to test with limited iterations
		{
			name:     "One relevant stream, 0 iterations",
			maxIters: 0,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
			},
			expectedIters:     0,
			expectedTraversal: nil,
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         1,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
			},
			expectedIters: 1,
			expectedTraversal: [][2]uint64{
				{1, 1},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         2,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream, 2 iterations",
			maxIters: 2,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
			},
			expectedIters: 2,
			expectedTraversal: [][2]uint64{
				{1, 1},
				{1, 2},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         3,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream, iterations equal to num of gauges",
			maxIters: 3,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
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
			name:     "One relevant stream, iterations is greater than num of gauges",
			maxIters: 4,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 2, 3),
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
			name:     "One irrelevant stream, 0 iterations",
			maxIters: 0,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
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
			name:     "One irrelevant stream, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
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
			name:     "One irrelevant stream, 2 iterations",
			maxIters: 2,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
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
			name:     "One irrelevant stream, iterations equal to num of gauges",
			maxIters: 3,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
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
			name:     "One irrelevant stream, iterations is greater than num of gauges",
			maxIters: 4,
			pointer: types.EpochPointer{
				StreamId:        0,
				GaugeId:         0,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 2, 3),
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
			name:     "One relevant stream, start from the valid gauge, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
			},
			expectedIters: 1,
			expectedTraversal: [][2]uint64{
				{1, 5},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         8,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream, start from the valid gauge, 2 iterations",
			maxIters: 2,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
			},
			expectedIters: 2,
			expectedTraversal: [][2]uint64{
				{1, 5},
				{1, 8},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         12,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream, start from the valid gauge, iteration equal to num of gauges",
			maxIters: 3,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{1, 5},
				{1, 8},
				{1, 12},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream, start from the valid gauge, iteration is greater than num of gauges",
			maxIters: 4,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{1, 5},
				{1, 8},
				{1, 12},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream, start from the invalid gauge, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         3, // invalid
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
			},
			expectedIters: 1,
			expectedTraversal: [][2]uint64{
				{1, 5},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         8,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One relevant stream, start from the invalid gauge, 2 iterations",
			maxIters: 2,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         3, // invalid
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
			},
			expectedIters: 2,
			expectedTraversal: [][2]uint64{
				{1, 5},
				{1, 8},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         12,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "One irrelevant stream, start from the valid gauge, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 5, 8, 12),
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
			name:     "One irrelevant stream, start from the valid gauge, 2 iterations",
			maxIters: 2,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 5, 8, 12),
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
			name:     "One irrelevant stream, start from the valid gauge, iteration equal to num of gauges",
			maxIters: 3,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 5, 8, 12),
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
			name:     "One irrelevant stream, start from the valid gauge, iteration is greater than num of gauges",
			maxIters: 4,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 5, 8, 12),
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
			name:     "One irrelevant stream, start from the invalid gauge, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         3, // invalid
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 5, 8, 12),
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
			name:     "One irrelevant stream, start from the invalid gauge, 2 iterations",
			maxIters: 2,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         3, // invalid
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "hour", 1, 5, 8, 12),
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
			name:     "Three streams: 1, 3 are relevant, start from the valid gauge, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "hour", 1, 5, 8, 12),
				newStream(3, "day", 1, 5, 8, 12),
			},
			expectedIters: 1,
			expectedTraversal: [][2]uint64{
				{1, 5},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         8,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Three streams: 1, 3 are relevant, start from the invalid gauge, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         6,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "hour", 1, 5, 8, 12),
				newStream(3, "day", 1, 5, 8, 12),
			},
			expectedIters: 1,
			expectedTraversal: [][2]uint64{
				{1, 8},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         12,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Three streams: 1, 3 are relevant, start from the invalid gauge, stop at the last element of the stream",
			maxIters: 2,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         6,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "hour", 1, 5, 8, 12),
				newStream(3, "day", 1, 5, 8, 12),
			},
			expectedIters: 2,
			expectedTraversal: [][2]uint64{
				{1, 8},
				{1, 12},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        3,
				GaugeId:         1,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Three streams: 1, 3 are relevant, start from the invalid gauge, stop at the next stream",
			maxIters: 3,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         6,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "hour", 1, 5, 8, 12),
				newStream(3, "day", 1, 5, 8, 12),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{1, 8},
				{1, 12},
				{3, 1},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        3,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Three streams: 1, 3 are relevant, start from the invalid gauge, more iterations than gauges",
			maxIters: 6,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         6,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "hour", 1, 5, 8, 12),
				newStream(3, "day", 1, 5, 8, 12),
			},
			expectedIters: 6,
			expectedTraversal: [][2]uint64{
				{1, 8},
				{1, 12},
				{3, 1},
				{3, 5},
				{3, 8},
				{3, 12},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Three streams: 1, 3 are relevant, start from the invalid gauge, 3 is empty",
			maxIters: 6,
			pointer: types.EpochPointer{
				StreamId:        1,
				GaugeId:         6,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "hour", 1, 5, 8, 12),
				newStream(3, "day"),
			},
			expectedIters: 2,
			expectedTraversal: [][2]uint64{
				{1, 8},
				{1, 12},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        types.MaxStreamID,
				GaugeId:         types.MaxGaugeID,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Three streams: 1, 3 are relevant, start from the invalid stream, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        2,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "hour", 1, 5, 8, 12),
				newStream(3, "day", 1, 5, 8, 12),
			},
			expectedIters: 1,
			expectedTraversal: [][2]uint64{
				{3, 1},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        3,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Three streams: 1, 3 are relevant, start from the invalid stream, 3 iterations",
			maxIters: 3,
			pointer: types.EpochPointer{
				StreamId:        2,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "hour", 1, 5, 8, 12),
				newStream(3, "day", 1, 5, 8, 12),
			},
			expectedIters: 3,
			expectedTraversal: [][2]uint64{
				{3, 1},
				{3, 5},
				{3, 8},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        3,
				GaugeId:         12,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Three streams: 1, 2, 3 are relevant, start from the relevant but empty stream, 1 iteration",
			maxIters: 1,
			pointer: types.EpochPointer{
				StreamId:        2,
				GaugeId:         1,
				EpochIdentifier: "day",
			},
			streams: []types.Stream{
				newStream(1, "day", 1, 5, 8, 12),
				newStream(2, "day"),
				newStream(3, "day", 1, 5, 8, 12),
			},
			expectedIters: 1,
			expectedTraversal: [][2]uint64{
				{3, 1},
			},
			expectedPointer: types.EpochPointer{
				StreamId:        3,
				GaugeId:         5,
				EpochIdentifier: "day",
			},
		},
		{
			name:     "Pointer stops at the middle gauge of the stream",
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
			t.Parallel()

			var traversal [][2]uint64
			newPointer, iters := keeper.IterateEpochPointer(tc.pointer, tc.streams, tc.maxIters, func(v keeper.StreamGauge) (stop bool, weight uint64) {
				traversal = append(traversal, [2]uint64{v.Stream.Id, v.Gauge.GaugeId})
				return false, 1
			})

			require.Equal(t, tc.expectedIters, iters)
			require.Equal(t, tc.expectedTraversal, traversal)
			require.Equal(t, tc.expectedPointer, newPointer)
		})
	}
}

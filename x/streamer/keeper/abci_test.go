package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (s *KeeperTestSuite) TestProcessEpochPointer() {
	tests := []struct {
		name                  string
		maxIterationsPerBlock uint64
		numGauges             int
		blocksInEpoch         int
		streams               []types.Stream
		expectedBlockResults  []blockResults
	}{
		{
			name:                  "1 block in the epoch",
			maxIterationsPerBlock: 9,
			numGauges:             16,
			blocksInEpoch:         1,
			streams: []types.Stream{
				{
					Id:                   1,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(25)},
							{GaugeId: 2, Weight: math.NewInt(25)},
							{GaugeId: 3, Weight: math.NewInt(25)},
							{GaugeId: 4, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   2,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 5, Weight: math.NewInt(25)},
							{GaugeId: 6, Weight: math.NewInt(25)},
							{GaugeId: 7, Weight: math.NewInt(25)},
							{GaugeId: 8, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   3,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 9, Weight: math.NewInt(25)},
							{GaugeId: 10, Weight: math.NewInt(25)},
							{GaugeId: 11, Weight: math.NewInt(25)},
							{GaugeId: 12, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   4,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 13, Weight: math.NewInt(25)},
							{GaugeId: 14, Weight: math.NewInt(25)},
							{GaugeId: 15, Weight: math.NewInt(25)},
							{GaugeId: 16, Weight: math.NewInt(25)},
						},
					},
				},
			},
			expectedBlockResults: []blockResults{
				{
					height: 0,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxStreamID,
							EpochIdentifier: "day",
						},
						{
							StreamId:        2,
							GaugeId:         6,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 6, coins: nil},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
					},
				},
			},
		},
		{
			name:                  "Several blocks in the epoch",
			maxIterationsPerBlock: 5,
			numGauges:             16,
			blocksInEpoch:         2,
			streams: []types.Stream{
				{
					Id:                   1,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(25)},
							{GaugeId: 2, Weight: math.NewInt(25)},
							{GaugeId: 3, Weight: math.NewInt(25)},
							{GaugeId: 4, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   2,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 5, Weight: math.NewInt(25)},
							{GaugeId: 6, Weight: math.NewInt(25)},
							{GaugeId: 7, Weight: math.NewInt(25)},
							{GaugeId: 8, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   3,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 9, Weight: math.NewInt(25)},
							{GaugeId: 10, Weight: math.NewInt(25)},
							{GaugeId: 11, Weight: math.NewInt(25)},
							{GaugeId: 12, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   4,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 13, Weight: math.NewInt(25)},
							{GaugeId: 14, Weight: math.NewInt(25)},
							{GaugeId: 15, Weight: math.NewInt(25)},
							{GaugeId: 16, Weight: math.NewInt(25)},
						},
					},
				},
			},
			expectedBlockResults: []blockResults{
				{
					height: 0,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        4,
							GaugeId:         14,
							EpochIdentifier: "day",
						},
						{
							StreamId:        0,
							GaugeId:         0,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 2, coins: nil},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 2nd stream
						{gaugeID: 5, coins: nil},
						{gaugeID: 6, coins: nil},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 14, coins: nil},
						{gaugeID: 15, coins: nil},
						{gaugeID: 16, coins: nil},
					},
				},
				{
					height: 1,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxStreamID,
							EpochIdentifier: "day",
						},
						{
							StreamId:        2,
							GaugeId:         7,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 50))},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 6, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
					},
				},
			},
		},
		{
			name:                  "Many blocks",
			maxIterationsPerBlock: 3,
			numGauges:             16,
			blocksInEpoch:         100,
			streams: []types.Stream{
				{
					Id:                   1,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(25)},
							{GaugeId: 2, Weight: math.NewInt(25)},
							{GaugeId: 3, Weight: math.NewInt(25)},
							{GaugeId: 4, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   2,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 5, Weight: math.NewInt(25)},
							{GaugeId: 6, Weight: math.NewInt(25)},
							{GaugeId: 7, Weight: math.NewInt(25)},
							{GaugeId: 8, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   3,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 9, Weight: math.NewInt(25)},
							{GaugeId: 10, Weight: math.NewInt(25)},
							{GaugeId: 11, Weight: math.NewInt(25)},
							{GaugeId: 12, Weight: math.NewInt(25)},
						},
					},
				},
				{
					Id:                   4,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("udym", 100)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 13, Weight: math.NewInt(25)},
							{GaugeId: 14, Weight: math.NewInt(25)},
							{GaugeId: 15, Weight: math.NewInt(25)},
							{GaugeId: 16, Weight: math.NewInt(25)},
						},
					},
				},
			},
			expectedBlockResults: []blockResults{
				{
					height: 0,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        1,
							GaugeId:         4,
							EpochIdentifier: "day",
						},
						{
							StreamId:        0,
							GaugeId:         0,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 75))},
						{streamID: 2, coins: nil},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: nil},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: nil},
						// 2nd stream
						{gaugeID: 5, coins: nil},
						{gaugeID: 6, coins: nil},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: nil},
						{gaugeID: 14, coins: nil},
						{gaugeID: 15, coins: nil},
						{gaugeID: 16, coins: nil},
					},
				},
				{
					height: 1,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        4,
							GaugeId:         15,
							EpochIdentifier: "day",
						},
						{
							StreamId:        0,
							GaugeId:         0,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 2, coins: nil},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 50))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 2nd stream
						{gaugeID: 5, coins: nil},
						{gaugeID: 6, coins: nil},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 15, coins: nil},
						{gaugeID: 16, coins: nil},
					},
				},
				{
					height: 2,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxGaugeID,
							EpochIdentifier: "day",
						},
						{
							StreamId:        2,
							GaugeId:         6,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 6, coins: nil},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
					},
				},
				{
					height: 3,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxGaugeID,
							EpochIdentifier: "day",
						},
						{
							StreamId:        3,
							GaugeId:         9,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 6, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 7, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 8, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
					},
				},
				{
					height: 3,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxGaugeID,
							EpochIdentifier: "day",
						},
						{
							StreamId:        3,
							GaugeId:         12,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 75))},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 6, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 7, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 8, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 3rd stream
						{gaugeID: 9, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 10, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 11, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
					},
				},
				{
					height: 4,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxGaugeID,
							EpochIdentifier: "day",
						},
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxStreamID,
							EpochIdentifier: "hour",
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 100))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 6, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 7, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 8, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 3rd stream
						{gaugeID: 9, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 10, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 11, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 12, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("udym", 25))},
					},
				},
			},
		},
	}

	// Run tests
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			s.CreateGaugesUntil(tc.numGauges)

			s.Require().LessOrEqual(len(tc.expectedBlockResults), tc.blocksInEpoch)

			// Update module params
			params := s.App.StreamerKeeper.GetParams(s.Ctx)
			params.MaxIterationsPerBlock = tc.maxIterationsPerBlock
			s.App.StreamerKeeper.SetParams(s.Ctx, params)

			for _, stream := range tc.streams {
				s.CreateStream(stream.DistributeTo.Records, stream.Coins, time.Now().Add(-time.Minute), stream.DistrEpochIdentifier, 1)
			}

			// Start epochs
			err := s.App.StreamerKeeper.BeforeEpochStart(s.Ctx, "day")
			s.Require().NoError(err)
			err = s.App.StreamerKeeper.BeforeEpochStart(s.Ctx, "hour")
			s.Require().NoError(err)

			for i := range tc.blocksInEpoch {
				err = s.App.StreamerKeeper.EndBlock(s.Ctx)
				s.Require().NoError(err)

				// Check expected rewards against actual rewards received
				gauges := s.App.IncentivesKeeper.GetGauges(s.Ctx)
				actualGauges := make(gaugeCoinsSlice, 0, len(gauges))
				for _, gauge := range gauges {
					actualGauges = append(actualGauges, gaugeCoins{gaugeID: gauge.Id, coins: gauge.Coins})
				}

				// Check block results
				idx := i
				if idx >= len(tc.expectedBlockResults) {
					idx = len(tc.expectedBlockResults) - 1
				}
				expected := tc.expectedBlockResults[idx]

				// Verify epoch pointers are valid
				pointers, err := s.App.StreamerKeeper.GetAllEpochPointers(s.Ctx)
				s.Require().NoError(err)
				// Equality here is important!
				s.Require().Equal(expected.epochPointers, pointers)

				// Verify gauges are rewarded. Equality here is important!
				s.Require().Equal(expected.gauges, actualGauges, "block height: %d\nexpect: %s\nactual: %s", i, expected, actualGauges)

				// Verify streams are valid
				active := s.App.StreamerKeeper.GetActiveStreams(s.Ctx)
				actualActive := make(distributedCoinsSlice, 0, len(gauges))
				for _, a := range active {
					actualActive = append(actualActive, distributedCoins{streamID: a.Id, coins: a.DistributedCoins})
				}
				// Equality here is important!
				s.Require().Equal(expected.distributedCoins, actualActive)
			}
		})
	}
}

type gaugeCoins struct {
	gaugeID uint64
	coins   sdk.Coins
}

func (g gaugeCoins) String() string {
	return fmt.Sprintf("gaugeID: %d, coins: %s", g.gaugeID, g.coins)
}

type gaugeCoinsSlice []gaugeCoins

func (s gaugeCoinsSlice) String() string {
	var result string
	result += "["
	for i, v := range s {
		result += v.String()
		if i < len(s)-1 {
			result += ", "
		}
	}
	result += "]"
	return result
}

type distributedCoins struct {
	streamID uint64
	coins    sdk.Coins
}

func (d distributedCoins) String() string {
	return fmt.Sprintf("streamID: %d, coins: %s", d.streamID, d.coins)
}

type distributedCoinsSlice []distributedCoins

func (s distributedCoinsSlice) String() string {
	var result string
	result += "["
	for i, v := range s {
		result += v.String()
		if i < len(s)-1 {
			result += ", "
		}
	}
	result += "]"
	return result
}

type blockResults struct {
	height           uint64
	epochPointers    []types.EpochPointer
	distributedCoins distributedCoinsSlice
	gauges           gaugeCoinsSlice
}

func (b blockResults) String() string {
	return fmt.Sprintf("height: %d, epochPointer: %v, distributedCoins: %s, gauges: %s", b.height, b.epochPointers, b.distributedCoins, b.gauges)
}

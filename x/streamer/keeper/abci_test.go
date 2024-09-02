package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (s *KeeperTestSuite) TestProcessEpochPointer() {
	addrs := apptesting.CreateRandomAccounts(2)
	tests := []struct {
		name                  string
		maxIterationsPerBlock uint64
		numGauges             int
		blocksToProcess       int
		initialLockups        []lockup
		streams               []types.Stream
		expectedBlockResults  []blockResults
	}{
		{
			// In this test, the number of gauges is less than the number of iterations. We simulate the
			// execution of the first block of the epoch:
			// 1. There are 4 streams, and each streams holds 200 stake
			// 2. Each stream has 4 gauges with 25% weight each => the number of gauges is 16
			// 3. We start with shorter epochs, so firstly we fill streams with the hour epoch (1 and 4)
			// 4. After, we continue with the longer streams (2 and 3)
			// 5. There are 9 iterations limit per block, so we fill first 9 gauges:
			//	* 4 gauges from stream 1
			//	* 4 gauges from stream 4
			//	* 1 gauge  from stream 2
			// 6. Each gauge gets 25% of the stream => 50 stake => 50 * 9 = 450 stake is totally distributed
			// 7. Initially, we have 2 lockup owners with 100 stake locked each, so each of them gets 50% of rewards
			//	of the stake denom => every owner will get 225 stake
			name:                  "1 block in the epoch",
			maxIterationsPerBlock: 9,
			numGauges:             16,
			blocksToProcess:       1,
			initialLockups: []lockup{ // every lockup owner receives 50% of the gauge rewards
				{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
				{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
			},
			streams: []types.Stream{
				{
					Id:                   1,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(50)},
							{GaugeId: 2, Weight: math.NewInt(50)},
							{GaugeId: 3, Weight: math.NewInt(50)},
							{GaugeId: 4, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   2,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 5, Weight: math.NewInt(50)},
							{GaugeId: 6, Weight: math.NewInt(50)},
							{GaugeId: 7, Weight: math.NewInt(50)},
							{GaugeId: 8, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   3,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 9, Weight: math.NewInt(50)},
							{GaugeId: 10, Weight: math.NewInt(50)},
							{GaugeId: 11, Weight: math.NewInt(50)},
							{GaugeId: 12, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   4,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 13, Weight: math.NewInt(50)},
							{GaugeId: 14, Weight: math.NewInt(50)},
							{GaugeId: 15, Weight: math.NewInt(50)},
							{GaugeId: 16, Weight: math.NewInt(50)},
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
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        2,
							GaugeId:         6,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used
						{
							StreamId:        types.MinStreamID,
							GaugeId:         types.MinStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 6, coins: nil},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 225))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 225))},
					},
				},
			},
		},
		{
			name:                  "Several blocks in the epoch",
			maxIterationsPerBlock: 5,
			numGauges:             16,
			blocksToProcess:       2,
			initialLockups: []lockup{ // every lockup owner receives 50% of the gauge rewards
				{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
				{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
			},
			streams: []types.Stream{
				{
					Id:                   1,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(50)},
							{GaugeId: 2, Weight: math.NewInt(50)},
							{GaugeId: 3, Weight: math.NewInt(50)},
							{GaugeId: 4, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   2,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 5, Weight: math.NewInt(50)},
							{GaugeId: 6, Weight: math.NewInt(50)},
							{GaugeId: 7, Weight: math.NewInt(50)},
							{GaugeId: 8, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   3,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 9, Weight: math.NewInt(50)},
							{GaugeId: 10, Weight: math.NewInt(50)},
							{GaugeId: 11, Weight: math.NewInt(50)},
							{GaugeId: 12, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   4,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 13, Weight: math.NewInt(50)},
							{GaugeId: 14, Weight: math.NewInt(50)},
							{GaugeId: 15, Weight: math.NewInt(50)},
							{GaugeId: 16, Weight: math.NewInt(50)},
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
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        0,
							GaugeId:         0,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used
						{
							StreamId:        types.MinStreamID,
							GaugeId:         types.MinStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 2, coins: nil},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
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
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 14, coins: nil},
						{gaugeID: 15, coins: nil},
						{gaugeID: 16, coins: nil},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 125))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 125))},
					},
				},
				{
					height: 1,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxStreamID,
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        2,
							GaugeId:         7,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used
						{
							StreamId:        types.MinStreamID,
							GaugeId:         types.MinStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 6, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 250))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 250))},
					},
				},
			},
		},
		{
			name:                  "Send all reward in one single block",
			maxIterationsPerBlock: 5,
			numGauges:             4,
			blocksToProcess:       5,
			initialLockups: []lockup{ // every lockup owner receives 50% of the gauge rewards
				{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
				{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
			},
			streams: []types.Stream{
				{
					Id:                   1,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 2)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(1)},
						},
					},
				},
				{
					Id:                   2,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 2)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 2, Weight: math.NewInt(1)},
						},
					},
				},
				{
					Id:                   3,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 2)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 3, Weight: math.NewInt(1)},
						},
					},
				},
				{
					Id:                   4,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 2)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 4, Weight: math.NewInt(1)},
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
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxStreamID,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used, however it points on the last gauge
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 2))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 2))},
						{streamID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 2))},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 2))},
					},
					gauges: []gaugeCoins{
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 2))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 2))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 2))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 2))},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 4))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 4))},
					},
				},
			},
		},
		{
			name:                  "Many blocks",
			maxIterationsPerBlock: 3,
			numGauges:             16,
			blocksToProcess:       200,
			initialLockups: []lockup{ // every lockup owner receives 50% of the gauge rewards
				{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
				{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
			},
			streams: []types.Stream{
				{
					Id:                   1,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 1, Weight: math.NewInt(50)},
							{GaugeId: 2, Weight: math.NewInt(50)},
							{GaugeId: 3, Weight: math.NewInt(50)},
							{GaugeId: 4, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   2,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 5, Weight: math.NewInt(50)},
							{GaugeId: 6, Weight: math.NewInt(50)},
							{GaugeId: 7, Weight: math.NewInt(50)},
							{GaugeId: 8, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   3,
					DistrEpochIdentifier: "day",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 9, Weight: math.NewInt(50)},
							{GaugeId: 10, Weight: math.NewInt(50)},
							{GaugeId: 11, Weight: math.NewInt(50)},
							{GaugeId: 12, Weight: math.NewInt(50)},
						},
					},
				},
				{
					Id:                   4,
					DistrEpochIdentifier: "hour",
					Coins:                sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
					DistributeTo: &types.DistrInfo{
						Records: []types.DistrRecord{
							{GaugeId: 13, Weight: math.NewInt(50)},
							{GaugeId: 14, Weight: math.NewInt(50)},
							{GaugeId: 15, Weight: math.NewInt(50)},
							{GaugeId: 16, Weight: math.NewInt(50)},
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
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        0,
							GaugeId:         0,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used
						{
							StreamId:        types.MinStreamID,
							GaugeId:         types.MinStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 150))},
						{streamID: 2, coins: nil},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: nil},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
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
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 75))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 75))},
					},
				},
				{
					height: 1,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        4,
							GaugeId:         15,
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        0,
							GaugeId:         0,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used
						{
							StreamId:        types.MinStreamID,
							GaugeId:         types.MinStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 2, coins: nil},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
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
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 15, coins: nil},
						{gaugeID: 16, coins: nil},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 150))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 150))},
					},
				},
				{
					height: 2,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxGaugeID,
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        2,
							GaugeId:         6,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used
						{
							StreamId:        types.MinStreamID,
							GaugeId:         types.MinStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 6, coins: nil},
						{gaugeID: 7, coins: nil},
						{gaugeID: 8, coins: nil},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 225))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 225))},
					},
				},
				{
					height: 3,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxGaugeID,
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        3,
							GaugeId:         9,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used
						{
							StreamId:        types.MinStreamID,
							GaugeId:         types.MinStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 3, coins: nil},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 6, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 7, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 8, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 3rd stream
						{gaugeID: 9, coins: nil},
						{gaugeID: 10, coins: nil},
						{gaugeID: 11, coins: nil},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 300))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 300))},
					},
				},
				{
					height: 3,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxGaugeID,
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        3,
							GaugeId:         12,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used
						{
							StreamId:        types.MinStreamID,
							GaugeId:         types.MinStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 150))},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 6, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 7, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 8, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 3rd stream
						{gaugeID: 9, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 10, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 11, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 12, coins: nil},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 375))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 375))},
					},
				},
				{
					height: 4,
					epochPointers: []types.EpochPointer{
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxGaugeID,
							EpochIdentifier: "hour",
							EpochDuration:   time.Hour,
						},
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxStreamID,
							EpochIdentifier: "day",
							EpochDuration:   24 * time.Hour,
						},
						// week epoch pointer is not used, however it points on the last gauge
						{
							StreamId:        types.MaxStreamID,
							GaugeId:         types.MaxStreamID,
							EpochIdentifier: "week",
							EpochDuration:   7 * 24 * time.Hour,
						},
					},
					distributedCoins: []distributedCoins{
						{streamID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
						{streamID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
					},
					gauges: []gaugeCoins{
						// 1st stream
						{gaugeID: 1, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 2, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 3, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 4, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 2nd stream
						{gaugeID: 5, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 6, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 7, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 8, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 3rd stream
						{gaugeID: 9, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 10, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 11, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 12, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						// 4th stream
						{gaugeID: 13, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 14, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 15, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
						{gaugeID: 16, coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 50))},
					},
					lockups: []lockup{
						{owner: addrs[0], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 400))},
						{owner: addrs[1], balance: sdk.NewCoins(sdk.NewInt64Coin("stake", 400))},
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

			for _, lock := range tc.initialLockups {
				s.LockTokens(lock.owner, lock.balance)
			}

			s.Require().LessOrEqual(len(tc.expectedBlockResults), tc.blocksToProcess)

			// Update module params
			params := s.App.StreamerKeeper.GetParams(s.Ctx)
			params.MaxIterationsPerBlock = tc.maxIterationsPerBlock
			s.App.StreamerKeeper.SetParams(s.Ctx, params)

			for _, stream := range tc.streams {
				s.CreateStream(stream.DistributeTo.Records, stream.Coins, time.Now().Add(-time.Minute), stream.DistrEpochIdentifier, 1)
			}

			// Start epochs
			err := s.App.StreamerKeeper.BeforeEpochStart(s.Ctx, "hour")
			s.Require().NoError(err)
			err = s.App.StreamerKeeper.BeforeEpochStart(s.Ctx, "day")
			s.Require().NoError(err)

			for i := range tc.blocksToProcess {
				err = s.App.StreamerKeeper.EndBlock(s.Ctx)
				s.Require().NoError(err)

				// Check expected rewards against actual rewards received
				gauges := s.App.IncentivesKeeper.GetGauges(s.Ctx)
				actualGauges := make(gaugeCoinsSlice, 0, len(gauges))
				for _, gauge := range gauges {
					actualGauges = append(actualGauges, gaugeCoins{gaugeID: gauge.Id, coins: gauge.DistributedCoins})
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
				// Equality here is important! Pointers must be filled from shorter to longer.
				types.SortEpochPointers(pointers)
				s.Require().Equal(expected.epochPointers, pointers)

				// Verify gauges are rewarded. Equality here is important!
				s.Require().Equal(expected.gauges, actualGauges, "block height: %d\nexpect: %s\nactual: %s", i, expected.gauges, actualGauges)

				// Verify lockup owner are rewarded
				for _, lock := range expected.lockups {
					actualBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, lock.owner)
					s.Require().Equal(lock.balance, actualBalance)
				}

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

type lockup struct {
	owner   sdk.AccAddress
	balance sdk.Coins
}

type blockResults struct {
	height           uint64
	epochPointers    []types.EpochPointer
	distributedCoins distributedCoinsSlice
	gauges           gaugeCoinsSlice
	lockups          []lockup
}

func (b blockResults) String() string {
	return fmt.Sprintf("height: %d, epochPointer: %v, distributedCoins: %s, gauges: %s", b.height, b.epochPointers, b.distributedCoins, b.gauges)
}

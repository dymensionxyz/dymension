package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

func (suite *KeeperTestSuite) TestDistribute() {
	tests := []struct {
		name    string
		streams []struct {
			coins       sdk.Coins
			numOfEpochs uint64
			distrInfo   *types.DistrInfo
		}
	}{
		{
			name: "single stream single coin",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				distrInfo   *types.DistrInfo
			}{{sdk.Coins{sdk.NewInt64Coin("stake", 100)}, 30, defaultDistrInfo}},
		},
		{
			name: "single stream multiple coins",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				distrInfo   *types.DistrInfo
			}{{sdk.Coins{sdk.NewInt64Coin("stake", 100), sdk.NewInt64Coin("udym", 300)}, 30, defaultDistrInfo}},
		},
		{
			name: "multiple streams multiple coins multiple epochs",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				distrInfo   *types.DistrInfo
			}{{sdk.Coins{sdk.NewInt64Coin("stake", 100), sdk.NewInt64Coin("udym", 300)}, 30, defaultDistrInfo},
				{sdk.Coins{sdk.NewInt64Coin("stake", 1000)}, 365, defaultDistrInfo},
				{sdk.Coins{sdk.NewInt64Coin("udym", 1000)}, 730, defaultDistrInfo},
			},
		},
	}
	for _, tc := range tests {
		suite.SetupTest()
		// setup streams and defined in the above tests, then distribute to them

		err := suite.CreateGauge()
		suite.Require().NoError(err)
		err = suite.CreateGauge()
		suite.Require().NoError(err)

		var streams []types.Stream
		var gaugesExpectedRewards = make(map[uint64]sdk.Coins)
		for _, stream := range tc.streams {
			// create a stream
			_, newstream := suite.CreateStream(stream.distrInfo, stream.coins, time.Now(), "day", stream.numOfEpochs)
			streams = append(streams, *newstream)

			// calculate expected rewards
			for _, coin := range stream.coins {
				epochAmt := coin.Amount.Quo(sdk.NewInt(int64(stream.numOfEpochs)))
				if !epochAmt.IsPositive() {
					continue
				}
				for _, record := range stream.distrInfo.Records {
					expectedAmtFromStream := epochAmt.Mul(record.Weight).Quo(stream.distrInfo.TotalWeight)
					expectedCoins := sdk.Coin{Denom: coin.Denom, Amount: expectedAmtFromStream}
					gaugesExpectedRewards[record.GaugeId] = gaugesExpectedRewards[record.GaugeId].Add(expectedCoins)
				}
			}
		}

		_, err = suite.App.StreamerKeeper.Distribute(suite.Ctx, streams)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		gauges := suite.App.IncentivesKeeper.GetGauges(suite.Ctx)
		suite.Require().Equal(len(gaugesExpectedRewards), len(gauges))
		for _, gauge := range gauges {
			suite.Require().Equal(gaugesExpectedRewards[gauge.Id], gauge.Coins)
		}
	}
}

// TestGetModuleToDistributeCoins tests the sum of coins yet to be distributed for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	var err error
	suite.SetupTest()
	err = suite.CreateGauge()
	suite.Require().NoError(err)
	err = suite.CreateGauge()
	suite.Require().NoError(err)

	// check that the sum of coins yet to be distributed is nil
	coins := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup a stream
	streamCoins := sdk.Coins{sdk.NewInt64Coin("stake", 600000)}
	_, _ = suite.CreateDefaultStream(streamCoins)

	// check that the sum of coins yet to be distributed is equal to the newly created streamCoins
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins)

	// create a new stream
	// check that the sum of coins yet to be distributed is equal to the stream1 and stream2 coins combined

	streamCoins2 := sdk.Coins{sdk.NewInt64Coin("udym", 300000)}
	_, _ = suite.CreateDefaultStream(streamCoins2)

	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins.Add(streamCoins2...))

	// move all created streams from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())
	streams := suite.App.StreamerKeeper.GetStreams(suite.Ctx)
	for _, stream := range streams {
		err := suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, stream)
		suite.Require().NoError(err)
	}

	// distribute coins to stakers
	distrCoins, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, streams)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 20000), sdk.NewInt64Coin("udym", 10000)}, distrCoins)

	// check stream changes after distribution
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins.Add(streamCoins2...).Sub(distrCoins...))
}

package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

var (
	destAddr1 = sdk.AccAddress([]byte("addr1---------------"))
	destAddr2 = sdk.AccAddress([]byte("addr2---------------"))
)

//TODO: destination is account/module

func (suite *KeeperTestSuite) TestDistribute() {
	tests := []struct {
		name    string
		streams []struct {
			coins       sdk.Coins
			numOfEpochs uint64
			destAddr    sdk.AccAddress
		}
	}{
		{
			name: "single stream single coin",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				destAddr    sdk.AccAddress
			}{{sdk.Coins{sdk.NewInt64Coin("stake", 100)}, 30, destAddr1}},
		},
		{
			name: "single stream multiple coins",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				destAddr    sdk.AccAddress
			}{{sdk.Coins{sdk.NewInt64Coin("stake", 100), sdk.NewInt64Coin("udym", 300)}, 30, destAddr1}},
		},
		{
			name: "multiple streams multiple coins multiple epochs",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				destAddr    sdk.AccAddress
			}{{sdk.Coins{sdk.NewInt64Coin("stake", 100), sdk.NewInt64Coin("udym", 300)}, 30, destAddr1},
				{sdk.Coins{sdk.NewInt64Coin("stake", 1000)}, 365, destAddr1},
				{sdk.Coins{sdk.NewInt64Coin("udym", 1000)}, 730, destAddr2},
			},
		},
	}
	for _, tc := range tests {
		suite.SetupTest()
		// setup streams and defined in the above tests, then distribute to them

		var streams []types.Stream
		var destAddrsExpectedRewards = make(map[string]sdk.Coins)
		for _, stream := range tc.streams {
			// create a stream
			_, newstream := suite.CreateStream(stream.destAddr, stream.coins, time.Time{}, "day", stream.numOfEpochs)
			streams = append(streams, *newstream)

			// calculate expected rewards
			expectedCoinsFromSream := destAddrsExpectedRewards[stream.destAddr.String()]
			for _, coin := range stream.coins {
				epochAmt := coin.Amount.Quo(sdk.NewInt(int64(stream.numOfEpochs)))
				if epochAmt.IsPositive() {
					newlyDistributedCoin := sdk.Coin{Denom: coin.Denom, Amount: epochAmt}
					expectedCoinsFromSream = expectedCoinsFromSream.Add(newlyDistributedCoin)
				}
			}
			destAddrsExpectedRewards[stream.destAddr.String()] = expectedCoinsFromSream
		}

		_, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, streams)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		for addr, expecetedBalance := range destAddrsExpectedRewards {
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sdk.MustAccAddressFromBech32(addr))
			suite.Require().Equal(expecetedBalance.String(), bal.String(), "test %v, dest %s", tc.name, addr)
		}
	}
}

// TestGetModuleToDistributeCoins tests the sum of coins yet to be distributed for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	suite.SetupTest()

	// check that the sum of coins yet to be distributed is nil
	coins := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup a non perpetual lock and stream
	streamCoins := sdk.Coins{sdk.NewInt64Coin("stake", 100)}
	_, _ = suite.CreateDefaultStream(streamCoins)

	// check that the sum of coins yet to be distributed is equal to the newly created streamCoins
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins)

	// create a new stream
	// check that the sum of coins yet to be distributed is equal to the stream1 and stream2 coins combined

	streamCoins2 := sdk.Coins{sdk.NewInt64Coin("udym", 300)}
	suite.FundModuleAcc(types.ModuleName, streamCoins2)
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
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check stream changes after distribution
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins.Add(streamCoins2...).Sub(distrCoins...))
}

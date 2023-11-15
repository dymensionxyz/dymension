package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

var (
	defaultExpectedStream = types.Stream{
		NumEpochsPaidOver: uint64(30),
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         time.Time{},
	}
)

func (suite *KeeperTestSuite) TestHookOperation() {
	suite.SetupTest()

	// initial module streams check
	streams := suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 0)

	// setup streams

	//daily stream, 30 epochs
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 30000)}
	_, _ = suite.CreateDefaultStream(coins)

	//daily stream, 2 epochs
	coins2 := sdk.Coins{sdk.NewInt64Coin("stake", 2000)}
	_, _ = suite.CreateStream(defaultDistrInfo, coins2, time.Now(), "day", 2)

	//weekly stream
	coins3 := sdk.Coins{sdk.NewInt64Coin("stake", 5000)}
	_, _ = suite.CreateStream(defaultDistrInfo, coins3, time.Now(), "week", 5)

	//future stream - non-active
	_, _ = suite.CreateStream(defaultDistrInfo, coins3, time.Now().Add(10*time.Minute), "day", 5)

	// check streams
	streams = suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 4)

	// check upcoming streams
	streams = suite.App.StreamerKeeper.GetUpcomingStreams(suite.Ctx)
	suite.Require().Len(streams, 4)

	// check active streams
	streams = suite.App.StreamerKeeper.GetActiveStreams(suite.Ctx)
	suite.Require().Len(streams, 0)

	/* ----------- call the epoch hook with month (no stream related) ----------- */
	ctx := suite.Ctx.WithBlockTime(time.Now())
	err := suite.App.StreamerKeeper.Hooks().AfterEpochEnd(ctx, "month", 0)
	suite.Require().NoError(err)

	// check active streams - all 3 but the future are active
	streams = suite.App.StreamerKeeper.GetActiveStreams(suite.Ctx)
	suite.Require().Len(streams, 3)

	/* --------- call the epoch hook with day (2 active and one future) --------- */
	err = suite.App.StreamerKeeper.Hooks().AfterEpochEnd(ctx, "day", 0)
	suite.Require().NoError(err)

	// check active streams
	streams = suite.App.StreamerKeeper.GetActiveStreams(suite.Ctx)
	suite.Require().Len(streams, 3)

	// check upcoming streams
	streams = suite.App.StreamerKeeper.GetUpcomingStreams(suite.Ctx)
	suite.Require().Len(streams, 1)

	// check distribution
	balance := suite.App.BankKeeper.GetBalance(ctx, defaultDestAddr, "stake")
	suite.Require().Equal(sdk.NewInt64Coin("stake", 2000), balance)

	/* ------------------------- call weekly epoch hook ------------------------- */
	err = suite.App.StreamerKeeper.Hooks().AfterEpochEnd(ctx, "week", 0)
	suite.Require().NoError(err)

	// check active streams
	streams = suite.App.StreamerKeeper.GetActiveStreams(suite.Ctx)
	suite.Require().Len(streams, 3)

	// check upcoming streams
	streams = suite.App.StreamerKeeper.GetUpcomingStreams(suite.Ctx)
	suite.Require().Len(streams, 1)

	// check distribution
	balance = suite.App.BankKeeper.GetBalance(ctx, defaultDestAddr, "stake")
	suite.Require().Equal(sdk.NewInt64Coin("stake", 3000), balance)

	/* ------- call daily epoch hook again, check both stream distirubute ------- */
	err = suite.App.StreamerKeeper.Hooks().AfterEpochEnd(ctx, "day", 0)
	suite.Require().NoError(err)

	// check distribution
	balance = suite.App.BankKeeper.GetBalance(ctx, defaultDestAddr, "stake")
	suite.Require().Equal(sdk.NewInt64Coin("stake", 5000), balance)

	/* ------- call daily epoch hook again, check both stream distirubute ------- */
	err = suite.App.StreamerKeeper.Hooks().AfterEpochEnd(ctx, "day", 0)
	suite.Require().NoError(err)

	//check finisihed stream
	streams = suite.App.StreamerKeeper.GetFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 1)

	// check distribution
	balance = suite.App.BankKeeper.GetBalance(ctx, defaultDestAddr, "stake")
	suite.Require().Equal(sdk.NewInt64Coin("stake", 6000), balance)
}

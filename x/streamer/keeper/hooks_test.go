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
	// test for module get streams
	suite.SetupTest()

	// initial module streams check
	streams := suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 0)

	// setup stream
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 12)}
	streamID, _ := suite.CreateDefaultStream(coins)
	var expectedNumEpochsPaidOver int = 30

	//FIXME: another stream with 2 epochs
	//FIXME: stream with differnt epoch identifeir

	// check streams
	streams = suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 1)
	expectedStream := defaultExpectedStream
	expectedStream.Id = streamID
	expectedStream.Coins = coins
	suite.Require().Equal(expectedStream.String(), streams[0].String())

	// check upcoming streams
	streams = suite.App.StreamerKeeper.GetUpcomingStreams(suite.Ctx)
	suite.Require().Len(streams, 1)

	// start distribution
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// check active streams
	streams = suite.App.StreamerKeeper.GetActiveStreams(suite.Ctx)
	suite.Require().Len(streams, 1)

	// check upcoming streams
	streams = suite.App.StreamerKeeper.GetUpcomingStreams(suite.Ctx)
	suite.Require().Len(streams, 0)

	// distribute coins to stakers
	distrCoins, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, []types.Stream{*stream})
	suite.Require().NoError(err)
	// We hardcoded 12 "stake" tokens when initializing stream
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(12/expectedNumEpochsPaidOver))}, distrCoins)

	//TODO: start epoch hook, check for activeness and distribution
	// check 2 stream distibuted, and one not
	//check 2nd stream finished after few eopchs
}

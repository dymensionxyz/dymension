package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v8/testing/simapp/apptesting"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"

	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// TestGRPCParams tests querying params via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCParams() {
	res, err := suite.querier.Params(suite.Ctx, &types.ParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NotNil(res.Params)

	// Check that we get the expected default params
	expectedParams := suite.App.StreamerKeeper.GetParams(suite.Ctx)
	suite.Require().Equal(expectedParams, res.Params)
}

// TestGRPCStreamByID tests querying streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCStreamByID() {
	// create a stream
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	streamID, stream := suite.CreateDefaultStream(coins)

	// ensure that querying for a stream with an ID that doesn't exist returns an error.
	res, err := suite.querier.StreamByID(suite.Ctx, &types.StreamByIDRequest{Id: 1000})
	suite.Require().Error(err)
	suite.Require().Equal(res, (*types.StreamByIDResponse)(nil))

	// check that querying a stream with an ID that exists returns the stream.
	res, err = suite.querier.StreamByID(suite.Ctx, &types.StreamByIDRequest{Id: streamID})
	suite.Require().NoError(err)
	suite.Require().NotEqual(res.Stream, nil)

	expectedStream := suite.ExpectedDefaultStream(streamID, stream.StartTime, coins)
	suite.Require().Equal(res.Stream.String(), expectedStream.String())
}

// TestGRPCStreams tests querying upcoming and active streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCStreams() {
	// ensure initially querying streams returns no streams
	res, err := suite.querier.Streams(suite.Ctx, &types.StreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a stream
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	streamID, stream := suite.CreateDefaultStream(coins)
	suite.Ctx = suite.Ctx.WithBlockTime(stream.StartTime.Add(time.Second))

	// query streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.Streams(suite.Ctx, &types.StreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := suite.ExpectedDefaultStream(streamID, stream.StartTime, coins)
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// create 10 more streams
	for i := 0; i < 10; i++ {
		// create a stream
		coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
		_, stream = suite.CreateDefaultStream(coins)
		suite.Ctx = suite.Ctx.WithBlockTime(stream.StartTime.Add(time.Second))
	}

	// check that setting page request limit to 10 will only return 10 out of the 11 streams
	filter := query.PageRequest{Limit: 10}
	res, err = suite.querier.Streams(suite.Ctx, &types.StreamsRequest{Pagination: &filter})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 10)

	filter = query.PageRequest{Limit: 13}
	res, err = suite.querier.Streams(suite.Ctx, &types.StreamsRequest{Pagination: &filter})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 11)
}

// TestGRPCActiveStreams tests querying active streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCActiveStreams() {
	// ensure initially querying active streams returns no streams
	res, err := suite.querier.ActiveStreams(suite.Ctx, &types.ActiveStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a stream and move it from upcoming to active
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
	streamID, stream := suite.CreateDefaultStream(coins)
	suite.Ctx = suite.Ctx.WithBlockTime(stream.StartTime.Add(time.Second))

	err = suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// query active streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.ActiveStreams(suite.Ctx, &types.ActiveStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)

	expectedStream := suite.ExpectedDefaultStream(streamID, stream.StartTime, coins)
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// create 20 more streams
	for i := 0; i < 20; i++ {
		coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
		_, stream := suite.CreateDefaultStream(coins)
		suite.Ctx = suite.Ctx.WithBlockTime(stream.StartTime.Add(time.Second))

		// move the first 9 streams from upcoming to active (now 10 active streams, 30 total streams)
		if i < 9 {
			err = suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
			suite.Require().NoError(err)
		}
	}

	// set page request limit to 5, expect only 5 active stream responses
	res, err = suite.querier.ActiveStreams(suite.Ctx, &types.ActiveStreamsRequest{Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 5)

	// set page request limit to 15, expect only 10 active stream responses
	res, err = suite.querier.ActiveStreams(suite.Ctx, &types.ActiveStreamsRequest{Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 10)
}

// TestGRPCUpcomingStreams tests querying upcoming streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCUpcomingStreams() {
	// ensure initially querying upcoming streams returns no streams
	res, err := suite.querier.UpcomingStreams(suite.Ctx, &types.UpcomingStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a stream
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
	streamID, stream := suite.CreateDefaultStream(coins)

	// query upcoming streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.UpcomingStreams(suite.Ctx, &types.UpcomingStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := suite.ExpectedDefaultStream(streamID, stream.StartTime, coins)

	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// setup 20 more upcoming streams
	for i := 0; i < 20; i++ {
		coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
		_, stream := suite.CreateDefaultStream(coins)
		suite.Ctx = suite.Ctx.WithBlockTime(stream.StartTime.Add(time.Second))

		// move the first 9 created streams to an active status
		// 1 + (20 -9) = 12 upcoming streams
		if i < 9 {
			err = suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
			suite.Require().NoError(err)
		}
	}

	// query upcoming streams with a page request of 5 should return 5 streams
	res, err = suite.querier.UpcomingStreams(suite.Ctx, &types.UpcomingStreamsRequest{Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 5)

	// query upcoming streams with a page request of 15 should return 12 streams
	res, err = suite.querier.UpcomingStreams(suite.Ctx, &types.UpcomingStreamsRequest{Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 12)
}

// TestGRPCToDistributeCoins tests querying coins that are going to be distributed via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCToDistributeCoins() {
	var err error

	err = suite.CreateGauge()
	suite.Require().NoError(err)
	err = suite.CreateGauge()
	suite.Require().NoError(err)

	// ensure initially querying to distribute coins returns no coins
	res, err := suite.querier.ModuleToDistributeCoins(suite.Ctx, &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// setup a non perpetual stream
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 300000)}
	streamID, stream := suite.CreateDefaultStream(coins)
	suite.Ctx = suite.Ctx.WithBlockTime(stream.StartTime.Add(time.Second))

	stream, err = suite.querier.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	suite.Require().NotNil(stream)

	// check to distribute coins after stream creation, but before stream active
	res, err = suite.querier.ModuleToDistributeCoins(suite.Ctx, &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	// check to distribute coins after stream creation
	// ensure this equals the coins within the previously created non-perpetual stream
	res, err = suite.querier.ModuleToDistributeCoins(suite.Ctx, &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	// move stream from an upcoming to an active status
	// this simulates the new epoch start
	// the stream is moved to active and its rewards are to be distributed during the epoch
	err = suite.App.StreamerKeeper.BeforeEpochStart(suite.Ctx, "day")
	suite.Require().NoError(err)

	// check to distribute coins after the epoch start
	// ensure this equals the coins within the previously created non-perpetual stream
	// the rewards are not distributed yet
	res, err = suite.querier.ModuleToDistributeCoins(suite.Ctx, &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	// trigger the epoch end. this will distribute all rewards assigned to this epoch
	distrCoins, err := suite.App.StreamerKeeper.AfterEpochEnd(suite.Ctx, "day")
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 10000)})

	// check stream changes after distribution
	// ensure the stream's filled epochs have been increased by 1
	// ensure we have distributed 4 out of the 10 stake tokens
	stream, err = suite.querier.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	suite.Require().NotNil(stream)
	suite.Require().Equal(stream.FilledEpochs, uint64(1))
	suite.Require().Equal(stream.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 10000)})

	// check that the to distribute coins is equal to the initial stream coin balance minus what has been distributed already (10-4=6)
	res, err = suite.querier.ModuleToDistributeCoins(suite.Ctx, &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins.Sub(distrCoins...))

	// trigger the next epoch start and then the next epoch end.
	// this simulates the executed epoch and consequently distributes the second round.
	err = suite.App.StreamerKeeper.BeforeEpochStart(suite.Ctx, "day")
	suite.Require().NoError(err)
	distrCoins, err = suite.App.StreamerKeeper.AfterEpochEnd(suite.Ctx, "day")
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 10000)}, distrCoins)

	// check stream changes after distribution
	// ensure the stream's filled epochs have been increased by 1
	// ensure we have distributed 4 out of the 10 stake tokens
	stream, err = suite.querier.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	suite.Require().NotNil(stream)
	suite.Require().Equal(stream.FilledEpochs, uint64(2))
	suite.Require().Equal(stream.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 20000)})

	// now that all coins have been distributed (4 in first found 6 in the second round)
	// to distribute coins should be null
	res, err = suite.querier.ModuleToDistributeCoins(suite.Ctx, &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{sdk.NewInt64Coin("stake", 280000)})
}

func (suite *KeeperTestSuite) TestPumpPressure() {
	rollapp1ID := suite.CreateDefaultRollapp()
	_ = suite.CreateDefaultRollapp()

	userAddr := apptesting.CreateRandomAccounts(1)[0]

	vote := sponsorshiptypes.MsgVote{
		Voter: userAddr.String(),
		Weights: []sponsorshiptypes.GaugeWeight{
			{GaugeId: 1, Weight: common.DYM.MulRaw(60)},
			{GaugeId: 2, Weight: common.DYM.MulRaw(40)},
		},
	}
	suite.CreateValVote(vote, common.DYM.MulRaw(100))

	coins := sdk.NewCoins(common.DymUint64(100))
	suite.CreatePumpStream(coins, time.Now(), "day", 30, &types.MsgCreateStream_PumpParams{})

	suite.Run("PumpPressure returns all rollapp pressures", func() {
		res, err := suite.querier.PumpPressure(suite.Ctx, &types.PumpPressureRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Len(res.Pressure, 2)
	})

	suite.Run("PumpPressureByRollapp returns specific rollapp pressure", func() {
		res, err := suite.querier.PumpPressureByRollapp(suite.Ctx, &types.PumpPressureByRollappRequest{
			RollappId: rollapp1ID,
		})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Equal(res.Pressure.RollappId, rollapp1ID)
		suite.Require().True(res.Pressure.Pressure.GT(math.ZeroInt()))
	})

	suite.Run("PumpPressureByUser returns user pressure for all rollapps", func() {
		res, err := suite.querier.PumpPressureByUser(suite.Ctx, &types.PumpPressureByUserRequest{
			Address: userAddr.String(),
		})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Len(res.Pressure, 2)
	})

	suite.Run("PumpPressureByUserByRollapp returns user pressure for specific rollapp", func() {
		res, err := suite.querier.PumpPressureByUserByRollapp(suite.Ctx, &types.PumpPressureByUserByRollappRequest{
			Address:   userAddr.String(),
			RollappId: rollapp1ID,
		})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Equal(res.Pressure.RollappId, rollapp1ID)
		suite.Require().True(res.Pressure.Pressure.GT(math.ZeroInt()))
	})
}

package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	query "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/dymensionxyz/dymension/x/streamer/types"
)

var _ = suite.TestingSuite(nil)

// TestGRPCStreamByID tests querying streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCStreamByID() {
	suite.SetupTest()

	// create a stream
	streamID, _, coins, startTime := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// ensure that querying for a stream with an ID that doesn't exist returns an error.
	res, err := suite.querier.StreamByID(sdk.WrapSDKContext(suite.Ctx), &types.StreamByIDRequest{Id: 1000})
	suite.Require().Error(err)
	suite.Require().Equal(res, (*types.StreamByIDResponse)(nil))

	// check that querying a stream with an ID that exists returns the stream.
	res, err = suite.querier.StreamByID(sdk.WrapSDKContext(suite.Ctx), &types.StreamByIDRequest{Id: streamID})
	suite.Require().NoError(err)
	suite.Require().NotEqual(res.Stream, nil)

	//FIXME: expected should be given from the SetupNewStream function
	expectedStream := types.Stream{
		Id:                streamID,
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Stream.String(), expectedStream.String())
}

// TestGRPCStreams tests querying upcoming and active streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCStreams() {
	suite.SetupTest()

	// ensure initially querying streams returns no streams
	res, err := suite.querier.Streams(sdk.WrapSDKContext(suite.Ctx), &types.StreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a stream
	streamID, _, coins, startTime := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// query streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.Streams(sdk.WrapSDKContext(suite.Ctx), &types.StreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := types.Stream{
		Id:                streamID,
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// create 10 more streams
	for i := 0; i < 10; i++ {
		suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))
	}

	// check that setting page request limit to 10 will only return 10 out of the 11 streams
	filter := query.PageRequest{Limit: 10}
	res, err = suite.querier.Streams(sdk.WrapSDKContext(suite.Ctx), &types.StreamsRequest{Pagination: &filter})
	suite.Require().Len(res.Data, 10)
}

// TestGRPCActiveStreams tests querying active streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCActiveStreams() {
	suite.SetupTest()

	// ensure initially querying active streams returns no streams
	res, err := suite.querier.ActiveStreams(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a stream and move it from upcoming to active
	_, stream, coins, startTime := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// query active streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.ActiveStreams(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := types.Stream{
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// create 20 more streams
	for i := 0; i < 20; i++ {
		_, stream, _, _ := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 9 streams from upcoming to active (now 10 active streams, 30 total streams)
		if i < 9 {
			suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
		}
	}

	// set page request limit to 5, expect only 5 active stream responses
	res, err = suite.querier.ActiveStreams(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsRequest{Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.Data, 5)

	// set page request limit to 15, expect only 10 active stream responses
	res, err = suite.querier.ActiveStreams(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsRequest{Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().Len(res.Data, 10)
}

// TestGRPCActiveStreamsPerDenom tests querying active streams by denom via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCActiveStreamsPerDenom() {
	suite.SetupTest()

	// ensure initially querying streams by denom returns no streams
	res, err := suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a stream
	streamID, stream, coins, startTime := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)

	// query streams by denom again, but this time expect the stream created earlier in the response
	res, err = suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{Denom: "lptoken", Pagination: nil})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := types.Stream{
		Id:                streamID,
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// setup 20 more streams with the pool denom
	for i := 0; i < 20; i++ {
		_, stream, _, _ := suite.SetupNewStreamWithDenom(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)}, "pool")
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 10 of 20 streams to an active status
		if i < 10 {
			suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
		}
	}

	// query active streams by lptoken denom with a page request of 5 should only return one stream
	res, err = suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{Denom: "lptoken", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.Data, 1)

	// query active streams by pool denom with a page request of 5 should return 5 streams
	res, err = suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.Data, 5)

	// query active streams by pool denom with a page request of 15 should return 10 streams
	res, err = suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().Len(res.Data, 10)
}

// TestGRPCUpcomingStreams tests querying upcoming streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCUpcomingStreams() {
	suite.SetupTest()

	// ensure initially querying upcoming streams returns no streams
	res, err := suite.querier.UpcomingStreams(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a stream
	streamID, _, coins, startTime := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// query upcoming streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.UpcomingStreams(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := types.Stream{
		Id:                streamID,
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// setup 20 more upcoming streams
	for i := 0; i < 20; i++ {
		_, stream, _, _ := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 9 created streams to an active status
		// 1 + (20 -9) = 12 upcoming streams
		if i < 9 {
			suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
		}
	}

	// query upcoming streams with a page request of 5 should return 5 streams
	res, err = suite.querier.UpcomingStreams(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsRequest{Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.Data, 5)

	// query upcoming streams with a page request of 15 should return 12 streams
	res, err = suite.querier.UpcomingStreams(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsRequest{Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().Len(res.Data, 12)
}

// TestGRPCUpcomingStreamsPerDenom tests querying upcoming streams by denom via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCUpcomingStreamsPerDenom() {
	suite.SetupTest()

	// ensure initially querying upcoming streams by denom returns no streams
	upcomingStreamRequest := types.UpcomingStreamsPerDenomRequest{Denom: "lptoken", Pagination: nil}
	res, err := suite.querier.UpcomingStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &upcomingStreamRequest)
	suite.Require().NoError(err)
	suite.Require().Len(res.UpcomingStreams, 0)

	// create a stream, and check upcoming stream is working
	streamID, stream, coins, startTime := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// query upcoming streams by denom again, but this time expect the stream created earlier in the response
	res, err = suite.querier.UpcomingStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &upcomingStreamRequest)
	suite.Require().NoError(err)
	suite.Require().Len(res.UpcomingStreams, 1)
	expectedStream := types.Stream{
		Id:                streamID,
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.UpcomingStreams[0].String(), expectedStream.String())

	// move stream from upcoming to active
	// ensure the query no longer returns a response
	suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	res, err = suite.querier.UpcomingStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &upcomingStreamRequest)
	suite.Require().NoError(err)
	suite.Require().Len(res.UpcomingStreams, 0)

	// setup 20 more upcoming streams with pool denom
	for i := 0; i < 20; i++ {
		_, stream, _, _ := suite.SetupNewStreamWithDenom(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)}, "pool")
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 10 created streams from upcoming to active
		// this leaves 10 upcoming streams
		if i < 10 {
			suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
		}
	}

	// query upcoming streams by lptoken denom with a page request of 5 should return 0 streams
	res, err = suite.querier.UpcomingStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsPerDenomRequest{Denom: "lptoken", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.UpcomingStreams, 0)

	// query upcoming streams by pool denom with a page request of 5 should return 5 streams
	res, err = suite.querier.UpcomingStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.UpcomingStreams, 5)

	// query upcoming streams by pool denom with a page request of 15 should return 10 streams
	res, err = suite.querier.UpcomingStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().Len(res.UpcomingStreams, 10)
}

// TestGRPCToDistributeCoins tests querying coins that are going to be distributed via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCToDistributeCoins() {
	suite.SetupTest()

	// ensure initially querying to distribute coins returns no coins
	res, err := suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create two locks with different durations
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a non perpetual stream
	streamID, _, coins, startTime := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	stream, err := suite.querier.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	suite.Require().NotNil(stream)
	streams := []types.Stream{*stream}

	// check to distribute coins after stream creation
	// ensure this equals the coins within the previously created non perpetual stream
	res, err = suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	// distribute coins to stakers
	distrCoins, err := suite.querier.Distribute(suite.Ctx, streams)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check stream changes after distribution
	// ensure the stream's filled epochs have been increased by 1
	// ensure we have distributed 4 out of the 10 stake tokens
	stream, err = suite.querier.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	suite.Require().NotNil(stream)
	suite.Require().Equal(stream.FilledEpochs, uint64(1))
	suite.Require().Equal(stream.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})
	streams = []types.Stream{*stream}

	// move stream from an upcoming to an active status
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	err = suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// check that the to distribute coins is equal to the initial stream coin balance minus what has been distributed already (10-4=6)
	res, err = suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins.Sub(distrCoins...))

	// distribute second round to stakers
	distrCoins, err = suite.querier.Distribute(suite.Ctx, streams)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 6)}, distrCoins)

	// now that all coins have been distributed (4 in first found 6 in the second round)
	// to distribute coins should be null
	res, err = suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))
}

// TestGRPCDistributedCoins tests querying coins that have been distributed via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCDistributedCoins() {
	suite.SetupTest()

	// create two locks with different durations
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a non perpetual stream
	streamID, _, _, startTime := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	stream, err := suite.querier.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	suite.Require().NotNil(stream)
	streams := []types.Stream{*stream}

	// move stream from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	err = suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.querier.Distribute(suite.Ctx, streams)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check stream changes after distribution
	// ensure the stream's filled epochs have been increased by 1
	// ensure we have distributed 4 out of the 10 stake tokens
	stream, err = suite.querier.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	suite.Require().NotNil(stream)
	suite.Require().Equal(stream.FilledEpochs, uint64(1))
	suite.Require().Equal(stream.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})
	streams = []types.Stream{*stream}

	// distribute second round to stakers
	distrCoins, err = suite.querier.Distribute(suite.Ctx, streams)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 6)}, distrCoins)
}

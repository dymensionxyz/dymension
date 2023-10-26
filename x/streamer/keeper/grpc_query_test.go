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
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	streamID, _ := suite.CreateDefaultStream(coins)

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
		Id:                   streamID,
		DistributeTo:         defaultDestAddr.String(),
		Coins:                coins,
		StartTime:            time.Time{},
		DistrEpochIdentifier: "day",
		NumEpochsPaidOver:    30,
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins{},
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
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	streamID, _ := suite.CreateDefaultStream(coins)

	// query streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.Streams(sdk.WrapSDKContext(suite.Ctx), &types.StreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := types.Stream{
		Id:                   streamID,
		DistributeTo:         defaultDestAddr.String(),
		Coins:                coins,
		StartTime:            time.Time{},
		DistrEpochIdentifier: "day",
		NumEpochsPaidOver:    30,
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins{},
	}
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// create 10 more streams
	for i := 0; i < 10; i++ {
		// create a stream
		coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
		_, _ = suite.CreateDefaultStream(coins)
		// suite.Ctx = suite.Ctx.WithBlockTime(time.Now.Add(time.Second))
	}

	// check that setting page request limit to 10 will only return 10 out of the 11 streams
	filter := query.PageRequest{Limit: 10}
	res, err = suite.querier.Streams(sdk.WrapSDKContext(suite.Ctx), &types.StreamsRequest{Pagination: &filter})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 10)

	filter = query.PageRequest{Limit: 13}
	res, err = suite.querier.Streams(sdk.WrapSDKContext(suite.Ctx), &types.StreamsRequest{Pagination: &filter})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 11)
}

// TestGRPCActiveStreams tests querying active streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCActiveStreams() {
	suite.SetupTest()

	// ensure initially querying active streams returns no streams
	res, err := suite.querier.ActiveStreams(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a stream and move it from upcoming to active
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
	streamID, stream := suite.CreateDefaultStream(coins)
	err = suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// query active streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.ActiveStreams(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := types.Stream{
		Id:                   streamID,
		DistributeTo:         defaultDestAddr.String(),
		Coins:                coins,
		StartTime:            time.Time{},
		DistrEpochIdentifier: "day",
		NumEpochsPaidOver:    30,
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins{},
	}
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// create 20 more streams
	for i := 0; i < 20; i++ {
		coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
		_, stream := suite.CreateDefaultStream(coins)

		// move the first 9 streams from upcoming to active (now 10 active streams, 30 total streams)
		if i < 9 {
			suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
		}
	}

	// set page request limit to 5, expect only 5 active stream responses
	res, err = suite.querier.ActiveStreams(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsRequest{Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 5)

	// set page request limit to 15, expect only 10 active stream responses
	res, err = suite.querier.ActiveStreams(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsRequest{Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().NoError(err)
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
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
	streamID, stream := suite.CreateDefaultStream(coins)
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// query streams by denom again, but this time expect the stream created earlier in the response
	res, err = suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{Denom: "stake", Pagination: nil})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := types.Stream{
		Id:                streamID,
		Coins:             coins,
		NumEpochsPaidOver: 30,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         time.Time{},
	}
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// setup 20 more streams with the pool denom
	for i := 0; i < 20; i++ {
		_, stream := suite.CreateStream(defaultDestAddr, sdk.Coins{sdk.NewInt64Coin("pool", 3)}, time.Time{}, "day", 30)
		// suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 10 of 20 streams to an active status
		if i < 10 {
			suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
		}
	}

	// query active streams by lptoken denom with a page request of 5 should only return one stream
	res, err = suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{Denom: "lptoken", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)

	// query active streams by pool denom with a page request of 5 should return 5 streams
	res, err = suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 5)

	// query active streams by pool denom with a page request of 15 should return 10 streams
	res, err = suite.querier.ActiveStreamsPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveStreamsPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().NoError(err)
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
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
	streamID, _ := suite.CreateDefaultStream(coins)

	// query upcoming streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.UpcomingStreams(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := types.Stream{
		Id:                streamID,
		Coins:             coins,
		NumEpochsPaidOver: 30,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         time.Time{},
	}
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

	// setup 20 more upcoming streams
	for i := 0; i < 20; i++ {
		coins := sdk.Coins{sdk.NewInt64Coin("stake", 3)}
		_, stream := suite.CreateDefaultStream(coins)
		// suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 9 created streams to an active status
		// 1 + (20 -9) = 12 upcoming streams
		if i < 9 {
			suite.querier.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
		}
	}

	// query upcoming streams with a page request of 5 should return 5 streams
	res, err = suite.querier.UpcomingStreams(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsRequest{Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 5)

	// query upcoming streams with a page request of 15 should return 12 streams
	res, err = suite.querier.UpcomingStreams(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingStreamsRequest{Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 12)
}

// TestGRPCToDistributeCoins tests querying coins that are going to be distributed via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCToDistributeCoins() {
	suite.SetupTest()

	// ensure initially querying to distribute coins returns no coins
	res, err := suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create two locks with different durations
	// addr1 := sdk.AccAddress([]byte("addr1---------------"))
	// addr2 := sdk.AccAddress([]byte("addr2---------------"))

	// setup a non perpetual stream
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	streamID, _ := suite.CreateDefaultStream(coins)

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
	// suite.Ctx = suite.Ctx.WithBlockTime(startTime)
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

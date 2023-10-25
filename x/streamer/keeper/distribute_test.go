package keeper_test

import (
	"strings"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

// TestDistribute tests that when the distribute command is executed on a provided stream
// that the correct amount of rewards is sent to the correct lock owners.
func (suite *KeeperTestSuite) TestDistribute() {
	defaultStream := perpStreamDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	doubleLengthStream := perpStreamDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: 2 * defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	noRewardStream := perpStreamDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{},
	}
	noRewardCoins := sdk.Coins{}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 2000)}
	fiveKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 5000)}
	tests := []struct {
		name            string
		users           []userLocks
		streams         []perpStreamDesc
		expectedRewards []sdk.Coins
	}{
		// stream 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should go to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one lockup, another user with two lockups, single default stream",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			streams:         []perpStreamDesc{defaultStream},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// stream 1 gives 3k coins. three locks, all eligible.
		// stream 2 gives 3k coins. one lock, to twoLockupUser.
		// 1k should to oneLockupUser and 5k to twoLockupUser.
		{
			name:            "One user with one lockup (default stream), another user with two lockups (double length stream)",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			streams:         []perpStreamDesc{defaultStream, doubleLengthStream},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
		// stream 1 gives zero rewards.
		// both oneLockupUser and twoLockupUser should get no rewards.
		{
			name:            "One user with one lockup, another user with two lockups, both with no rewards stream",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			streams:         []perpStreamDesc{noRewardStream},
			expectedRewards: []sdk.Coins{noRewardCoins, noRewardCoins},
		},
		// stream 1 gives no rewards.
		// stream 2 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one lockup and another user with two lockups. No rewards and a default stream",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			streams:         []perpStreamDesc{noRewardStream, defaultStream},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
	}
	for _, tc := range tests {
		suite.SetupTest()
		// setup streams and the locks defined in the above tests, then distribute to them
		streams := suite.SetupStreams(tc.streams, defaultLPDenom)
		addrs := suite.SetupUserLocks(tc.users)
		_, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, streams)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
			suite.Require().Equal(tc.expectedRewards[i].String(), bal.String(), "test %v, person %d", tc.name, i)
		}
	}
}

// TestSyntheticDistribute tests that when the distribute command is executed on a provided stream
// the correct amount of rewards is sent to the correct synthetic lock owners.
func (suite *KeeperTestSuite) TestSyntheticDistribute() {
	defaultStream := perpStreamDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	doubleLengthStream := perpStreamDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: 2 * defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	noRewardStream := perpStreamDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{},
	}
	noRewardCoins := sdk.Coins{}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 2000)}
	fiveKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 5000)}
	tests := []struct {
		name            string
		users           []userLocks
		streams         []perpStreamDesc
		expectedRewards []sdk.Coins
	}{
		// stream 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should go to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup, another user with two synthetic lockups, both with default stream",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			streams:         []perpStreamDesc{defaultStream},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// stream 1 gives 3k coins. three locks, all eligible.
		// stream 2 gives 3k coins. one lock, to twoLockupUser.
		// 1k should to oneLockupUser and 5k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup (default stream), another user with two synthetic lockups (double length stream)",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			streams:         []perpStreamDesc{defaultStream, doubleLengthStream},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
		// stream 1 gives zero rewards.
		// both oneLockupUser and twoLockupUser should get no rewards.
		{
			name:            "One user with one synthetic lockup, another user with two synthetic lockups, both with no rewards stream",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			streams:         []perpStreamDesc{noRewardStream},
			expectedRewards: []sdk.Coins{noRewardCoins, noRewardCoins},
		},
		// stream 1 gives no rewards.
		// stream 2 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup (no rewards stream), another user with two synthetic lockups (default stream)",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			streams:         []perpStreamDesc{noRewardStream, defaultStream},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
	}
	for _, tc := range tests {
		suite.SetupTest()
		// setup streams and the synthetic locks defined in the above tests, then distribute to them
		streams := suite.SetupStreams(tc.streams, defaultLPSyntheticDenom)
		addrs := suite.SetupUserSyntheticLocks(tc.users)
		_, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, streams)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			var rewards string
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
			// extract the superbonding tokens from the rewards distribution
			// TODO: figure out a less hacky way of doing this
			if strings.Contains(bal.String(), "lptoken/superbonding,") {
				rewards = strings.Split(bal.String(), "lptoken/superbonding,")[1]
			}
			suite.Require().Equal(tc.expectedRewards[i].String(), rewards, "test %v, person %d", tc.name, i)
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
	_, streamID, streamCoins, startTime := suite.SetupLockAndStream(false)

	// check that the sum of coins yet to be distributed is equal to the newly created streamCoins
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins)

	// add coins to the previous stream and check that the sum of coins yet to be distributed includes these new coins
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToStream(addCoins, streamID)
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins.Add(addCoins...))

	// create a new stream
	// check that the sum of coins yet to be distributed is equal to the stream1 and stream2 coins combined
	_, _, streamCoins2, _ := suite.SetupNewStream(false, sdk.Coins{sdk.NewInt64Coin("stake", 1000)})
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins.Add(addCoins...).Add(streamCoins2...))

	// move all created streams from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, []types.Stream{*stream})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check stream changes after distribution
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins.Add(addCoins...).Add(streamCoins2...).Sub(distrCoins...))
}

// TestGetModuleDistributedCoins tests that the sum of coins that have been distributed so far for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleDistributedCoins() {
	suite.SetupTest()

	// check that the sum of coins yet to be distributed is nil
	coins := suite.App.StreamerKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup a non perpetual lock and stream
	_, streamID, _, startTime := suite.SetupLockAndStream(false)

	// check that the sum of coins yet to be distributed is equal to the newly created streamCoins
	coins = suite.App.StreamerKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// move all created streams from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, []types.Stream{*stream})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 5)})

	// check stream changes after distribution
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, distrCoins)
}

// TestNoLockPerpetualStreamDistribution tests that the creation of a perp stream that has no locks associated does not distribute any tokens.
func (suite *KeeperTestSuite) TestNoLockPerpetualStreamDistribution() {
	suite.SetupTest()

	// setup a perpetual stream with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	streamID, _, _, startTime := suite.SetupNewStream(true, coins)

	// ensure the created stream has not completed distribution
	streams := suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 1)

	// ensure the not finished stream matches the previously created stream
	expectedStream := types.Stream{
		Id:          streamID,
		IsPerpetual: true,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(streams[0].String(), expectedStream.String())

	// move the created stream from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, []types.Stream{*stream})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	streams = suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 1)
	suite.Require().Equal(streams[0].String(), expectedStream.String())
}

// TestNoLockNonPerpetualStreamDistribution tests that the creation of a non perp stream that has no locks associated does not distribute any tokens.
func (suite *KeeperTestSuite) TestNoLockNonPerpetualStreamDistribution() {
	suite.SetupTest()

	// setup non-perpetual stream with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	streamID, _, _, startTime := suite.SetupNewStream(false, coins)

	// ensure the created stream has not completed distribution
	streams := suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 1)

	// ensure the not finished stream matches the previously created stream
	expectedStream := types.Stream{
		Id:          streamID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(streams[0].String(), expectedStream.String())

	// move the created stream from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, []types.Stream{*stream})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	streams = suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
	suite.Require().Len(streams, 1)
	suite.Require().Equal(streams[0].String(), expectedStream.String())
}

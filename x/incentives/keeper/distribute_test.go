package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

var _ = suite.TestingSuite(nil)

// TestDistribute tests that when the distribute command is executed on a provided gauge
// that the correct amount of rewards is sent to the correct lock owners.
func (suite *KeeperTestSuite) TestDistribute() {
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 300000000000000000)},
	}
	doubleLengthGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: 2 * defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 300000000000000000)},
	}
	noRewardGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{},
	}
	noRewardCoins := sdk.Coins{}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 100000000000000000)}
	twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 200000000000000000)}
	fiveKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 500000000000000000)}
	tests := []struct {
		name            string
		users           []userLocks
		gauges          []perpGaugeDesc
		expectedRewards []sdk.Coins
	}{
		// gauge 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should go to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one lockup, another user with two lockups, single default gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// gauge 1 gives 3k coins. three locks, all eligible.
		// gauge 2 gives 3k coins. one lock, to twoLockupUser.
		// 1k should to oneLockupUser and 5k to twoLockupUser.
		{
			name:            "One user with one lockup (default gauge), another user with two lockups (double length gauge)",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge, doubleLengthGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
		// gauge 1 gives zero rewards.
		// both oneLockupUser and twoLockupUser should get no rewards.
		{
			name:            "One user with one lockup, another user with two lockups, both with no rewards gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge},
			expectedRewards: []sdk.Coins{noRewardCoins, noRewardCoins},
		},
		// gauge 1 gives no rewards.
		// gauge 2 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one lockup and another user with two lockups. No rewards and a default gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge, defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
	}
	for _, tc := range tests {
		suite.SetupTest()
		// setup gauges and the locks defined in the above tests, then distribute to them
		gauges := suite.SetupGauges(tc.gauges, defaultLPDenom)
		addrs := suite.SetupUserLocks(tc.users)
		suite.Ctx = suite.Ctx.WithBlockTime(time.Now())
		_, err := suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, gauges)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
			suite.Require().Equal(tc.expectedRewards[i].String(), bal.String(), "test %v, person %d", tc.name, i)
		}
	}
}

// TestGetModuleToDistributeCoins tests the sum of coins yet to be distributed for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	suite.SetupTest()

	// check that the sum of coins yet to be distributed is nil
	coins := suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins{})

	// setup a non perpetual lock and gauge
	_, gaugeID, gaugeCoins, startTime := suite.SetupLockAndGauge(false)

	// check that the sum of coins yet to be distributed is equal to the newly created gaugeCoins
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins)

	// add coins to the previous gauge and check that the sum of coins yet to be distributed includes these new coins
	addCoins := sdk.Coins{sdk.NewInt64Coin("adym", 200000000000000000)}
	suite.AddToGauge(addCoins, gaugeID)
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...))

	// create a new gauge
	// check that the sum of coins yet to be distributed is equal to the gauge1 and gauge2 coins combined
	_, _, gaugeCoins2, _ := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("adym", 1000)})
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...))

	// move all created gauges from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("adym", 150000000000000000)})

	// check gauge changes after distribution
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...).Sub(distrCoins...))
}

// TestGetModuleDistributedCoins tests that the sum of coins that have been distributed so far for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleDistributedCoins() {
	suite.SetupTest()

	// check that the sum of coins yet to be distributed is nil
	coins := suite.App.IncentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins{})

	// setup a non perpetual lock and gauge
	_, gaugeID, _, startTime := suite.SetupLockAndGauge(false)

	// check that the sum of coins yet to be distributed is equal to the newly created gaugeCoins
	coins = suite.App.IncentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins{})

	// move all created gauges from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("adym", 50000000000000000)})

	// check gauge changes after distribution
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, distrCoins)
}

// TestNoLockPerpetualGaugeDistribution tests that the creation of a perp gauge that has no locks associated does not distribute any tokens.
func (suite *KeeperTestSuite) TestNoLockPerpetualGaugeDistribution() {
	suite.SetupTest()

	// setup a perpetual gauge with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(true, coins)

	// ensure the created gauge has not completed distribution
	gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// ensure the not finished gauge matches the previously created gauge
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: true,
		DistributeTo: &types.Gauge_Asset{Asset: &lockuptypes.QueryCondition{
			Denom:    "lptoken",
			Duration: time.Second,
		}},
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())

	// move the created gauge from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{})

	// check state is same after distribution
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}

// TestNoLockNonPerpetualGaugeDistribution tests that the creation of a non perp gauge that has no locks associated does not distribute any tokens.
func (suite *KeeperTestSuite) TestNoLockNonPerpetualGaugeDistribution() {
	suite.SetupTest()

	// setup non-perpetual gauge with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(false, coins)

	// ensure the created gauge has not completed distribution
	gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// ensure the not finished gauge matches the previously created gauge
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: &types.Gauge_Asset{Asset: &lockuptypes.QueryCondition{
			Denom:    "lptoken",
			Duration: time.Second,
		}},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())

	// move the created gauge from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{})

	// check state is same after distribution
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}

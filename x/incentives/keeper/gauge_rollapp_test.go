package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// TestDistribute tests that when the distribute command is executed on a provided gauge
// that the correct amount of rewards is sent to the correct lock owners.
func (suite *KeeperTestSuite) TestDistributeToRollappGauges() {
	// defaultGauge := perpGaugeDesc{
	// 	lockDenom:    defaultLPDenom,
	// 	lockDuration: defaultLockDuration,
	// 	rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	// }
	// noRewardGauge := perpGaugeDesc{
	// 	lockDenom:    defaultLPDenom,
	// 	lockDuration: defaultLockDuration,
	// 	rewardAmount: sdk.Coins{},
	// }
	// testGauges := []perpGaugeDesc{defaultGauge, noRewardGauge}

	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	// twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 2000)}

	// testUsers := []userLocks{oneLockupUser, twoLockupUser}
	// expectedRewards := []sdk.Coins{oneKRewardCoins, twoKRewardCoins}

	// // setup gauges and the locks defined in the above tests, then distribute to them
	// gauges := suite.SetupGauges(testGauges, defaultLPDenom)
	// addrs := suite.SetupUserLocks(testUsers)

	testCases := []struct {
		name        string
		rewards     sdk.Coins
		noSequencer bool
	}{
		{
			name:    "rollapp gauge with sequencer",
			rewards: oneKRewardCoins,
		},
		{
			name:    "rollapp gauge with no rewards",
			rewards: sdk.Coins{},
		},
		{
			name:        "rollapp gauge with no sequencer",
			rewards:     oneKRewardCoins,
			noSequencer: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// create rollapp and check rollapp gauge created
			rollapp := suite.CreateDefaultRollapp()
			res, err := suite.querier.RollappGauges(sdk.WrapSDKContext(suite.Ctx), &types.GaugesRequest{})
			suite.Require().NoError(err)
			suite.Require().Len(res.Data, 1)

			var proposer string
			if !tc.noSequencer {
				proposer = suite.CreateDefaultSequencer(suite.Ctx, rollapp)
			}

			if tc.rewards.Len() > 0 {
				suite.AddToGauge(tc.rewards, res.Data[0].Id)
			}

			_, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, res.Data)
			suite.Require().NoError(err)
			// check expected rewards against actual rewards received
			if proposer != "" {
				bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sdk.AccAddress(proposer))
				suite.Require().Equal(tc.rewards.String(), bal.String())
			}
		})
	}
}

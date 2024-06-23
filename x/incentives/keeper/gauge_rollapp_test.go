package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// TestDistributeToRollappGauges tests distributing rewards to rollapp gauges.
func (suite *KeeperTestSuite) TestDistributeToRollappGauges() {
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
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

			gaugeId := res.Data[0].Id

			var proposerAddr sdk.AccAddress
			if !tc.noSequencer {
				addr := suite.CreateDefaultSequencer(suite.Ctx, rollapp)
				proposerAddr, _ = sdk.AccAddressFromBech32(addr)
			}

			if tc.rewards.Len() > 0 {
				suite.AddToGauge(tc.rewards, gaugeId)
			}

			gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
			suite.Require().NoError(err)
			_, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)
			// check expected rewards against actual rewards received
			if !proposerAddr.Empty() {
				bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, proposerAddr)
				suite.Require().Equal(tc.rewards.String(), bal.String())
			}
		})
	}
}

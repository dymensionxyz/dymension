package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// TestDistributeToRollappGauges tests distributing rewards to rollapp gauges.
func (suite *KeeperTestSuite) TestDistributeToRollappGauges() {
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	addrs := apptesting.CreateRandomAccounts(1)

	testCases := []struct {
		name         string
		rewards      sdk.Coins
		rollappOwner sdk.AccAddress
	}{
		{
			name:         "rollapp gauge",
			rewards:      oneKRewardCoins,
			rollappOwner: addrs[0],
		},
		{
			name:         "rollapp gauge with no rewards",
			rewards:      sdk.Coins{},
			rollappOwner: addrs[0],
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Create rollapp and check rollapp gauge created
			_ = suite.CreateDefaultRollapp(tc.rollappOwner)
			res, err := suite.querier.RollappGauges(sdk.WrapSDKContext(suite.Ctx), new(types.GaugesRequest))
			suite.Require().NoError(err)
			suite.Require().NotNil(res)
			suite.Require().Len(res.Data, 1)
			suite.Require().NotNil(res.Data[0].GetRollapp())

			gaugeId := res.Data[0].Id

			// Top up the gauge before distributing
			if tc.rewards.Len() > 0 {
				suite.AddToGauge(tc.rewards, gaugeId)
			}

			gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
			suite.Require().NoError(err)

			// Distribute to the rollapp owner
			_, err = suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)

			// Check expected rewards against actual rewards received
			ownerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, tc.rollappOwner)
			suite.Require().ElementsMatch(tc.rewards, ownerBalances, "expect: %v, actual: %v", tc.rewards, ownerBalances)
		})
	}
}

// TestDistributeToRollappGaugesAfterOwnerChange tests distributing rewards to rollapp gauges given the
// fact that the owner might change. The test:
// 1. Distributes rewards to the gauge with the initial owner
// 2. Changes the rollapp owner
// 3. Distributes rewards to the gauge with the final owner
// 4. Validates balances of both initial and final owners
func (suite *KeeperTestSuite) TestDistributeToRollappGaugesAfterOwnerChange() {
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	addrs := apptesting.CreateRandomAccounts(2)

	testCases := []struct {
		name         string
		rewards      sdk.Coins
		initialOwner sdk.AccAddress
		finalOwner   sdk.AccAddress
	}{
		{
			name:         "rollapp gauge",
			rewards:      oneKRewardCoins,
			initialOwner: addrs[0],
			finalOwner:   addrs[1],
		},
		{
			name:         "rollapp gauge with no rewards",
			rewards:      sdk.Coins{},
			initialOwner: addrs[0],
			finalOwner:   addrs[1],
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Create rollapp and check rollapp gauge created
			_ = suite.CreateDefaultRollapp(tc.initialOwner)
			res, err := suite.querier.RollappGauges(sdk.WrapSDKContext(suite.Ctx), new(types.GaugesRequest))
			suite.Require().NoError(err)
			suite.Require().NotNil(res)
			suite.Require().Len(res.Data, 1)
			suite.Require().NotNil(res.Data[0].GetRollapp())

			gaugeId := res.Data[0].Id
			rollappID := res.Data[0].GetRollapp().RollappId

			// Distribute to the initial owner
			{
				if tc.rewards.Len() > 0 {
					suite.AddToGauge(tc.rewards, gaugeId)
				}

				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
				suite.Require().NoError(err)

				_, err = suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge})
				suite.Require().NoError(err)
			}

			// Transfer rollapp ownership
			suite.TransferRollappOwnership(tc.initialOwner, tc.finalOwner, rollappID)

			// Distribute to the final owner
			{
				if tc.rewards.Len() > 0 {
					suite.AddToGauge(tc.rewards, gaugeId)
				}

				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
				suite.Require().NoError(err)

				_, err = suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge})
				suite.Require().NoError(err)
			}

			// Check expected rewards against actual rewards received
			initialOwnerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, tc.initialOwner)
			suite.Require().ElementsMatch(tc.rewards, initialOwnerBalances, "expect: %v, actual: %v", tc.rewards, initialOwnerBalances)

			finalOwnerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, tc.finalOwner)
			suite.Require().ElementsMatch(tc.rewards, finalOwnerBalances, "expect: %v, actual: %v", tc.rewards, finalOwnerBalances)
		})
	}
}

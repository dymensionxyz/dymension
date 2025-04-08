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
			res, err := suite.querier.RollappGauges(suite.Ctx, new(types.GaugesRequest))
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

// TestDistributeToRollappGaugesWithEndorsementModes tests distributing rewards to rollapp gauges with different endorsement modes
func (suite *KeeperTestSuite) TestDistributeToRollappGaugesWithEndorsementModes() {
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	addrs := apptesting.CreateRandomAccounts(2)
	endorsementModes := []types.Params_EndorsementMode{types.Params_ActiveOnly, types.Params_AllRollapps}

	for _, mode := range endorsementModes {
		// Your code here
		suite.SetupTest()

		// Create 1st rollapp and set it as launched
		rollapp_launched := suite.CreateDefaultRollapp(addrs[0])
		rollapp := suite.App.RollappKeeper.MustGetRollapp(suite.Ctx, rollapp_launched)
		rollapp.Launched = true
		suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

		// Create 2nd rollapp (not launched)
		_ = suite.CreateDefaultRollapp(addrs[1])

		res, err := suite.querier.RollappGauges(suite.Ctx, new(types.GaugesRequest))
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Len(res.Data, 2)
		suite.Require().NotNil(res.Data[0].GetRollapp())
		suite.Require().NotNil(res.Data[1].GetRollapp())
		gaugeId1 := res.Data[0].Id
		gaugeId2 := res.Data[1].Id

		// set endorsement mode
		params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
		params.EndorsementMode = mode
		suite.App.IncentivesKeeper.SetParams(suite.Ctx, params)

		// Top up the gauge before distributing
		suite.AddToGauge(oneKRewardCoins, gaugeId1)
		suite.AddToGauge(oneKRewardCoins, gaugeId2)
		gauge1, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId1)
		suite.Require().NoError(err)
		gauge2, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId2)
		suite.Require().NoError(err)
		_, err = suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge1, *gauge2})
		suite.Require().NoError(err)

		// Check if rewards were distributed to all rollapps
		activeOwnerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addrs[0])
		suite.Require().ElementsMatch(oneKRewardCoins, activeOwnerBalances, "active rollapp should receive rewards")

		nonActiveOwnerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addrs[1])
		if mode == types.Params_AllRollapps {
			suite.Require().ElementsMatch(oneKRewardCoins, nonActiveOwnerBalances, "inactive rollapp should receive rewards")
		} else {
			suite.Require().Empty(nonActiveOwnerBalances, "inactive rollapp should not receive rewards")
			// assert the non active gauge still holds it's coins
			gauge2, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId2)
			suite.Require().NoError(err)
			suite.Require().ElementsMatch(oneKRewardCoins, gauge2.GetCoins())
		}
	}
}

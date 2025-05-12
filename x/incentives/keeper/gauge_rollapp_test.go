package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// TestDistributeToRollappGauges tests distributing rewards to rollapp gauges.
func (suite *KeeperTestSuite) TestDistributeToRollappGauges() {
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}

	testCases := []struct {
		name    string
		rewards sdk.Coins
	}{
		{
			name:    "rollapp gauge",
			rewards: oneKRewardCoins,
		},
		{
			name:    "rollapp gauge with no rewards",
			rewards: sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Create rollapp and check rollapp gauge created
			rollapp_id := suite.CreateDefaultRollapp()
			rollapp := suite.App.RollappKeeper.MustGetRollapp(suite.Ctx, rollapp_id)
			rollappOwner := sdk.MustAccAddressFromBech32(rollapp.Owner)
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
			ownerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, rollappOwner)
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
			_ = suite.CreateDefaultRollappFrom(tc.initialOwner)
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

// TestDistributeToRollappGaugesWithRollappGaugesModes tests distributing rewards to rollapp gauges with different endorsement modes
func (suite *KeeperTestSuite) TestDistributeToRollappGaugesWithRollappGaugesModes() {
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	rollappGaugesModes := []types.Params_RollappGaugesModes{types.Params_ActiveOnly, types.Params_AllRollapps}

	for _, mode := range rollappGaugesModes {
		// Your code here
		suite.SetupTest()

		// Create 1st rollapp and set it as launched
		rollapp1_id, _ := suite.CreateDefaultRollappAndProposer()
		rollapp1 := suite.App.RollappKeeper.MustGetRollapp(suite.Ctx, rollapp1_id)
		owner1 := sdk.MustAccAddressFromBech32(rollapp1.Owner)

		// Create 2nd rollapp (not launched)
		owner2 := apptesting.CreateRandomAccounts(1)[0]
		_ = suite.CreateDefaultRollappFrom(owner2)

		// Create 3rd rollapp (launched, but not active)
		owner3 := apptesting.CreateRandomAccounts(1)[0]
		rollapp3_id := suite.CreateDefaultRollappFrom(owner3)
		seq := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollapp3_id)
		seq.Dishonor = 1000000000000000000
		suite.App.SequencerKeeper.SetSequencer(suite.Ctx, seq)

		res, err := suite.querier.RollappGauges(suite.Ctx, new(types.GaugesRequest))
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Len(res.Data, 3)
		suite.Require().NotNil(res.Data[0].GetRollapp())
		suite.Require().NotNil(res.Data[1].GetRollapp())
		suite.Require().NotNil(res.Data[2].GetRollapp())
		gaugeId1 := res.Data[0].Id
		gaugeId2 := res.Data[1].Id
		gaugeId3 := res.Data[2].Id

		// set endorsement mode
		params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
		params.RollappGaugesMode = mode
		suite.App.IncentivesKeeper.SetParams(suite.Ctx, params)

		// Top up the gauge before distributing
		suite.AddToGauge(oneKRewardCoins, gaugeId1)
		suite.AddToGauge(oneKRewardCoins, gaugeId2)
		suite.AddToGauge(oneKRewardCoins, gaugeId3)
		gauge1, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId1)
		suite.Require().NoError(err)
		gauge2, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId2)
		suite.Require().NoError(err)
		gauge3, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId3)
		suite.Require().NoError(err)
		_, err = suite.App.IncentivesKeeper.DistributeOnEpochEnd(suite.Ctx, []types.Gauge{*gauge1, *gauge2, *gauge3})
		suite.Require().NoError(err)

		// Check if rewards were distributed to all rollapps
		activeOwnerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, owner1)
		suite.Require().ElementsMatch(oneKRewardCoins, activeOwnerBalances, "active rollapp should receive rewards")

		nonLaunchedOwnerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, owner2)
		nonActiveOwnerBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, owner3)
		if mode == types.Params_AllRollapps {
			suite.Require().ElementsMatch(oneKRewardCoins, nonLaunchedOwnerBalances, "inactive rollapp should receive rewards")
			suite.Require().ElementsMatch(oneKRewardCoins, nonActiveOwnerBalances, "non active rollapp should receive rewards")
		} else {
			suite.Require().Empty(nonLaunchedOwnerBalances, "inactive rollapp should not receive rewards")
			// assert the non active gauge still holds it's coins
			gauge2, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId2)
			suite.Require().NoError(err)
			suite.Require().ElementsMatch(oneKRewardCoins, gauge2.GetCoins())
			gauge3, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId3)
			suite.Require().NoError(err)
			suite.Require().ElementsMatch(oneKRewardCoins, gauge3.GetCoins())
		}
	}
}

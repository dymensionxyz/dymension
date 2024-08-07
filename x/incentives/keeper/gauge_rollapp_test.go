package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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

			creatorPK := ed25519.GenPrivKey().PubKey()
			creatorAddr := sdk.AccAddress(creatorPK.Address())

			// Create rollapp and check rollapp gauge created
			_ = suite.CreateDefaultRollapp(creatorAddr)
			res, err := suite.querier.RollappGauges(sdk.WrapSDKContext(suite.Ctx), new(types.GaugesRequest))
			suite.Require().NoError(err)
			suite.Require().NotNil(res)
			suite.Require().Len(res.Data, 1)

			gaugeId := res.Data[0].Id

			if tc.rewards.Len() > 0 {
				suite.AddToGauge(tc.rewards, gaugeId)
			}

			suite.App.RollappKeeper.UpdateRollapp(suite.Ctx, &rollapptypes.MsgUpdateRollappInformation{
				Creator:          "",
				RollappId:        "",
				InitialSequencer: "",
				Alias:            "",
				GenesisChecksum:  "",
				Metadata:         nil,
			})

			gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
			suite.Require().NoError(err)

			_, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)

			// Check expected rewards against actual rewards received
			creatorBalances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, creatorAddr)
			suite.Require().ElementsMatch(tc.rewards, creatorBalances)
		})
	}
}

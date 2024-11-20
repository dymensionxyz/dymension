package keeper_test

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// TODO: test slashing and Rewardee

func (suite *RollappTestSuite) TestSubmitRollappFraud() {
	numOfBlocks := uint64(10)
	initialHeight := uint64(1)

	testCases := []struct {
		name          string
		msgRevision   uint64
		msgHeight     uint64
		expectedError bool
	}{
		{
			name:          "first revision proposal",
			msgRevision:   0,
			msgHeight:     40,
			expectedError: false,
		},
		{
			name:          "future proposal",
			msgRevision:   2,
			msgHeight:     300,
			expectedError: false,
		},
		{
			name:          "wrong revision",
			msgRevision:   0,
			msgHeight:     80,
			expectedError: true,
		},
		{
			name:          "wrong future revision",
			msgRevision:   5,
			msgHeight:     300,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // Reset state for each test case
			suite.App.RollappKeeper.SetHooks(nil)

			suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialHeight))
			rollappId, proposer := suite.CreateDefaultRollappAndProposer()

			// set transferEnabled to true
			rollapp := suite.App.RollappKeeper.MustGetRollapp(suite.Ctx, rollappId)
			rollapp.GenesisState.TransfersEnabled = true
			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			var (
				lastHeight uint64 = 1
				err        error
			)

			// Create first batch of states (1-100)
			for i := int64(0); i < 10; i++ {
				suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeight() + i)
				lastHeight, err = suite.PostStateUpdate(suite.Ctx, rollappId, proposer, lastHeight, numOfBlocks)
				suite.Require().NoError(err)
			}

			// Force a fork at height 50
			err = suite.App.RollappKeeper.HardFork(suite.Ctx, rollappId, 50)
			suite.Require().NoError(err)
			lastHeight = 50

			// Create more states after fork (51-150)
			for i := int64(10); i < 15; i++ {
				suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeight() + i)
				lastHeight, err = suite.PostStateUpdateWithRevision(suite.Ctx, rollappId, proposer, lastHeight, numOfBlocks, 1)
				suite.Require().NoError(err)
			}

			// Force another fork at height 120
			err = suite.App.RollappKeeper.HardFork(suite.Ctx, rollappId, 120)
			suite.Require().NoError(err)

			// assert revision correctness
			rollapp = suite.App.RollappKeeper.MustGetRollapp(suite.Ctx, rollappId)
			suite.Require().EqualValues(rollapp.GetRevisionForHeight(1).Number, 0)
			suite.Require().EqualValues(rollapp.GetRevisionForHeight(49).Number, 0)
			suite.Require().EqualValues(rollapp.GetRevisionForHeight(50).Number, 1)
			suite.Require().EqualValues(rollapp.GetRevisionForHeight(55).Number, 1)
			suite.Require().EqualValues(rollapp.GetRevisionForHeight(120).Number, 2)
			suite.Require().EqualValues(rollapp.GetRevisionForHeight(300).Number, 2)

			msg := &types.MsgRollappFraudProposal{
				Authority:              suite.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName).String(),
				RollappId:              rollappId,
				FraudHeight:            tc.msgHeight,
				FraudRevision:          tc.msgRevision,
				PunishSequencerAddress: "",
				Rewardee:               "",
			}

			_, err = suite.keeper().SubmitRollappFraud(suite.Ctx, msg)

			if !tc.expectedError {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestDecreaseBond() {
	suite.SetupTest()
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId := suite.CreateDefaultRollapp()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)))
	// setup an unbonded sequencer
	unbondedSequencerAddress := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	unbondedSequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types.Unbonded
	suite.App.SequencerKeeper.UpdateSequencer(suite.Ctx, unbondedSequencer, unbondedSequencer.Status)
	// setup a jailed sequencer
	jailedSequencerAddress := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	jailedSequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, jailedSequencerAddress)
	jailedSequencer.Jailed = true
	suite.App.SequencerKeeper.UpdateSequencer(suite.Ctx, jailedSequencer, jailedSequencer.Status)

	testCase := []struct {
		name        string
		msg         types.MsgDecreaseBond
		expectedErr error
	}{
		{
			name: "invalid sequencer",
			msg: types.MsgDecreaseBond{
				Creator:        "invalid_address",
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
			expectedErr: types.ErrUnknownSequencer,
		},
		{
			name: "sequencer is not bonded",
			msg: types.MsgDecreaseBond{
				Creator:        unbondedSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
			expectedErr: types.ErrInvalidSequencerStatus,
		},
		{
			name: "sequencer is jailed",
			msg: types.MsgDecreaseBond{
				Creator:        jailedSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
			expectedErr: types.ErrSequencerJailed,
		},
		{
			name: "decreased bond value to less than minimum bond value",
			msg: types.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 100),
			},
			expectedErr: types.ErrInsufficientBond,
		},
		{
			name: "valid decrease bond",
			msg: types.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
		},
	}

	for _, tc := range testCase {
		suite.Run(tc.name, func() {
			resp, err := suite.msgServer.DecreaseBond(suite.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				suite.Require().ErrorIs(err, tc.expectedErr)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				expectedCompletionTime := suite.Ctx.BlockHeader().Time.Add(suite.App.SequencerKeeper.UnbondingTime(suite.Ctx))
				suite.Require().Equal(expectedCompletionTime, resp.CompletionTime)
				// check if the unbonding is set correctly
				unbondings := suite.App.SequencerKeeper.GetMatureDecreasingBondSequencers(suite.Ctx, expectedCompletionTime)
				suite.Require().Len(unbondings, 1)
				suite.Require().Equal(tc.msg.Creator, unbondings[0].SequencerAddress)
				suite.Require().Equal(tc.msg.DecreaseAmount, unbondings[0].UnbondAmount)
			}
		})
	}
}

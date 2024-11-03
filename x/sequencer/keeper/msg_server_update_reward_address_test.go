package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUpdateRewardAddress() {
	rollappId, pk := suite.CreateDefaultRollapp()
	defaultSequencerAddress := suite.CreateSequencer(suite.Ctx, rollappId, pk)
	rewardAddr := sample.AccAddress()

	testCase := []struct {
		name        string
		msg         types.MsgUpdateRewardAddress
		expectedErr error
	}{
		{
			name: "valid",
			msg: types.MsgUpdateRewardAddress{
				Creator:    defaultSequencerAddress,
				RewardAddr: rewardAddr,
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCase {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateRewardAddress(suite.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				suite.Require().ErrorIs(err, tc.expectedErr)
			} else {
				suite.Require().NoError(err)
				seq, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, tc.msg.Creator)
				suite.Require().Equal(tc.msg.RewardAddr, seq.RewardAddr)
			}
		})
	}
}

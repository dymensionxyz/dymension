package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestUpdateRewardAddress() {
	ra := s.createRollapp()
	seqAlice := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	rewardAddr := sample.AccAddress()

	testCase := []struct {
		name        string
		msg         types.MsgUpdateRewardAddress
		expectedErr error
	}{
		{
			name: "valid",
			msg: types.MsgUpdateRewardAddress{
				Creator:    seqAlice.Address,
				RewardAddr: rewardAddr,
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCase {
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateRewardAddress(s.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
			} else {
				s.Require().NoError(err)
				seq, _ := s.App.SequencerKeeper.RealSequencer(s.Ctx, tc.msg.Creator)
				s.Require().Equal(tc.msg.RewardAddr, seq.RewardAddr)
			}
		})
	}
}

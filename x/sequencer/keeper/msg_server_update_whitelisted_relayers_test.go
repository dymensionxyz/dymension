package keeper_test

import (
	"slices"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestUpdateWhitelistedRelayers() {
	ra := s.createRollapp()
	seqAlice := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	relayers := []string{sample.AccAddress(), sample.AccAddress()}

	testCase := []struct {
		name        string
		msg         types.MsgUpdateWhitelistedRelayers
		expectedErr error
	}{
		{
			name: "valid",
			msg: types.MsgUpdateWhitelistedRelayers{
				Creator:  seqAlice.Address,
				Relayers: relayers,
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCase {
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateWhitelistedRelayers(s.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
			} else {
				s.Require().NoError(err)
				seq, _ := s.App.SequencerKeeper.RealSequencer(s.Ctx, tc.msg.Creator)
				slices.Sort(tc.msg.Relayers)
				s.Require().Equal(tc.msg.Relayers, seq.WhitelistedRelayers)
			}
		})
	}
}

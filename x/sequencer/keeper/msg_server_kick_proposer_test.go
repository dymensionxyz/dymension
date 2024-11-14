package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
)

func (s *SequencerTestSuite) TestKickProposerBasicFlow() {
	s.App.RollappKeeper.SetHooks(nil)
	ra := s.createRollapp()
	seqAlice := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, seqAlice))

	_, err := s.PostStateUpdate(s.Ctx, ra.RollappId, seqAlice.Address, 1, 10)
	s.Require().NoError(err)

	// bob tries to kick alice but he doesn't have a sequencer
	m := &types.MsgKickProposer{Creator: pkAddr(bob)}
	_, err = s.msgServer.KickProposer(s.Ctx, m)
	utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)

	// bob creates a sequencer
	seqBob := s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, bond)
	// suppose he's unbonded
	seqBob.Status = types.Unbonded
	s.k().SetSequencer(s.Ctx, seqBob)

	// bob tries to kick alice but he's not bonded
	_, err = s.msgServer.KickProposer(s.Ctx, m)
	utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)

	// bob bonds, but alice is not below kick threshold
	seqBob.Status = types.Bonded
	s.k().SetSequencer(s.Ctx, seqBob)
	_, err = s.msgServer.KickProposer(s.Ctx, m)
	s.Require().Error(err)
	s.Require().True(s.k().IsProposer(s.Ctx, seqAlice))
	s.Require().False(s.k().IsProposer(s.Ctx, seqBob))

	// alice falls to threshold
	seqAlice.SetTokensCoin(kick)
	s.k().SetSequencer(s.Ctx, seqAlice)
	_, err = s.msgServer.KickProposer(s.Ctx, m)
	s.Require().NoError(err)
	s.Require().False(s.k().IsProposer(s.Ctx, seqAlice))

	// bob is now proposer
	s.Require().True(s.k().IsProposer(s.Ctx, seqBob))
	seqAlice = s.k().GetSequencer(s.Ctx, seqAlice.Address)
	s.Require().Equal(types.Unbonded, seqAlice.Status)
}

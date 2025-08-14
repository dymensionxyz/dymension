package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
)

func (s *SequencerTestSuite) TestKickProposerBasicFlow() {
	ra := s.createRollapp()
	seqAlice := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, seqAlice))
	s.submitAFewRollappStates(ra.RollappId)

	// bob tries to kick alice but he doesn't have a sequencer
	m := &types.MsgKickProposer{Creator: pkAddr(bob)}
	_, err := s.msgServer.KickProposer(s.Ctx, m)
	utest.IsErr(s.Require(), err, gerrc.ErrNotFound)

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
	seqAlice.SetPenalty(types.DefaultDishonorKickThreshold)
	s.k().SetSequencer(s.Ctx, seqAlice)
	_, err = s.msgServer.KickProposer(s.Ctx, m)
	s.Require().NoError(err)
	s.Require().False(s.k().IsProposer(s.Ctx, seqAlice))

	// bob is now proposer
	s.Require().True(s.k().IsProposer(s.Ctx, seqBob))
	seqAlice = s.k().GetSequencer(s.Ctx, seqAlice.Address)
	s.Require().Equal(types.Unbonded, seqAlice.Status)

	// alice can get tokens back (assuming no unfinalized states etc)
	s.k().SetUnbondBlockers()
	_, err = s.msgServer.Unbond(s.Ctx, &types.MsgUnbond{Creator: pkAddr(alice)})
	s.Require().NoError(err)
}

func (s *SequencerTestSuite) TestKickProposerSelfKickPrevented() {
	ra := s.createRollapp()
	seqAlice := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, seqAlice))
	
	// Submit some state updates to avoid hard fork issues
	s.submitAFewRollappStates(ra.RollappId)
	
	// Make alice kickable by setting penalty to threshold
	seqAlice.SetPenalty(types.DefaultDishonorKickThreshold)
	s.k().SetSequencer(s.Ctx, seqAlice)
	
	// Alice tries to kick herself (self-kick)
	m := &types.MsgKickProposer{Creator: pkAddr(alice)}
	_, err := s.msgServer.KickProposer(s.Ctx, m)
	
	// Self-kick should be prevented
	utest.IsErr(s.Require(), err, gerrc.ErrPermissionDenied)
	s.Require().Contains(err.Error(), "sequencer cannot kick itself")
	
	// Alice should still be the proposer
	s.Require().True(s.k().IsProposer(s.Ctx, seqAlice))
	
	// Alice should still be bonded
	seqAliceAfter := s.k().GetSequencer(s.Ctx, seqAlice.Address)
	s.Require().Equal(types.Bonded, seqAliceAfter.Status)
	
	// Create another sequencer (Bob) who can properly kick Alice
	seqBob := s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, bond)
	
	// Bob successfully kicks Alice
	mBob := &types.MsgKickProposer{Creator: pkAddr(bob)}
	_, err = s.msgServer.KickProposer(s.Ctx, mBob)
	s.Require().NoError(err)
	
	// Bob is now proposer
	s.Require().True(s.k().IsProposer(s.Ctx, seqBob))
	s.Require().False(s.k().IsProposer(s.Ctx, seqAlice))
}

package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
)

// Normal flow where there are two sequencers A, B and everything is graceful
func (s *SequencerTestSuite) TestRotationHappyFlow() {
	// init
	ra := s.createRollapp()
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))
	s.submitAFewRollappStates(ra.RollappId)

	// proposer tries to unbond
	mUnbond := &types.MsgUnbond{Creator: pkAddr(alice)}
	res, err := s.msgServer.Unbond(s.Ctx, mUnbond)
	s.Require().NoError(err)

	// notice period has not yet elapsed
	err = s.k().ChooseSuccessorForFinishedNotices(s.Ctx, s.Ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))

	// proposer cannot yet submit last
	err = s.k().OnProposerLastBlock(s.Ctx, s.seq(alice))
	utest.IsErr(s.Require(), err, gerrc.ErrFault)

	// advance clock past notice
	s.Require().True(res.GetNoticePeriodCompletionTime().After(s.Ctx.BlockTime()))
	s.Ctx = s.Ctx.WithBlockTime(*res.GetNoticePeriodCompletionTime())

	// notice period has now elapsed
	err = s.k().ChooseSuccessorForFinishedNotices(s.Ctx, s.Ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.seq(bob)))

	// proposer can submit last
	err = s.k().OnProposerLastBlock(s.Ctx, s.seq(alice))
	s.Require().NoError(err)
	s.Require().False(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(bob)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))
}

// Make sure a sequencer cannot be proposer twice
func (s *SequencerTestSuite) TestRotationReOptInFlow() {
	// init
	ra := s.createRollapp()
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, ucoin.SimpleMul(bond, 3))
	s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, ucoin.SimpleMul(bond, 2)) // bob has prio over charlie
	s.createSequencerWithBond(s.Ctx, ra.RollappId, charlie, ucoin.SimpleMul(bond, 1))
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))
	s.submitAFewRollappStates(ra.RollappId)

	prop := alice
	succ := bob

	for range 2 {
		// proposer tries to unbond
		mUnbond := &types.MsgUnbond{Creator: pkAddr(prop)}
		res, err := s.msgServer.Unbond(s.Ctx, mUnbond)
		s.Require().NoError(err)

		// notice period has not yet elapsed
		err = s.k().ChooseSuccessorForFinishedNotices(s.Ctx, s.Ctx.BlockTime())
		s.Require().NoError(err)
		s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(succ)))

		// proposer cannot yet submit last
		err = s.k().OnProposerLastBlock(s.Ctx, s.seq(prop))
		utest.IsErr(s.Require(), err, gerrc.ErrFault)

		// advance clock past notice
		s.Require().True(res.GetNoticePeriodCompletionTime().After(s.Ctx.BlockTime()))
		s.Ctx = s.Ctx.WithBlockTime(*res.GetNoticePeriodCompletionTime())

		// notice period has now elapsed
		err = s.k().ChooseSuccessorForFinishedNotices(s.Ctx, s.Ctx.BlockTime())
		s.Require().NoError(err)

		// proposer can submit last
		err = s.k().OnProposerLastBlock(s.Ctx, s.seq(prop))
		s.Require().NoError(err)
		s.Require().False(s.k().IsProposer(s.Ctx, s.seq(prop)))
		s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(succ)))

		// We can rotate Alice -> Bob but not Bob -> Alice
		s.Require().Equal(succ == bob, s.k().IsProposer(s.Ctx, s.seq(succ)))
		prop, succ = succ, prop
	}
}

// A wants to rotate but there is no B to take over. Proposer should be sentinel afterwards.
func (s *SequencerTestSuite) TestRotationNoSuccessor() {
	s.App.RollappKeeper.SetHooks(nil)
	// init
	ra := s.createRollapp()
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.k().SentinelSequencer(s.Ctx)))
	s.submitAFewRollappStates(ra.RollappId)

	// proposer tries to unbond
	mUnbond := &types.MsgUnbond{Creator: pkAddr(alice)}
	res, err := s.msgServer.Unbond(s.Ctx, mUnbond)
	s.Require().NoError(err)

	// advance clock past notice
	s.Require().True(res.GetNoticePeriodCompletionTime().After(s.Ctx.BlockTime()))
	s.Ctx = s.Ctx.WithBlockTime(*res.GetNoticePeriodCompletionTime())

	// notice period has now elapsed
	err = s.k().ChooseSuccessorForFinishedNotices(s.Ctx, s.Ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.k().SentinelSequencer(s.Ctx)))

	// proposer can submit last
	err = s.k().OnProposerLastBlock(s.Ctx, s.seq(alice))
	s.Require().NoError(err)
	s.Require().False(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().True(s.k().IsProposer(s.Ctx, s.k().SentinelSequencer(s.Ctx)))
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.k().SentinelSequencer(s.Ctx)))
}

// A wants to rotate. After B is marked successor he also wants to rotate, before A has finished.
func (s *SequencerTestSuite) TestRotationProposerAndSuccessorBothUnbond() {
	// init
	ra := s.createRollapp()
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, ucoin.SimpleMul(bond, 3))
	s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, ucoin.SimpleMul(bond, 2)) // bob has prio over charlie
	s.createSequencerWithBond(s.Ctx, ra.RollappId, charlie, ucoin.SimpleMul(bond, 1))
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))
	s.submitAFewRollappStates(ra.RollappId)

	// proposer tries to unbond
	mUnbond := &types.MsgUnbond{Creator: pkAddr(alice)}
	res, err := s.msgServer.Unbond(s.Ctx, mUnbond)
	s.Require().NoError(err)

	// advance clock past proposer notice
	s.Require().True(res.GetNoticePeriodCompletionTime().After(s.Ctx.BlockTime()))
	s.Ctx = s.Ctx.WithBlockTime(*res.GetNoticePeriodCompletionTime())

	// notice period has now elapsed
	err = s.k().ChooseSuccessorForFinishedNotices(s.Ctx, s.Ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.seq(bob)), "successor", s.k().GetSuccessor(s.Ctx, ra.RollappId).Address)

	// successor tries to unbond, but it fails
	mUnbond = &types.MsgUnbond{Creator: pkAddr(bob)}
	_, err = s.msgServer.Unbond(s.Ctx, mUnbond)
	utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)

	// proposer can submit last
	err = s.k().OnProposerLastBlock(s.Ctx, s.seq(alice))
	s.Require().NoError(err)
	s.Require().False(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(bob)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))

	// successor tries to unbond this time it works
	mUnbond = &types.MsgUnbond{Creator: pkAddr(bob)}
	_, err = s.msgServer.Unbond(s.Ctx, mUnbond)
	s.Require().NoError(err)
}

package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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

// A wants to rotate but there is no B to take over. Proposer should be sentinel afterwards.
func (s *SequencerTestSuite) TestRotationNoSuccessor() {
	// init
	ra := s.createRollapp()
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.k().SentinelSequencer(s.Ctx)))

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
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, bond)
	s.createSequencerWithBond(s.Ctx, ra.RollappId, charlie, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))

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
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.seq(bob)))

	// successor tries to unbond
	mUnbond = &types.MsgUnbond{Creator: pkAddr(bob)}
	res, err = s.msgServer.Unbond(s.Ctx, mUnbond)
	s.Require().NoError(err)

	// advance clock past successor notice
	s.Require().True(res.GetNoticePeriodCompletionTime().After(s.Ctx.BlockTime()))
	s.Ctx = s.Ctx.WithBlockTime(*res.GetNoticePeriodCompletionTime())

	// proposer can submit last
	err = s.k().OnProposerLastBlock(s.Ctx, s.seq(alice))
	s.Require().NoError(err)
	s.Require().False(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(bob)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))

	// notice period for original successor (bob) has now elapsed too
	err = s.k().ChooseSuccessorForFinishedNotices(s.Ctx, s.Ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(bob)))
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.seq(charlie)))

	// proposer can submit last
	err = s.k().OnProposerLastBlock(s.Ctx, s.seq(bob))
	s.Require().NoError(err)
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(charlie)))
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.k().SentinelSequencer(s.Ctx)))
}

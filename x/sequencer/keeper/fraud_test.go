package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

// Can eventually slash to below kickable threshold
func (s *SequencerTestSuite) TestSlashLivenessFlow() {
	ra := s.createRollapp()
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	seq := s.seq(alice)
	s.Require().True(s.k().IsProposer(s.Ctx, seq))

	s.Require().False(s.k().Kickable(s.Ctx, seq))
	ok := false
	last := seq.TokensCoin()
	for range 100000000 {
		err := s.k().SlashLiveness(s.Ctx, ra.RollappId)
		s.Require().NoError(err)
		s.Require().True(s.k().IsProposer(s.Ctx, seq))
		seq = s.seq(alice)
		mod := s.moduleBalance()
		s.Require().True(mod.Equal(seq.TokensCoin()))
		s.Require().True(seq.TokensCoin().IsLT(last))
		if s.k().Kickable(s.Ctx, seq) {
			ok = true
			break
		}
	}
	s.Require().True(ok)
}

// check the basic properties and that funds are allocated to the right place
func (s *SequencerTestSuite) TestFraud() {
	ra := s.createRollapp()

	s.Run("unbonded and not proposer anymore", func() {
		s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
		seq := s.seq(alice)

		s.k().SetProposer(s.Ctx, ra.RollappId, seq.Address)
		err := s.k().HandleFraud(s.Ctx, seq, nil)
		s.Require().NoError(err)

		seq = s.seq(alice)
		s.Require().False(s.k().IsProposer(s.Ctx, seq))
		s.Require().True(s.k().IsProposer(s.Ctx, s.k().SentinelSequencer(s.Ctx)))
		s.Require().False(seq.Bonded())
	})
	s.Run("without rewardee", func() {
		s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, bond)
		seq := s.seq(bob)

		s.k().SetProposer(s.Ctx, ra.RollappId, seq.Address)
		err := s.k().HandleFraud(s.Ctx, seq, nil)
		s.Require().NoError(err)

		seq = s.seq(bob)
		mod := s.moduleBalance()
		s.Require().True(seq.TokensCoin().IsZero())
		s.Require().True(mod.Equal(seq.TokensCoin()))
	})
	s.Run("with rewardee", func() {
		s.createSequencerWithBond(s.Ctx, ra.RollappId, charlie, bond)
		seq := s.seq(charlie)
		rewardee := pkAcc(randomTMPubKey())
		rewardeeBalBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardee)

		s.k().SetProposer(s.Ctx, ra.RollappId, seq.Address)
		err := s.k().HandleFraud(s.Ctx, seq, &rewardee)
		s.Require().NoError(err)

		seq = s.seq(charlie)
		mod := s.moduleBalance()
		s.Require().True(seq.TokensCoin().IsZero())
		s.Require().True(mod.Equal(seq.TokensCoin()))
		rewardeeBalAfter := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardee)
		s.Require().False(rewardeeBalAfter.IsEqual(rewardeeBalBefore))
	})
}

// a full flow 'e2e' to make sure things are sensible
// There are many many different scenarios that could be tested
// Here pick one which might be typical/realistic
// 1. Sequencer is active
// 2. Sequencer is does notice and starts to rotate
// 3. Sequencer does a fraud
// 4. Another sequencer opts in and becomes proposer
func (s *SequencerTestSuite) TestFraudFullFlowDuringRotation() {
	ra := s.createRollapp()
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, ucoin.SimpleMul(bond, 3))
	s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, ucoin.SimpleMul(bond, 2))
	s.createSequencerWithBond(s.Ctx, ra.RollappId, charlie, ucoin.SimpleMul(bond, 1))
	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))

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
	s.Require().True(s.k().IsSuccessor(s.Ctx, s.seq(bob)))

	// instead of submitting last, proposer does a fraud
	err = s.k().HandleFraud(s.Ctx, s.seq(alice), nil)
	s.Require().NoError(err)
	s.Require().False(s.k().IsProposer(s.Ctx, s.seq(alice)))
	s.Require().False(s.k().IsProposer(s.Ctx, s.seq(bob)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(bob)))
	s.Require().False(s.k().IsProposer(s.Ctx, s.seq(charlie)))
	s.Require().False(s.k().IsSuccessor(s.Ctx, s.seq(charlie)))
	s.Require().False(s.seq(bob).OptedIn)
	s.Require().False(s.seq(charlie).OptedIn)

	mOptIn := &types.MsgUpdateOptInStatus{Creator: pkAddr(bob), OptedIn: true}
	_, err = s.msgServer.UpdateOptInStatus(s.Ctx, mOptIn)
	s.Require().NoError(err)
	s.Require().True(s.seq(bob).OptedIn)

	s.Require().True(s.k().IsProposer(s.Ctx, s.seq(bob)))
}

package keeper_test

// Can eventually slash to below kickable threshold
func (s *SequencerTestSuite) TestSlashLivenessFlow() {
	ra := s.createRollapp()
	s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	seq := s.seq(alice)
	s.Require().True(s.k().IsProposer(s.Ctx, seq))

	s.Require().False(s.k().Kickable(s.Ctx, seq))
	ok := false
	for range 100000000 {
		err := s.k().SlashLiveness(s.Ctx, ra.RollappId)
		s.Require().NoError(err)
		s.Require().True(s.k().IsProposer(s.Ctx, seq))
		seq = s.seq(alice)
		mod := s.moduleBalance()
		s.Require().True(mod.Equal(seq.TokensCoin()))
		if s.k().Kickable(s.Ctx, seq) {
			ok = true
			break
		}
	}
	s.Require().True(ok)
}

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
		rewardee := pkAccAddr(randPK())
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

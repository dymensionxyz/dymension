package keeper_test

func (s *SequencerTestSuite) TestKickProposerBasic() {
	ra := s.createRollapp()
	seqAlice := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	seqBob := s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, seqAlice))
	s.Require().False(s.k().IsProposer(s.Ctx, seqBob))
}

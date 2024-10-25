package keeper_test

func (s *SequencerTestSuite) TestRotationHappyFlow() {
	ra := s.createRollapp()
	seqAlice := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	seqBob := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.Require().True(s.k().IsProposer(s.Ctx, seqAlice))
	s.Require().False(s.k().IsSuccessor(s.Ctx, seqBob))
}

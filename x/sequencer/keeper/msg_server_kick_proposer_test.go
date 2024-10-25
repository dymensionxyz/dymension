package keeper_test

func (s *SequencerTestSuite) TestKickProposerBasic() {
	ra := s.createRollapp()
	seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
}

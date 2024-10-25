package keeper_test

func (s *SequencerTestSuite) TestCreateSequencer() {
	ra := s.createRollapp()
	s.fundSequencer(alice, bond)
	msg := createSequencerMsg(ra.RollappId, alice)
	msg.Bond = bond
	_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
	s.Require().NoError(err)
}

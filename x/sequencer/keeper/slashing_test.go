package keeper_test

func (s *SequencerTestSuite) TestSlashBasic() {
	s.Run("slash at zero does not error", func() {
		// There shouldn't be an error if the sequencer has no tokens
		k := s.App.SequencerKeeper
		rollappId, pk := s.CreateDefaultRollapp()
		seqAddr := s.CreateDefaultSequencer(s.Ctx, rollappId, pk)
		seq, found := k.GetSequencer(s.Ctx, seqAddr)
		s.Require().True(found)
		err := k.Slash(s.Ctx, &seq, seq.Tokens)
		s.Require().NoError(err)
		err = k.Slash(s.Ctx, &seq, seq.Tokens)
		s.Require().NoError(err)
	})
}

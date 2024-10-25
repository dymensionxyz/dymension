package keeper_test

func (s *SequencerTestSuite) TestSlashBasic() {
	s.Run("slash at zero does not error", func() {
		// There shouldn't be an error if the sequencer has no tokens
		k := s.App.SequencerKeeper
		rollappId, pk := s.createRollapp()
		seqAddr := s.createSequencer(s.Ctx, rollappId, pk)
		seq, found := k.GetSequencer(s.Ctx, seqAddr)
		s.Require().True(found)
		err := k.Slash(s.Ctx, &seq, seq.Tokens)
		s.Require().NoError(err)
		err = k.Slash(s.Ctx, &seq, seq.Tokens)
		s.Require().NoError(err)
	})
}

func (s *SequencerTestSuite) TestJailUnknownSequencer() {
	s.createRollapp()
	keeper := s.App.SequencerKeeper

	err := keeper.JailSequencerOnFraud(s.Ctx, "unknown_sequencer")
	s.ErrorIs(err, types.ErrSequencerNotFound)
}

func (s *SequencerTestSuite) TestJailUnbondedSequencer() {
	keeper := s.App.SequencerKeeper
	s.Ctx = s.Ctx.WithBlockHeight(20)
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	rollappId, _ := s.CreateDefaultRollappAndProposer()
	seqAddr := s.CreateDefaultSequencer(s.Ctx, rollappId) // bonded non proposer

	// unbond the non-proposer
	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	res, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)
	unbondTime := res.GetUnbondingCompletionTime()
	keeper.UnbondAllMatureSequencers(s.Ctx, unbondTime.Add(1*time.Second))
	seq, found := keeper.GetSequencer(s.Ctx, seqAddr)
	s.Require().True(found)
	s.Equal(seq.Address, seqAddr)
	s.Equal(seq.Status, types.Unbonded)

	// jail the unbonded sequencer
	err = keeper.JailSequencerOnFraud(s.Ctx, seqAddr)
	s.ErrorIs(err, types.ErrInvalidSequencerStatus)
}

func (s *SequencerTestSuite) TestJailUnbondingSequencer() {
	keeper := s.App.SequencerKeeper
	s.Ctx = s.Ctx.WithBlockHeight(20)
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	rollappId, _ := s.CreateDefaultRollappAndProposer()
	seqAddr := s.CreateDefaultSequencer(s.Ctx, rollappId) // bonded non proposer

	// unbond the non-proposer
	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)
	seq, ok := keeper.GetSequencer(s.Ctx, seqAddr)
	s.Require().True(ok)
	s.Equal(seq.Status, types.Unbonding)

	// jail the unbonding sequencer
	err = keeper.JailSequencerOnFraud(s.Ctx, seqAddr)
	s.NoError(err)
	s.assertJailed(seqAddr)
}

func (s *SequencerTestSuite) TestJailProposerSequencer() {
	keeper := s.App.SequencerKeeper
	s.Ctx = s.Ctx.WithBlockHeight(20)
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	rollappId, proposer := s.CreateDefaultRollappAndProposer()
	err := keeper.JailSequencerOnFraud(s.Ctx, proposer)
	s.NoError(err)
	s.assertJailed(proposer)

	_, found := keeper.GetProposerLegacy(s.Ctx, rollappId)
	s.Require().False(found)
}

func (s *SequencerTestSuite) TestJailBondReducingSequencer() {
	keeper := s.App.SequencerKeeper
	s.Ctx = s.Ctx.WithBlockHeight(20)
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	rollappId, pk := s.createRollapp()
	seqAddr := s.createSequencerWithBond(s.Ctx, rollappId, pk, bond.AddAmount(sdk.NewInt(20)))

	reduceBondMsg := types.MsgDecreaseBond{Creator: seqAddr, DecreaseAmount: sdk.NewInt64Coin(bond.Denom, 10)}
	resp, err := s.msgServer.DecreaseBond(s.Ctx, &reduceBondMsg)
	s.Require().NoError(err)
	bondReductions := keeper.GetMatureDecreasingBondIDs(s.Ctx, resp.GetCompletionTime())
	s.Require().Len(bondReductions, 1)

	err = keeper.JailSequencerOnFraud(s.Ctx, seqAddr)
	s.NoError(err)

	bondReductions = keeper.GetMatureDecreasingBondIDs(s.Ctx, resp.GetCompletionTime())
	s.Require().Len(bondReductions, 0)
	s.assertJailed(seqAddr)
}

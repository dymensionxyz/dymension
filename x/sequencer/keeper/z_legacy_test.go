package keeper_test

/*



// create sequencer before genesisInfo is set
func (s *SequencerTestSuite) TestCreateSequencerBeforeGenesisInfoL() {
	goCtx := types.WrapSDKContext(s.Ctx)
	rollappId, pk := s.createRollappWithInitialSequencer()

	// mess up the genesisInfo
	rollapp := s.raK().MustGetRollapp(s.Ctx, rollappId)
	rollapp.GenesisInfo.Bech32Prefix = ""
	s.raK().SetRollapp(s.Ctx, rollapp)

	addr := types.AccAddress(pk.Address())
	err := testutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, types.NewCoins(bond))
	s.Require().NoError(err)

	pkAny, err := types3.NewAnyWithValue(pk)
	s.Require().Nil(err)
	sequencerMsg := types2.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata: types2.SequencerMetadata{
			Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
		},
	}

	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.Require().Error(err)

	// set the genesisInfo
	rollapp.GenesisInfo.Bech32Prefix = "rol"
	s.raK().SetRollapp(s.Ctx, rollapp)

	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.Require().NoError(err)
}

// create sequencer before prelaunch
func (s *SequencerTestSuite) TestCreateSequencerBeforePrelaunchL() {
	rollappId, pk := s.createRollappWithInitialSequencer()

	// set prelaunch time to the rollapp
	preLaunchTime := time.Now()
	rollapp := s.raK().MustGetRollapp(s.Ctx, rollappId)
	rollapp.PreLaunchTime = &preLaunchTime
	s.raK().SetRollapp(s.Ctx, rollapp)

	addr := types.AccAddress(pk.Address())
	err := testutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, types.NewCoins(bond))
	s.Require().NoError(err)

	pkAny, err := types3.NewAnyWithValue(pk)
	s.Require().Nil(err)
	sequencerMsg := types2.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata: types2.SequencerMetadata{
			Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
		},
	}

	_, err = s.msgServer.CreateSequencer(types.WrapSDKContext(s.Ctx), &sequencerMsg)
	s.Require().Error(err)

	s.Ctx = s.Ctx.WithBlockTime(preLaunchTime.Add(time.Second))
	_, err = s.msgServer.CreateSequencer(types.WrapSDKContext(s.Ctx), &sequencerMsg)
	s.Require().NoError(err)
}

func (s *SequencerTestSuite) TestInvariants() {
	numOfRollapps := 5
	numOfSequencers := 5

	var rollappToTest string

	// create rollapps and sequencers
	for i := 0; i < numOfRollapps; i++ {
		rollapp, pk := s.createRollappWithInitialSequencer()

		// create sequencers
		seqAddr := make([]string, numOfSequencers)
		seqAddr[0] = s.createSequencerWithPk(s.Ctx, rollapp, pk)
		for j := 1; j < numOfSequencers; j++ {
			pki := ed25519.GenPrivKey().PubKey()
			seqAddr[j] = s.createSequencerWithPk(s.Ctx, rollapp, pki)
		}

	}

	rollappid := rollappToTest

	// Test the test: make sure all status have entries
	seqBonded := s.k().RollappSequencersByStatus(s.Ctx, rollappid, types.Bonded)
	seqUnbonded := s.k().RollappSequencersByStatus(s.Ctx, rollappid, types.Unbonded)

	if len(seqBonded) == 0 || len(seqUnbonded) == 0 {
		s.T().Fatal("Test setup failed")
	}
	// additional rollapp with no sequencers
	s.createRollappWithInitialSequencer()

	msg, ok := keeper.AllInvariants(s.App.SequencerKeeper)(s.Ctx)
	s.Require().False(ok, msg)
}

*/

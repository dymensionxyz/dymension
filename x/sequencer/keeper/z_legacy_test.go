package keeper_test

/*
func (s *SequencerTestSuite) TestCreateSequencerL() {
	goCtx := types.WrapSDKContext(s.Ctx)

	// sequencersExpect is the expected result of query all
	sequencersExpect := []*types2.Sequencer{}

	// rollappSequencersExpect is a map from rollappId to a map of sequencer addresses list
	type rollappSequencersExpectKey struct {
		rollappId, sequencerAddress string
	}
	rollappSequencersExpect := make(map[rollappSequencersExpectKey]string)
	rollappExpectedProposers := make(map[string]string)

	const numRollapps = 3
	rollappIDs := make([]string, numRollapps)
	// for 3 rollapps, test 10 sequencers creations
	for j := 0; j < numRollapps; j++ {
		rollapp := types4.Rollapp{
			RollappId: urand.RollappID(),
			Owner:     aliceAddr,
			Launched:  true,
			Metadata: &types4.RollappMetadata{
				Website:     "https://dymension.xyz",
				Description: "Sample description",
				LogoUrl:     "https://dymension.xyz/logo.png",
				Telegram:    "https://t.me/rolly",
				X:           "https://x.dymension.xyz",
			},
			GenesisInfo: types4.GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "1234567890abcdefg",
				InitialSupply:   types.NewInt(1000),
				NativeDenom: types4.DenomMetadata{
					Display:  "DEN",
					Base:     "aden",
					Exponent: 18,
				},
			},
		}
		s.raK().SetRollapp(s.Ctx, rollapp)

		rollappId := rollapp.GetRollappId()
		rollappIDs[j] = rollappId

		for i := 0; i < 10; i++ {
			pubkey := ed25519.GenPrivKey().PubKey()
			addr := types.AccAddress(pubkey.Address())
			err := testutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, types.NewCoins(bond))
			s.Require().NoError(err)
			pkAny, err := types3.NewAnyWithValue(pubkey)
			s.Require().Nil(err)

			// sequencer is the sequencer to create
			sequencerMsg := types2.MsgCreateSequencer{
				Creator:      addr.String(),
				DymintPubKey: pkAny,
				Bond:         bond,
				RollappId:    rollappId,
				Metadata: types2.SequencerMetadata{
					Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			}
			// sequencerExpect is the expected result of creating a sequencer
			sequencerExpect := types2.Sequencer{
				Address:      sequencerMsg.GetCreator(),
				DymintPubKey: sequencerMsg.GetDymintPubKey(),
				Status:       types2.Bonded,
				RollappId:    rollappId,
				Tokens:       types.NewCoins(bond),
				Metadata:     sequencerMsg.GetMetadata(),
			}

			// create sequencer
			createResponse, err := s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
			s.Require().Nil(err)
			s.Require().EqualValues(types2.MsgCreateSequencerResponse{}, *createResponse)

			// query the specific sequencer
			queryResponse, err := s.queryClient.Sequencer(goCtx, &types2.QueryGetSequencerRequest{
				SequencerAddress: sequencerMsg.GetCreator(),
			})
			s.Require().Nil(err)
			s.equalSequencer(&sequencerExpect, &queryResponse.Sequencer)

			// add the sequencer to the list of get all expected list
			sequencersExpect = append(sequencersExpect, &sequencerExpect)

			if i == 0 {
				rollappExpectedProposers[rollappId] = sequencerExpect.Address
			}

			sequencersRes, totalRes := getAll(s)
			s.Require().EqualValues(len(sequencersExpect), totalRes)
			// verify that query all contains all the sequencers that were created
			s.verifyAll(sequencersExpect, sequencersRes)

			// add the sequencer to the list of specific rollapp
			rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerExpect.Address}] = sequencerExpect.Address
		}
	}

	totalFound := 0
	// check query by rollapp
	for i := 0; i < numRollapps; i++ {
		rollappId := rollappIDs[i]
		queryAllResponse, err := s.queryClient.SequencersByRollapp(goCtx,
			&types2.QueryGetSequencersByRollappRequest{RollappId: rollappId})
		s.Require().Nil(err)
		// verify that all the addresses of the rollapp are found
		for _, sequencer := range queryAllResponse.Sequencers {
			s.Require().EqualValues(rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencer.Address}],
				sequencer.Address)
		}
		totalFound += len(queryAllResponse.Sequencers)

		// check that the first sequencer created is the active sequencer
		proposer, err := s.queryClient.GetProposerByRollapp(goCtx,
			&types2.QueryGetProposerByRollappRequest{RollappId: rollappId})
		s.Require().Nil(err)
		s.Require().EqualValues(proposer.ProposerAddr, rollappExpectedProposers[rollappId])
	}
	s.Require().EqualValues(totalFound, len(rollappSequencersExpect))
}



func (s *SequencerTestSuite) TestCreateSequencerInitialSequencerAsProposerL() {
	const alex = "dym1te3lcav5c2jn8tdcrhnyl8aden6lglw266kcdd"

	type sequencer struct {
		creatorName string
		expProposer bool
	}
	testCases := []struct {
		name,
		rollappInitialSeq string
		sequencers []sequencer
		malleate   func(rollappID string)
		expErr     error
	}{
		{
			name:              "Single initial sequencer is the first proposer",
			sequencers:        []sequencer{{creatorName: "alex", expProposer: true}},
			rollappInitialSeq: alex,
		}, {
			name:              "Two sequencers - one is the proposer",
			sequencers:        []sequencer{{creatorName: "alex", expProposer: true}, {creatorName: "bob", expProposer: false}},
			rollappInitialSeq: fmt.Sprintf("%s,%s", aliceAddr, alex),
		}, {
			name:              "One sequencer - failed because no initial sequencer",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: aliceAddr,
			expErr:            types2.ErrNotInitialSequencer,
		}, {
			name:              "Any sequencer can be the first proposer",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: true}, {creatorName: "steve", expProposer: false}},
			rollappInitialSeq: "*",
		}, {
			name:              "success - any sequencer can be the first proposer, rollapp launched",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: aliceAddr,
			malleate: func(rollappID string) {
				r, _ := s.raK().GetRollapp(s.Ctx, rollappID)
				r.Launched = true
				s.raK().SetRollapp(s.Ctx, r)
			},
			expErr: nil,
		}, {
			name:              "success - no initial sequencer, rollapp launched",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: "*",
			malleate: func(rollappID string) {
				r, _ := s.raK().GetRollapp(s.Ctx, rollappID)
				r.Launched = true
				s.raK().SetRollapp(s.Ctx, r)
			},
			expErr: nil,
		},
	}

	for _, tc := range testCases {

		goCtx := types.WrapSDKContext(s.Ctx)
		rollappId := s.createRollapp(tc.rollappInitialSeq)

		if tc.malleate != nil {
			tc.malleate(rollappId)
		}

		for _, seq := range tc.sequencers {
			addr, pk := sample.AccFromSecret(seq.creatorName)
			pkAny, _ := types3.NewAnyWithValue(pk)

			err := testutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, types.NewCoins(bond))
			s.Require().NoError(err)

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
			s.Require().ErrorIs(err, tc.expErr, tc.name)

			if tc.expErr != nil {
				return
			}

			// check that the sequencer is the proposer
			proposer := s.k().GetProposer(s.Ctx, rollappId)
			if seq.expProposer {
				s.Require().Equal(addr.String(), proposer.Address, tc.name)
			} else {
				s.Require().NotEqual(addr.String(), proposer.Address, tc.name)
			}
		}
	}
}



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

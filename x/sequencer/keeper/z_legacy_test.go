package keeper_test

/*

func (s *SequencerTestSuite) TestMinBondL() {
	testCases := []struct {
		name          string
		requiredBond  types.Coin
		bond          types.Coin
		expectedError error
	}{
		{
			name:          "Valid bond",
			requiredBond:  bond,
			bond:          bond,
			expectedError: nil,
		},
		{
			name:          "Insufficient bond",
			requiredBond:  bond,
			bond:          types.NewCoin(bond.Denom, bond.Amount.Quo(types.NewInt(2))),
			expectedError: types2.ErrInsufficientBond,
		},
		{
			name:          "wrong bond denom",
			requiredBond:  bond,
			bond:          types.NewCoin("nonbonddenom", bond.Amount),
			expectedError: types2.ErrInvalidDenom,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			seqParams := types2.DefaultParams()
			seqParams.MinBond = tc.requiredBond
			s.k().SetParams(s.Ctx, seqParams)

			rollappId, pk := s.createRollappWithInitialSequencer()

			// fund account
			addr := types.AccAddress(pk.Address())
			pkAny, err := types3.NewAnyWithValue(pk)
			s.Require().Nil(err)
			err = testutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, types.NewCoins(tc.bond))
			s.Require().Nil(err)

			sequencerMsg1 := types2.MsgCreateSequencer{
				Creator:      addr.String(),
				DymintPubKey: pkAny,
				Bond:         tc.bond,
				RollappId:    rollappId,
				Metadata: types2.SequencerMetadata{
					Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			}

			_, createErr := s.msgServer.CreateSequencer(s.Ctx, &sequencerMsg1)

			if tc.expectedError != nil {
				s.Require().ErrorAs(createErr, &tc.expectedError, tc.name)
			} else {
				s.Require().NoError(createErr)
				sequencer, err := s.k().GetRealSequencer(s.Ctx, addr.String())
				s.Require().NoError(err)
				if tc.requiredBond.IsNil() {
					s.Require().True(sequencer.Tokens.IsZero(), tc.name)
				} else {
					s.Require().Equal(types.NewCoins(tc.requiredBond), sequencer.Tokens, tc.name)
				}
			}
		})
	}
}

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

func (s *SequencerTestSuite) TestCreateSequencerAlreadyExistsL() {
	goCtx := types.WrapSDKContext(s.Ctx)

	rollappId, pk := s.createRollappWithInitialSequencer()
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
	s.Require().Nil(err)

	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.EqualError(err, types2.ErrSequencerAlreadyExists.Error())

	// unbond the sequencer
	unbondMsg := types2.MsgUnbond{Creator: addr.String()}
	_, err = s.msgServer.Unbond(goCtx, &unbondMsg)
	s.Require().NoError(err)

	// create the sequencer again, expect to fail anyway
	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.EqualError(err, types2.ErrSequencerAlreadyExists.Error())
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

func (s *SequencerTestSuite) TestCreateSequencerUnknownRollappIdL() {
	goCtx := types.WrapSDKContext(s.Ctx)

	pubkey := ed25519.GenPrivKey().PubKey()
	addr := types.AccAddress(pubkey.Address())
	err := testutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, types.NewCoins(bond))
	s.Require().NoError(err)

	pkAny, err := types3.NewAnyWithValue(pubkey)
	s.Require().Nil(err)
	sequencerMsg := types2.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    "rollappId",
		Metadata:     types2.SequencerMetadata{},
	}

	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	utest.IsErr(s.Require(), types4.ErrRollappNotFound, err)
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

func (s *SequencerTestSuite) TestIncreaseBondL() {
	rollappId, pk0 := s.createRollappWithInitialSequencer()
	// setup a default sequencer
	defaultSequencerAddress := s.createSequencerWithPk(s.Ctx, rollappId, pk0)
	// setup an unbonded sequencer
	pk1 := ed25519.GenPrivKey().PubKey()
	unbondedSequencerAddress := s.createSequencerWithPk(s.Ctx, rollappId, pk1)
	unbondedSequencer, _ := s.k().GetRealSequencer(s.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types2.Unbonded
	s.k().SetSequencer(s.Ctx, unbondedSequencer)

	// fund all the sequencers which have been setup
	bondAmount := types.NewInt64Coin(types2.DefaultParams().MinBond.Denom, 100)
	err := testutil.FundAccount(s.App.BankKeeper, s.Ctx, types.MustAccAddressFromBech32(defaultSequencerAddress), types.NewCoins(bondAmount))
	s.Require().NoError(err)
	err = testutil.FundAccount(s.App.BankKeeper, s.Ctx, types.MustAccAddressFromBech32(unbondedSequencerAddress), types.NewCoins(bondAmount))
	s.Require().NoError(err)

	testCase := []struct {
		name        string
		msg         types2.MsgIncreaseBond
		expectedErr error
	}{
		{
			name: "valid",
			msg: types2.MsgIncreaseBond{
				Creator:   defaultSequencerAddress,
				AddAmount: bondAmount,
			},
			expectedErr: nil,
		},
		{
			name: "invalid sequencer",
			msg: types2.MsgIncreaseBond{
				Creator:   sample.AccAddress(), // a random address which is not a registered sequencer
				AddAmount: bondAmount,
			},
			expectedErr: types2.ErrSequencerNotFound,
		},
		{
			name: "sequencer doesn't have enough balance",
			msg: types2.MsgIncreaseBond{
				Creator:   defaultSequencerAddress,
				AddAmount: types.NewInt64Coin(types2.DefaultParams().MinBond.Denom, 99999999), // very high amount which sequencer doesn't have
			},
			expectedErr: errors.ErrInsufficientFunds,
		},
	}

	for _, tc := range testCase {
		s.Run(tc.name, func() {
			_, err := s.msgServer.IncreaseBond(s.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
			} else {
				s.Require().NoError(err)
				expectedBond := types2.DefaultParams().MinBond.Add(bondAmount)
				seq, _ := s.k().GetRealSequencer(s.Ctx, defaultSequencerAddress)
				s.Require().Equal(expectedBond, seq.Tokens[0])
			}
		})
	}
}

func (s *SequencerTestSuite) TestDecreaseBondL() {
	bondDenom := types2.DefaultParams().MinBond.Denom
	rollappId, pk := s.createRollappWithInitialSequencer()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := s.createSequencerWithBond(s.Ctx, rollappId, pk, bond.AddAmount(types.NewInt(20)))
	// setup an unbonded sequencer
	unbondedPk := ed25519.GenPrivKey().PubKey()
	unbondedSequencerAddress := s.createSequencerWithPk(s.Ctx, rollappId, unbondedPk)
	unbondedSequencer, _ := s.k().GetRealSequencer(s.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types2.Unbonded
	s.k().SetSequencer(s.Ctx, unbondedSequencer)

	testCase := []struct {
		name        string
		msg         types2.MsgDecreaseBond
		expectedErr error
	}{
		{
			name: "invalid sequencer",
			msg: types2.MsgDecreaseBond{
				Creator:        "invalid_address",
				DecreaseAmount: types.NewInt64Coin(bondDenom, 10),
			},
			expectedErr: types2.ErrSequencerNotFound,
		},
		{
			name: "decreased bond value to less than minimum bond value",
			msg: types2.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: types.NewInt64Coin(bondDenom, 100),
			},
			expectedErr: types2.ErrInsufficientBond,
		},
		{
			name: "trying to decrease more bond than they have tokens bonded",
			msg: types2.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: bond.AddAmount(types.NewInt(30)),
			},
			expectedErr: types2.ErrInsufficientBond,
		},
		{
			name: "valid decrease bond",
			msg: types2.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: types.NewInt64Coin(bondDenom, 10),
			},
		},
	}

	for _, tc := range testCase {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.DecreaseBond(s.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
			}
		})
	}
}

func (s *SequencerTestSuite) TestDecreaseBond_BondDecreaseInProgressL() {
	bondDenom := types2.DefaultParams().MinBond.Denom
	rollappId, pk := s.createRollappWithInitialSequencer()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := s.createSequencerWithBond(s.Ctx, rollappId, pk, bond.AddAmount(types.NewInt(20)))
	// decrease the bond of the sequencer
	_, err := s.msgServer.DecreaseBond(s.Ctx, &types2.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: types.NewInt64Coin(bondDenom, 10),
	})
	s.Require().NoError(err)
	// try to decrease the bond again - should be fine as still not below minbond
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1).WithBlockTime(s.Ctx.BlockTime().Add(10))
	_, err = s.msgServer.DecreaseBond(s.Ctx, &types2.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: types.NewInt64Coin(bondDenom, 10),
	})
	s.Require().NoError(err)
	// try to decrease the bond again - should err as below minbond
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1).WithBlockTime(s.Ctx.BlockTime().Add(10))
	_, err = s.msgServer.DecreaseBond(s.Ctx, &types2.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: types.NewInt64Coin(bondDenom, 10),
	})
	s.Require().ErrorIs(err, types2.ErrInsufficientBond)
}

func (s *SequencerTestSuite) TestUnbondingNonProposerL() {
	rollappId, pk := s.createRollappWithInitialSequencer()
	proposerAddr := s.createSequencerWithPk(s.Ctx, rollappId, pk)

	bondedAddr := s.CreateDefaultSequencer(s.Ctx, rollappId)
	s.Require().NotEqual(proposerAddr, bondedAddr)

	proposer := s.k().GetProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(proposerAddr, proposer.Address)

	// unbonded non proposer
	bondedSeq, err := s.k().GetRealSequencer(s.Ctx, bondedAddr)
	s.Require().True(found)
	s.Equal(types2.Bonded, bondedSeq.Status)

	unbondMsg := types2.MsgUnbond{Creator: bondedAddr}
	_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	// check sequencer operating status
	bondedSeq, err = s.k().GetRealSequencer(s.Ctx, bondedAddr)
	s.Require().True(found)
	s.Equal(types2.Unbonding, bondedSeq.Status)

	s.k().UnbondAllMatureSequencers(s.Ctx, bondedSeq.UnbondTime.Add(10*time.Second))
	bondedSeq, err = s.k().GetRealSequencer(s.Ctx, bondedAddr)
	s.Require().True(found)
	s.Equal(types2.Unbonded, bondedSeq.Status)

	// check proposer not changed
	proposer, ok = s.k().GetProposerLegacy(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(proposerAddr, proposer.Address)

	// try to unbond again. already unbonded, we expect error
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().Error(err)
}

func (s *SequencerTestSuite) TestUnbondingProposerL() {
	s.Ctx = s.Ctx.WithBlockHeight(10)

	rollappId, proposerAddr := s.CreateDefaultRollappAndProposer()
	_ = s.createSequencerWithPk(s.Ctx, rollappId, ed25519.GenPrivKey().PubKey())

	// unbond proposer
	unbondMsg := types2.MsgUnbond{Creator: proposerAddr}
	_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	// check proposer still bonded and notice period started
	p := s.k().GetProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(proposerAddr, p.Address)
	s.Equal(s.Ctx.BlockHeight(), p.UnbondRequestHeight)

	// unbonding again, we expect error as sequencer is in notice period
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().Error(err)

	// next proposer should not be set yet
	_, ok = s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().False(ok)

	// check notice period queue
	m, err := s.k().NoticeElapsedSequencers(s.Ctx, p.NoticePeriodTime.Add(-1*time.Second))
	s.Require().NoError(err)
	s.Require().Len(m, 0)
	m, err = s.k().NoticeElapsedSequencers(s.Ctx, p.NoticePeriodTime.Add(1*time.Second))
	s.Require().NoError(err)
	s.Require().Len(m, 1)
}

func (s *SequencerTestSuite) TestExpectedNextProposerL() {
	type testCase struct {
		name                    string
		numSeqAddrs             int
		expectEmptyNextProposer bool
	}

	testCases := []testCase{
		{"No additional sequencers", 0, true},
		{"few", 4, false},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rollappId, pk := s.createRollappWithInitialSequencer()
			_ = s.createSequencerWithBond(s.Ctx, rollappId, pk, bond) // proposer, with highest bond

			seqAddrs := make([]string, tc.numSeqAddrs)
			currBond := types.NewCoin(bond.Denom, bond.Amount.Quo(types.NewInt(10)))
			for i := 0; i < len(seqAddrs); i++ {
				currBond = currBond.AddAmount(bond.Amount)
				pubkey := ed25519.GenPrivKey().PubKey()
				seqAddrs[i] = s.createSequencerWithBond(s.Ctx, rollappId, pubkey, currBond)
			}
			next := s.k().ExpectedNextProposer(s.Ctx, rollappId)
			if tc.expectEmptyNextProposer {
				s.Require().Empty(next.Address)
				return
			}

			expectedNextProposer := seqAddrs[len(seqAddrs)-1]
			s.Equal(expectedNextProposer, next.Address)
		})
	}
}

// TestStartRotation tests the StartRotation function which is called when a sequencer has finished its notice period
func (s *SequencerTestSuite) TestStartRotationL() {
	rollappId, pk := s.createRollappWithInitialSequencer()
	addr1 := s.createSequencerWithPk(s.Ctx, rollappId, pk)

	_ = s.CreateDefaultSequencer(s.Ctx, rollappId)
	_ = s.CreateDefaultSequencer(s.Ctx, rollappId)


// ubnond proposer
	unbondMsg := types2.MsgUnbond{Creator: addr1}
	_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	// check proposer still bonded and notice period started
	p := s.k().GetProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(addr1, p.Address)
	s.Equal(s.Ctx.BlockHeight(), p.UnbondRequestHeight)

	m := s.k().GetMatureNoticePeriodSequencers(s.Ctx, p.NoticePeriodTime.Add(-10*time.Second))
	s.Require().Len(m, 0)
	m = s.k().GetMatureNoticePeriodSequencers(s.Ctx, p.NoticePeriodTime.Add(10*time.Second))
	s.Require().Len(m, 1)
	s.k().MatureSequencersWithNoticePeriod(s.Ctx, p.NoticePeriodTime.Add(10*time.Second))

	// validate nextProposer is set
	n, ok := s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Require().NotEmpty(n.Address)

	// validate proposer not changed
	p, _ = s.k().GetProposerLegacy(s.Ctx, rollappId)
	s.Equal(addr1, p.Address)
}

func (s *SequencerTestSuite) TestRotateProposerL() {
	rollappId, pk := s.createRollappWithInitialSequencer()
	addr1 := s.createSequencerWithPk(s.Ctx, rollappId, pk)
	addr2 := s.createSequencerWithPk(s.Ctx, rollappId, ed25519.GenPrivKey().PubKey())

	unbondMsg := types2.MsgUnbond{Creator: addr1}
	res, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	// mature notice period
	s.k().MatureSequencersWithNoticePeriod(s.Ctx, res.GetNoticePeriodCompletionTime().Add(10*time.Second))
	_, ok := s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().True(ok)

	// simulate lastBlock received
	err = s.k().completeRotationLeg(s.Ctx, rollappId)
	s.Require().NoError(err)

	// assert addr2 is now proposer
	p := s.k().GetProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(addr2, p.Address)
	// assert addr1 is unbonding
	u, _ := s.k().GetSequencer(s.Ctx, addr1)
	s.Equal(types2.Unbonding, u.Status)
	// assert nextProposer is nil
	_, ok = s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().False(ok)
}

func (s *SequencerTestSuite) TestRotateProposerNoNextProposerL() {
	rollappId, pk := s.createRollappWithInitialSequencer()
	addr1 := s.createSequencerWithPk(s.Ctx, rollappId, pk)

	unbondMsg := types2.MsgUnbond{Creator: addr1}
	res, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	// mature notice period
	s.k().MatureSequencersWithNoticePeriod(s.Ctx, res.GetNoticePeriodCompletionTime().Add(10*time.Second))
	// simulate lastBlock received
	err = s.k().completeRotationLeg(s.Ctx, rollappId)
	s.Require().NoError(err)

	_ := s.k().GetProposer(s.Ctx, rollappId)
	s.Require().False(ok)

	_, ok = s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().False(ok)
}

// Both the proposer and nextProposer tries to unbond
func (s *SequencerTestSuite) TestStartRotationTwiceL() {
	s.Ctx = s.Ctx.WithBlockHeight(10)

	rollappId, pk := s.createRollappWithInitialSequencer()
	addr1 := s.createSequencerWithPk(s.Ctx, rollappId, pk)
	addr2 := s.createSequencerWithPk(s.Ctx, rollappId, ed25519.GenPrivKey().PubKey())

	// unbond proposer
	unbondMsg := types2.MsgUnbond{Creator: addr1}
	_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	p := s.k().GetProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(addr1, p.Address)
	s.Equal(s.Ctx.BlockHeight(), p.UnbondRequestHeight)

	s.k().MatureSequencersWithNoticePeriod(s.Ctx, p.NoticePeriodTime.Add(10*time.Second))
	s.Require().True(s.k().isRotatingLeg(s.Ctx, rollappId))

	n, ok := s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(addr2, n.Address)

	// unbond nextProposer before rotation completes
	s.Ctx = s.Ctx.WithBlockHeight(20)
	unbondMsg = types2.MsgUnbond{Creator: addr2}
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	// check nextProposer is still the nextProposer and notice period started
	n, ok = s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(addr2, n.Address)
	s.Require().True(n.IsNoticePeriodInProgress())

	// rotation completes before notice period ends for addr2 (the nextProposer)
	err = s.k().completeRotationLeg(s.Ctx, rollappId) // simulate lastBlock received
	s.Require().NoError(err)

	// validate addr2 is now proposer and still with notice period
	p, _ = s.k().GetProposerLegacy(s.Ctx, rollappId)
	s.Equal(addr2, p.Address)
	s.Require().True(p.IsNoticePeriodInProgress())

	// validate nextProposer is unset after rotation completes
	n, ok = s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().False(ok)

	// mature notice period for addr2
	s.k().MatureSequencersWithNoticePeriod(s.Ctx, p.NoticePeriodTime.Add(10*time.Second))
	// validate nextProposer is set
	n, ok = s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Require().Empty(n.Address)
}

func (s *SequencerTestSuite) TestUnbondingMultiple() {
	s.Ctx = s.Ctx.WithBlockHeight(10)
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	keeper := s.App.SequencerKeeper

	rollappId, pk1 := s.createRollappWithInitialSequencer()
	rollappId2, pk2 := s.createRollappWithInitialSequencer()

	numOfSequencers := 4
	numOfSequencers2 := 3
	unbodingSeq := 2

	seqAddr1 := make([]string, numOfSequencers)
	seqAddr2 := make([]string, numOfSequencers2)

	// create 5 sequencers for rollapp1
	seqAddr1[0] = s.createSequencerWithPk(s.Ctx, rollappId, pk1)
	for i := 1; i < numOfSequencers; i++ {
		seqAddr1[i] = s.createSequencerWithPk(s.Ctx, rollappId, ed25519.GenPrivKey().PubKey())
	}

	// create 3 sequencers for rollapp2
	seqAddr2[0] = s.createSequencerWithPk(s.Ctx, rollappId2, pk2)
	for i := 1; i < numOfSequencers2; i++ {
		seqAddr2[i] = s.createSequencerWithPk(s.Ctx, rollappId2, ed25519.GenPrivKey().PubKey())
	}

	// start unbonding for 2 sequencers in each rollapp
	s.Ctx = s.Ctx.WithBlockHeight(20)
	now := time.Now()
	unbondTime := now.Add(keeper.GetParams(s.Ctx).UnbondingTime)
	s.Ctx = s.Ctx.WithBlockTime(now)
	for i := 1; i < unbodingSeq+1; i++ {
		unbondMsg := types2.MsgUnbond{Creator: seqAddr1[i]}
		_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
		s.Require().NoError(err)

		unbondMsg = types2.MsgUnbond{Creator: seqAddr2[i]}
		_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
		s.Require().NoError(err)
	}

	// before unbonding time reached
	sequencers := keeper.GetMatureUnbondingSequencers(s.Ctx, now)
	s.Require().Len(sequencers, 0)

	sequencers = keeper.GetMatureUnbondingSequencers(s.Ctx, unbondTime.Add(-1*time.Second))
	s.Require().Len(sequencers, 0)

	// past unbonding time
	sequencers = keeper.GetMatureUnbondingSequencers(s.Ctx, unbondTime.Add(1*time.Second))
	s.Require().Len(sequencers, 4)
}

func (s *SequencerTestSuite) TestTokensRefundOnUnbond() {
	denom := bond.Denom
	blockheight := 20
	var err error

	rollappId, pk := s.createRollappWithInitialSequencer()
	_ = s.createSequencerWithPk(s.Ctx, rollappId, pk)

	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := s.createSequencerWithPk(s.Ctx, rollappId, pk1)
	sequencer1, _ := s.k().GetSequencer(s.Ctx, addr1)
	s.Require().True(sequencer1.Status == types2.Bonded)
	s.Require().False(sequencer1.Tokens.IsZero())

	pk2 := ed25519.GenPrivKey().PubKey()
	addr2 := s.createSequencerWithPk(s.Ctx, rollappId, pk2)
	sequencer2, _ := s.k().GetSequencer(s.Ctx, addr2)
	s.Require().True(sequencer2.Status == types2.Bonded)
	s.Require().False(sequencer2.Tokens.IsZero())

	s.Ctx = s.Ctx.WithBlockHeight(int64(blockheight))
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	// start the 1st unbond
	unbondMsg := types2.MsgUnbond{Creator: addr1}
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)
	sequencer1, _ = s.k().GetSequencer(s.Ctx, addr1)
	s.Require().True(sequencer1.Status == types2.Unbonding)
	s.Require().Equal(sequencer1.UnbondRequestHeight, int64(blockheight))
	s.Require().False(sequencer1.Tokens.IsZero())

	// start the 2nd unbond later
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(5 * time.Minute))
	unbondMsg = types2.MsgUnbond{Creator: addr2}
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)
	sequencer2, _ = s.k().GetSequencer(s.Ctx, addr2)
	s.Require().True(sequencer2.Status == types2.Unbonding)
	s.Require().False(sequencer2.Tokens.IsZero())

	balanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, types.MustAccAddressFromBech32(addr1), denom)
	s.k().UnbondAllMatureSequencers(s.Ctx, sequencer1.UnbondTime.Add(1*time.Second))
	balanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, types.MustAccAddressFromBech32(addr1), denom)

	// Check stake refunded
	sequencer1, _ = s.k().GetSequencer(s.Ctx, addr1)
	s.Equal(types2.Unbonded, sequencer1.Status)
	s.True(sequencer1.Tokens.IsZero())
	s.True(balanceBefore.Add(bond).IsEqual(balanceAfter), "expected %s, got %s", balanceBefore.Add(bond), balanceAfter)

	// check the 2nd unbond still not happened
	sequencer2, _ = s.k().GetSequencer(s.Ctx, addr2)
	s.Equal(types2.Unbonding, sequencer2.Status)
	s.False(sequencer2.Tokens.IsZero())
}



func (s *SequencerTestSuite) TestSlashBasic() {
	s.Run("slash at zero does not error", func() {
		// There shouldn't be an error if the sequencer has no tokens
		k := s.App.SequencerKeeper
		rollappId, pk := s.createRollappWithInitialSequencer()
		seqAddr := s.createSequencerWithPk(s.Ctx, rollappId, pk)
		seq, found := k.GetSequencer(s.Ctx, seqAddr)
		s.Require().True(found)
		err := k.Slash(s.Ctx, &seq, seq.Tokens)
		s.Require().NoError(err)
		err = k.Slash(s.Ctx, &seq, seq.Tokens)
		s.Require().NoError(err)
	})
}

func (s *SequencerTestSuite) TestJailUnknownSequencer() {
	s.createRollappWithInitialSequencer()
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

	rollappId, pk := s.createRollappWithInitialSequencer()
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

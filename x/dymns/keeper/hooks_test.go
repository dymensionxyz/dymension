package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_rollappHooks_RollappCreated() {
	const price1L = 9
	const price2L = 8
	const price3L = 7
	const price4L = 6
	const price5L = 5
	const price6L = 4
	const price7PL = 3

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
		moduleParams.Price.AliasPriceSteps = []sdkmath.Int{
			sdk.NewInt(price1L).Mul(priceMultiplier),
			sdk.NewInt(price2L).Mul(priceMultiplier),
			sdk.NewInt(price3L).Mul(priceMultiplier),
			sdk.NewInt(price4L).Mul(priceMultiplier),
			sdk.NewInt(price5L).Mul(priceMultiplier),
			sdk.NewInt(price6L).Mul(priceMultiplier),
			sdk.NewInt(price7PL).Mul(priceMultiplier),
		}
		return moduleParams
	})
	s.SaveCurrentContext()

	creatorAccAddr := sdk.AccAddress(testAddr(1).bytes())
	dymNameOwnerAcc := testAddr(2)
	anotherAcc := testAddr(3)

	tests := []struct {
		name                    string
		addRollApps             []string
		preRunSetup             func(s *KeeperTestSuite)
		originalCreatorBalance  int64
		originalModuleBalance   int64
		rollAppId               string
		alias                   string
		wantErr                 bool
		wantErrContains         string
		wantSuccess             bool
		wantLaterCreatorBalance int64
		postTest                func(s *KeeperTestSuite)
	}{
		{
			name:                    "pass - register without problem",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 2,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 2,
		},
		{
			name:                   "pass - mapping RollApp ID <=> Alias should be set",
			addRollApps:            []string{"rollapp_1-1"},
			originalCreatorBalance: price5L,
			rollAppId:              "rollapp_1-1",
			alias:                  "alias",
			wantErr:                false,
			wantSuccess:            true,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.True(found)
				s.Equal("alias", alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_1-1", rollAppId)
			},
		},
		{
			name:                    "pass - if input alias is empty, do nothing",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  0,
			rollAppId:               "rollapp_1-1",
			alias:                   "",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 0,
			postTest: func(s *KeeperTestSuite) {
				// mapping should not be created

				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:                    "pass - Alias cost subtracted from creator and burned",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 10,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 1 char",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price1L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "a",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 2 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price2L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "ab",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 3 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price3L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "dog",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 4 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price4L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "pool",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 5 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "angel",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 6 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price6L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "bridge",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 7 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price7PL + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "academy",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 7+ chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price7PL + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "dymension",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "fail - RollApp not exists",
			addRollApps:             nil,
			originalCreatorBalance:  price1L,
			rollAppId:               "nad_0-0",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "not a RollApp chain-id",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
		},
		{
			name:                    "fail - mapping should not be created",
			addRollApps:             nil,
			originalCreatorBalance:  price1L,
			rollAppId:               "nad_0-0",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "not",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				_, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "nad_0-0")
				s.False(found)

				_, found = s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.False(found)
			},
		},
		{
			name:                    "fail - reject bad alias",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price1L,
			rollAppId:               "rollapp_1-1",
			alias:                   "@@@",
			wantErr:                 true,
			wantErrContains:         "invalid alias format",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
		},
		{
			name:        "pass - can register if alias is not used",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price5L + 2,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 2,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.True(found)
				s.Equal("alias", alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_1-1", rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is presents as chain-id in params",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "bridge",
							Aliases: []string{"b"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "bridge",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "bridge")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is presents as alias of a chain-id in params",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "dym",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "dym")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is a RollApp-ID",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.pureSetRollApp(rollapptypes.Rollapp{
					RollappId: "rollapp",
					Owner:     creatorAccAddr.String(),
				})

				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.Require().NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "rollapp",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "rollapp")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias used by another RollApp",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "alias")
				s.Require().NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_2-2", rollAppId)
			},
		},
		{
			name:                    "fail - reject if creator does not have enough balance to pay the fee",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  1,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "insufficient funds",
			wantSuccess:             false,
			wantLaterCreatorBalance: 1,
		},
		{
			name:                   "pass - can resolve address using alias",
			addRollApps:            []string{"rollapp_1-1"},
			preRunSetup:            nil,
			originalCreatorBalance: price5L,
			rollAppId:              "rollapp_1-1",
			alias:                  "alias",
			wantErr:                false,
			wantSuccess:            true,
			postTest: func(s *KeeperTestSuite) {
				dymName := dymnstypes.DymName{
					Name:       "my-name",
					Owner:      dymNameOwnerAcc.bech32(),
					Controller: dymNameOwnerAcc.bech32(),
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "rollapp_1-1",
							Value:   dymNameOwnerAcc.bech32(),
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "rollapp_1-1",
							Path:    "sub",
							Value:   anotherAcc.bech32(),
						},
					},
				}
				s.setDymNameWithFunctionsAfter(dymName)

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "my-name@rollapp_1-1")
				s.Require().NoError(err)
				s.Equal(dymNameOwnerAcc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "my-name@alias")
				s.Require().NoError(err)
				s.Equal(dymNameOwnerAcc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub.my-name@alias")
				s.Require().NoError(err)
				s.Equal(anotherAcc.bech32(), outputAddr)

				outputs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, anotherAcc.bech32(), "rollapp_1-1")
				s.Require().NoError(err)
				s.Require().NotEmpty(outputs)
				s.Equal("sub.my-name@alias", outputs[0].String())
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().NotEqual(tt.wantSuccess, tt.wantErr, "mis-configured test case")

			s.RefreshContext()

			if tt.originalCreatorBalance > 0 {
				s.mintToAccount2(creatorAccAddr.String(), sdk.NewInt(tt.originalCreatorBalance).Mul(priceMultiplier))
			}

			if tt.originalModuleBalance > 0 {
				s.mintToModuleAccount2(sdk.NewInt(tt.originalModuleBalance).Mul(priceMultiplier))
			}

			for _, rollAppId := range tt.addRollApps {
				s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
					RollappId: rollAppId,
					Owner:     creatorAccAddr.String(),
				})
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(s)
			}

			err := s.dymNsKeeper.GetRollAppHooks().RollappCreated(s.ctx, tt.rollAppId, tt.alias, creatorAccAddr)

			defer func() {
				if s.T().Failed() {
					return
				}

				laterModuleBalance := s.moduleBalance2()
				s.Equal(
					sdk.NewInt(tt.originalModuleBalance).Mul(priceMultiplier),
					laterModuleBalance,
					"module balance should not be changed regardless of success because of burn",
				)

				if tt.postTest != nil {
					tt.postTest(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				return
			}

			s.Require().NoError(err)

			laterCreatorBalance := s.balance2(creatorAccAddr.String())
			s.Equal(
				sdk.NewInt(tt.wantLaterCreatorBalance).Mul(priceMultiplier),
				laterCreatorBalance,
				"creator balance mismatch",
			)

			// event should be fired
			func() {
				if tt.alias == "" {
					return
				}

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

				for _, event := range events {
					if event.Type == dymnstypes.EventTypeSell {
						return
					}
				}

				s.T().Errorf("event %s not found", dymnstypes.EventTypeSell)
			}()
		})
	}

	s.Run("if alias is empty, do nothing", func() {
		originalTxGas := s.ctx.GasMeter().GasConsumed()

		err := s.dymNsKeeper.GetRollAppHooks().RollappCreated(s.ctx, "rollapp_1-1", "", creatorAccAddr)
		s.Require().NoError(err)

		s.Equal(originalTxGas, s.ctx.GasMeter().GasConsumed(), "should not consume gas")
	})
}

func (s *KeeperTestSuite) Test_rollappHooks_OnRollAppIdChanged() {
	const previousRollAppId = "rollapp_1-1"
	const newRollAppId = "rollapp_1-2"

	const name1 = "name1"
	const name2 = "name2"

	user1Acc := testAddr(1)
	user2Acc := testAddr(2)
	user3Acc := testAddr(3)
	user4Acc := testAddr(4)
	user5Acc := testAddr(5)

	genRollApp := func(rollAppId string, aliases ...string) *rollapp {
		ra := newRollApp(rollAppId)
		for _, alias := range aliases {
			_ = ra.WithAlias(alias)
		}
		return ra
	}

	tests := []struct {
		name      string
		setupFunc func(s *KeeperTestSuite)
		testFunc  func(s *KeeperTestSuite)
	}{
		{
			name: "pass - normal migration, with alias, with Dym-Name",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "alias"))
				s.persistRollApp(*genRollApp(newRollAppId))

				s.requireRollApp(previousRollAppId).HasAlias("alias")
				s.requireRollApp(newRollAppId).HasNoAlias()

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.NotEmpty(outputDymNameAddrs)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.ErrorContains(err, "not found")

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs)
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("alias")

				//

				_, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.ErrorContains(err, "not found")

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs)

				//

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.NotEmpty(outputDymNameAddrs)
			},
		},
		{
			name: "pass - normal migration, with multiple aliases, with multiple Dym-Names",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "one", "two", "three"))
				s.persistRollApp(*genRollApp(newRollAppId))

				s.requireRollApp(previousRollAppId).HasAlias("one", "two", "three")
				s.requireRollApp(newRollAppId).HasNoAlias()

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				s.setDymNameWithFunctionsAfter(
					newDN(name2, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						cfgN(previousRollAppId, "sub2", user2Acc.bech32()).
						cfgN(previousRollAppId, "sub3", user3Acc.bech32()).
						cfgN("", "", user4Acc.bech32()).
						cfgN("", "sub5", user5Acc.bech32()).
						build(),
				)

				//

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name2+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 3) // 1 from name1, 2 from name2

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub2."+name2+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub3."+name2+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user3Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user3Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name2+"@"+s.chainId)
				s.NoError(err)
				s.Equal(user4Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user4Acc.bech32(), s.chainId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user4Acc.bech32(), previousRollAppId)
				s.NoError(err)
				if s.Len(outputDymNameAddrs, 1) {
					s.Equal(name2+"@one", outputDymNameAddrs[0].String()) // result of fallback lookup
				}

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub5."+name2+"@"+s.chainId)
				s.NoError(err)
				s.Equal(user5Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user5Acc.bech32(), s.chainId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user5Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs) // no fallback lookup because it's a sub-name
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("one", "two", "three")

				//

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name2+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 3) // 1 from name1, 2 from name2

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub2."+name2+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub3."+name2+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user3Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user3Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name2+"@"+s.chainId)
				s.NoError(err)
				s.Equal(user4Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user4Acc.bech32(), s.chainId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user4Acc.bech32(), newRollAppId)
				s.NoError(err)
				if s.Len(outputDymNameAddrs, 1) {
					s.Equal(name2+"@one", outputDymNameAddrs[0].String()) // result of fallback lookup
				}

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub5."+name2+"@"+s.chainId)
				s.NoError(err)
				s.Equal(user5Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user5Acc.bech32(), s.chainId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user5Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs) // no fallback lookup because it's a sub-name
			},
		},
		{
			name: "fail - when migrate alias failed, should not change anything",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "alias"))
				s.persistRollApp(*genRollApp(newRollAppId))

				s.requireRollApp(previousRollAppId).HasAlias("alias")
				s.requireRollApp(newRollAppId).HasNoAlias()

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.NotEmpty(outputDymNameAddrs)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.ErrorContains(err, "not found")

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs)

				// delete the rollapp to make the migration fail
				s.rollAppKeeper.RemoveRollapp(s.ctx, previousRollAppId)

				s.Require().False(s.dymNsKeeper.IsRollAppId(s.ctx, previousRollAppId))
			},
			testFunc: func(s *KeeperTestSuite) {
				// unchanged

				s.requireRollApp(newRollAppId).HasNoAlias()

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.NotEmpty(outputDymNameAddrs)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.ErrorContains(err, "not found")

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs)
			},
		},
		{
			name: "pass - when the new RollApp has existing aliases, merge them",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "one", "two", "three"))
				s.persistRollApp(*genRollApp(newRollAppId, "four", "five", "six"))

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				//

				s.requireRollApp(previousRollAppId).HasAlias("one", "two", "three")
				s.requireRollApp(newRollAppId).HasAlias("four", "five", "six")

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("one", "two", "three", "four", "five", "six")

				// others

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)
			},
		},
		{
			name: "pass - when previous RollApp has no alias",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId))
				s.persistRollApp(*genRollApp(newRollAppId, "new"))

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				//

				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("new")

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("new")

				// others

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)
			},
		},
		{
			name: "pass - when the new RollApp has existing aliases, priority previous default alias",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "old", "journey"))
				s.persistRollApp(*genRollApp(newRollAppId, "new"))

				s.requireRollApp(previousRollAppId).HasAlias("old", "journey")
				s.requireRollApp(newRollAppId).HasAlias("new")
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAliasesWithOrder("old", "new", "journey")
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.setupFunc != nil {
				tt.setupFunc(s)
			}

			s.dymNsKeeper.GetFutureRollAppHooks().OnRollAppIdChanged(s.ctx, previousRollAppId, newRollAppId)

			if tt.testFunc != nil {
				tt.testFunc(s)
			}
		})
	}
}

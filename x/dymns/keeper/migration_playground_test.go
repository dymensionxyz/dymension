package keeper_test

import (
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

func (s *KeeperTestSuite) Test_MigrateStoreForPlayground() {
	rollApp1MultiAliases := *newRollApp("rollapp_1111-1").WithAlias("rol1").WithAlias("rol11")
	rollApp2 := *newRollApp("rollapp_2222-2").WithAlias("rol2")
	rollApp3WithoutAlias := *newRollApp("rollapp_3333-3")
	rollApp4WithoutAlias := *newRollApp("rollapp_4444-4")
	rollApp5WithoutAlias := *newRollApp("rollapp_5555-5")

	ownerAcc := testAddr(1)

	tests := []struct {
		name         string
		setupFunc    func(s *KeeperTestSuite)
		testFunc     func(s *KeeperTestSuite)
		wantFirstRun bool
	}{
		{
			name: "can migrate Dym-Name",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(rollApp1MultiAliases)

				dymName := newDN("a", ownerAcc.bech32()).
					cfgN(rollApp1MultiAliases.rollAppId, "", ownerAcc.bech32()).
					build()

				s.Require().Equal(rollApp1MultiAliases.rollAppId, dymName.Configs[0].ChainId)

				err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
				s.Require().NoError(err)
			},
			testFunc: func(s *KeeperTestSuite) {
				dymNameLater := s.dymNsKeeper.GetDymName(s.ctx, "a")
				s.Require().NotNil(dymNameLater)

				s.Require().Equal(
					dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp1MultiAliases.rollAppId),
					dymNameLater.Configs[0].ChainId,
				)
			},
			wantFirstRun: true,
		},
		{
			name: "can migrate multiple Dym-Names",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(rollApp1MultiAliases)
				s.persistRollApp(rollApp2)

				for _, dn := range []string{"a", "b"} {
					dymName := newDN(dn, ownerAcc.bech32()).
						cfgN(rollApp1MultiAliases.rollAppId, "", ownerAcc.bech32()).
						cfgN(rollApp2.rollAppId, "", ownerAcc.bech32()).
						cfgN("blumbus_100-1", "", ownerAcc.bech32()).
						build()

					s.Require().Equal(rollApp1MultiAliases.rollAppId, dymName.Configs[0].ChainId)
					s.Require().Equal(rollApp2.rollAppId, dymName.Configs[1].ChainId)

					err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
					s.Require().NoError(err)
				}
			},
			testFunc: func(s *KeeperTestSuite) {
				for _, dn := range []string{"a", "b"} {
					dymNameLater := s.dymNsKeeper.GetDymName(s.ctx, dn)
					s.Require().NotNil(dymNameLater)

					s.Equal(
						dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp1MultiAliases.rollAppId),
						dymNameLater.Configs[0].ChainId,
					)
					s.Equal(
						dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp2.rollAppId),
						dymNameLater.Configs[1].ChainId,
					)
					s.Equal(
						"blumbus_100-1",
						dymNameLater.Configs[2].ChainId,
						"non-RollApp should not be changed",
					)
				}
			},
			wantFirstRun: true,
		},
		{
			name: "can migrate mixed migrated-state Dym-Names",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(rollApp1MultiAliases)
				s.persistRollApp(rollApp2)

				for _, dn := range []string{"a", "b"} {
					dymName := newDN(dn, ownerAcc.bech32()).
						cfgN(rollApp1MultiAliases.rollAppId, "", ownerAcc.bech32()).
						cfgNForRollApp(rollApp2.rollAppId, "", ownerAcc.bech32()).
						cfgN("blumbus_100-1", "", ownerAcc.bech32()).
						build()

					s.Require().Equal(rollApp1MultiAliases.rollAppId, dymName.Configs[0].ChainId)
					s.Require().Equal(dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp2.rollAppId), dymName.Configs[1].ChainId)

					err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
					s.Require().NoError(err)
				}
			},
			testFunc: func(s *KeeperTestSuite) {
				for _, dn := range []string{"a", "b"} {
					dymNameLater := s.dymNsKeeper.GetDymName(s.ctx, dn)
					s.Require().NotNil(dymNameLater)

					s.Equal(
						dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp1MultiAliases.rollAppId),
						dymNameLater.Configs[0].ChainId,
					)
					s.Equal(
						dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp2.rollAppId),
						dymNameLater.Configs[1].ChainId,
					)
					s.Equal(
						"blumbus_100-1",
						dymNameLater.Configs[2].ChainId,
						"non-RollApp should not be changed",
					)
				}
			},
			wantFirstRun: true,
		},
		{
			name: "can migrate Dym-Name and accept nothing need migration",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(rollApp1MultiAliases)

				dymName := newDN("a", ownerAcc.bech32()).
					cfgNForRollApp(rollApp1MultiAliases.rollAppId, "", ownerAcc.bech32()).
					build()

				s.Require().Equal(
					dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp1MultiAliases.rollAppId),
					dymName.Configs[0].ChainId,
				)

				err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
				s.Require().NoError(err)
			},
			testFunc: func(s *KeeperTestSuite) {
				dymNameLater := s.dymNsKeeper.GetDymName(s.ctx, "a")
				s.Require().NotNil(dymNameLater)

				s.Require().Equal(
					dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp1MultiAliases.rollAppId),
					dymNameLater.Configs[0].ChainId,
				)
			},
			wantFirstRun: false, // no need to migrate
		},
		{
			name: "can migrate alias",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(rollApp3WithoutAlias)

				const alias = "rol3"

				store := s.ctx.KVStore(s.dymNsStoreKey)

				// simulate the legacy alias => RollApp (full) ID mapping
				store.Set(dymnstypes.AliasToBuyOrderIdsRvlKey(alias), []byte(rollApp3WithoutAlias.rollAppId))

				// simulate the legacy RollApp (full) ID => alias mapping
				store.Set(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp3WithoutAlias.rollAppId)...), s.cdc.MustMarshal(&dymnstypes.MultipleAliases{
					Aliases: []string{alias},
				}))
			},
			testFunc: func(s *KeeperTestSuite) {
				store := s.ctx.KVStore(s.dymNsStoreKey)

				const alias = "rol3"
				eip155 := dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp3WithoutAlias.rollAppId)

				s.Equal(eip155, string(store.Get(dymnstypes.AliasToRollAppEip155IdRvlKey(alias))))

				s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp3WithoutAlias.rollAppId)...)))

				s.True(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(eip155)...)))

				s.requireRollApp(rollApp3WithoutAlias.rollAppId).HasOnlyAlias(alias)
				s.requireAlias(alias).LinkedToRollApp(rollApp3WithoutAlias.rollAppId)
			},
			wantFirstRun: true,
		},
		{
			name: "can migrate multiple aliases",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(rollApp1MultiAliases)
				s.persistRollApp(rollApp2)
				s.persistRollApp(rollApp3WithoutAlias)
				s.persistRollApp(rollApp4WithoutAlias)
				s.persistRollApp(rollApp5WithoutAlias)

				{
					const alias3 = "rol3"

					store := s.ctx.KVStore(s.dymNsStoreKey)

					// simulate the legacy alias => RollApp (full) ID mapping
					store.Set(dymnstypes.AliasToBuyOrderIdsRvlKey(alias3), []byte(rollApp3WithoutAlias.rollAppId))

					// simulate the legacy RollApp (full) ID => alias mapping
					store.Set(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp3WithoutAlias.rollAppId)...), s.cdc.MustMarshal(&dymnstypes.MultipleAliases{
						Aliases: []string{alias3},
					}))
				}
				{
					const alias4 = "rol4"
					const alias42 = "rol42"

					store := s.ctx.KVStore(s.dymNsStoreKey)

					// simulate the legacy alias => RollApp (full) ID mapping
					store.Set(dymnstypes.AliasToBuyOrderIdsRvlKey(alias4), []byte(rollApp4WithoutAlias.rollAppId))
					store.Set(dymnstypes.AliasToBuyOrderIdsRvlKey(alias42), []byte(rollApp4WithoutAlias.rollAppId))

					// simulate the legacy RollApp (full) ID => alias mapping
					store.Set(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp4WithoutAlias.rollAppId)...), s.cdc.MustMarshal(&dymnstypes.MultipleAliases{
						Aliases: []string{alias4, alias42},
					}))
				}
			},
			testFunc: func(s *KeeperTestSuite) {
				store := s.ctx.KVStore(s.dymNsStoreKey)

				{
					s.requireRollApp(rollApp1MultiAliases.rollAppId).HasAliasesWithOrder("rol1", "rol11")
					s.requireAlias("rol1").LinkedToRollApp(rollApp1MultiAliases.rollAppId)
					s.requireAlias("rol11").LinkedToRollApp(rollApp1MultiAliases.rollAppId)
				}

				{
					s.requireRollApp(rollApp2.rollAppId).HasOnlyAlias("rol2")
					s.requireAlias("rol2").LinkedToRollApp(rollApp2.rollAppId)
				}

				{
					const alias3 = "rol3"
					eip155 := dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp3WithoutAlias.rollAppId)

					s.Equal(eip155, string(store.Get(dymnstypes.AliasToRollAppEip155IdRvlKey(alias3))))

					s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp3WithoutAlias.rollAppId)...)))

					s.True(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(eip155)...)))

					s.requireRollApp(rollApp3WithoutAlias.rollAppId).HasOnlyAlias(alias3)
					s.requireAlias(alias3).LinkedToRollApp(rollApp3WithoutAlias.rollAppId)
				}

				{
					const alias4 = "rol4"
					const alias42 = "rol42"
					eip155 := dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp4WithoutAlias.rollAppId)

					s.Equal(eip155, string(store.Get(dymnstypes.AliasToRollAppEip155IdRvlKey(alias4))))
					s.Equal(eip155, string(store.Get(dymnstypes.AliasToRollAppEip155IdRvlKey(alias42))))

					s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp4WithoutAlias.rollAppId)...)))

					s.True(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(eip155)...)))

					s.requireRollApp(rollApp4WithoutAlias.rollAppId).HasAliasesWithOrder(alias4, alias42)
					s.requireAlias(alias4).LinkedToRollApp(rollApp4WithoutAlias.rollAppId)
					s.requireAlias(alias42).LinkedToRollApp(rollApp4WithoutAlias.rollAppId)
				}

				{
					s.requireRollApp(rollApp5WithoutAlias.rollAppId).HasNoAlias()

					s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp5WithoutAlias.rollAppId)...)))

					eip155 := dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp5WithoutAlias.rollAppId)
					s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(eip155)...)))
				}
			},
			wantFirstRun: true,
		},
		{
			name: "can migrate mixed migrated-state aliases",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(rollApp1MultiAliases)
				s.persistRollApp(rollApp2)
				s.persistRollApp(rollApp3WithoutAlias)
				s.persistRollApp(rollApp4WithoutAlias)
				s.persistRollApp(rollApp5WithoutAlias)

				{
					const alias22 = "rol22"

					store := s.ctx.KVStore(s.dymNsStoreKey)

					// simulate the legacy alias => RollApp (full) ID mapping
					store.Set(dymnstypes.AliasToBuyOrderIdsRvlKey(alias22), []byte(rollApp2.rollAppId))

					// simulate the legacy RollApp (full) ID => alias mapping
					store.Set(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp2.rollAppId)...), s.cdc.MustMarshal(&dymnstypes.MultipleAliases{
						Aliases: []string{alias22},
					}))
				}

				{
					const alias3 = "rol3"

					store := s.ctx.KVStore(s.dymNsStoreKey)

					// simulate the legacy alias => RollApp (full) ID mapping
					store.Set(dymnstypes.AliasToBuyOrderIdsRvlKey(alias3), []byte(rollApp3WithoutAlias.rollAppId))

					// simulate the legacy RollApp (full) ID => alias mapping
					store.Set(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp3WithoutAlias.rollAppId)...), s.cdc.MustMarshal(&dymnstypes.MultipleAliases{
						Aliases: []string{alias3},
					}))
				}
				{
					const alias4 = "rol4"
					const alias42 = "rol42"

					store := s.ctx.KVStore(s.dymNsStoreKey)

					// simulate the legacy alias => RollApp (full) ID mapping
					store.Set(dymnstypes.AliasToBuyOrderIdsRvlKey(alias4), []byte(rollApp4WithoutAlias.rollAppId))
					store.Set(dymnstypes.AliasToBuyOrderIdsRvlKey(alias42), []byte(rollApp4WithoutAlias.rollAppId))

					// simulate the legacy RollApp (full) ID => alias mapping
					store.Set(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp4WithoutAlias.rollAppId)...), s.cdc.MustMarshal(&dymnstypes.MultipleAliases{
						Aliases: []string{alias4, alias42},
					}))
				}
			},
			testFunc: func(s *KeeperTestSuite) {
				store := s.ctx.KVStore(s.dymNsStoreKey)

				{
					s.requireRollApp(rollApp1MultiAliases.rollAppId).HasAliasesWithOrder("rol1", "rol11")
					s.requireAlias("rol1").LinkedToRollApp(rollApp1MultiAliases.rollAppId)
					s.requireAlias("rol11").LinkedToRollApp(rollApp1MultiAliases.rollAppId)
				}

				{
					s.requireRollApp(rollApp2.rollAppId).HasAliasesWithOrder("rol2", "rol22")
					s.requireAlias("rol2").LinkedToRollApp(rollApp2.rollAppId)
					s.requireAlias("rol22").LinkedToRollApp(rollApp2.rollAppId)
				}

				{
					const alias3 = "rol3"
					eip155 := dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp3WithoutAlias.rollAppId)

					s.Equal(eip155, string(store.Get(dymnstypes.AliasToRollAppEip155IdRvlKey(alias3))))

					s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp3WithoutAlias.rollAppId)...)))

					s.True(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(eip155)...)))

					s.requireRollApp(rollApp3WithoutAlias.rollAppId).HasOnlyAlias(alias3)
					s.requireAlias(alias3).LinkedToRollApp(rollApp3WithoutAlias.rollAppId)
				}

				{
					const alias4 = "rol4"
					const alias42 = "rol42"
					eip155 := dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp4WithoutAlias.rollAppId)

					s.Equal(eip155, string(store.Get(dymnstypes.AliasToRollAppEip155IdRvlKey(alias4))))
					s.Equal(eip155, string(store.Get(dymnstypes.AliasToRollAppEip155IdRvlKey(alias42))))

					s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp4WithoutAlias.rollAppId)...)))

					s.True(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(eip155)...)))

					s.requireRollApp(rollApp4WithoutAlias.rollAppId).HasAliasesWithOrder(alias4, alias42)
					s.requireAlias(alias4).LinkedToRollApp(rollApp4WithoutAlias.rollAppId)
					s.requireAlias(alias42).LinkedToRollApp(rollApp4WithoutAlias.rollAppId)
				}

				{
					s.requireRollApp(rollApp5WithoutAlias.rollAppId).HasNoAlias()

					s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(rollApp5WithoutAlias.rollAppId)...)))

					eip155 := dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollApp5WithoutAlias.rollAppId)
					s.False(store.Has(append(dymnstypes.KeyPrefixRollAppEip155IdToAliases, []byte(eip155)...)))
				}
			},
			wantFirstRun: true,
		},
		{
			name: "can migrate alias and accept nothing need migration",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(rollApp1MultiAliases)
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1MultiAliases.rollAppId).HasAliasesWithOrder("rol1", "rol11")
				s.requireAlias("rol1").LinkedToRollApp(rollApp1MultiAliases.rollAppId)
				s.requireAlias("rol11").LinkedToRollApp(rollApp1MultiAliases.rollAppId)
			},
			wantFirstRun: false,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()
			tt.setupFunc(s)

			for run := 1; run < 5; run++ {
				anyMigrated := s.dymNsKeeper.MigrateStoreForPlayground(s.ctx)

				tt.testFunc(s)

				if run == 1 {
					s.Equal(tt.wantFirstRun, anyMigrated)
				} else {
					s.False(anyMigrated)
				}

				if anyMigrated {
					var foundEvent bool
					for _, event := range s.ctx.EventManager().Events() {
						if event.Type == "dymns_migrated_for_playground" {
							foundEvent = true
							break
						}
					}
					s.True(foundEvent, "event should be emitted")
				}
			}
		})
	}
}

package keeper_test

import (
	"reflect"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestKeeper_MigrateChainIds() {
	addr1a := testAddr(1).bech32()
	addr2a := testAddr(2).bech32()
	cosmos3A := testAddr(3).bech32C("cosmos")

	tests := []struct {
		name                  string
		dymNames              []dymnstypes.DymName
		replacement           []dymnstypes.MigrateChainId
		chainsAliasParams     []dymnstypes.AliasesOfChainId
		additionalSetup       func(s *KeeperTestSuite)
		wantErr               bool
		wantErrContains       string
		wantDymNames          []dymnstypes.DymName
		wantChainsAliasParams []dymnstypes.AliasesOfChainId
	}{
		{
			name: "pass - can migrate",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "cosmoshub-3",
						Path:    "",
						Value:   cosmos3A,
					}},
				},
				{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
				},
			},
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "blumbus_111-1",
					NewChainId:      "blumbus_111-2",
				},
			},
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "cosmoshub-3",
					Aliases: []string{"cosmos"},
				},
			},
			additionalSetup: nil,
			wantErr:         false,
			wantDymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "cosmoshub-4",
						Path:    "",
						Value:   cosmos3A,
					}},
				},
				{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
				},
			},
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "cosmoshub-4",
					Aliases: []string{"cosmos"},
				},
			},
		},
		{
			name: "pass - can migrate params alias chain-id",
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "blumbus_111-1",
					NewChainId:      "blumbus_111-2",
				},
			},
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "cosmoshub-3",
					Aliases: []string{"cosmos"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"bb"},
				},
				{
					ChainId: s.chainId,
					Aliases: []string{"dym"},
				},
			},
			wantErr: false,
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "cosmoshub-4",
					Aliases: []string{"cosmos"},
				},
				{
					ChainId: "blumbus_111-2",
					Aliases: []string{"bb"},
				},
				{
					ChainId: s.chainId,
					Aliases: []string{"dym"},
				},
			},
		},
		{
			name: "pass - can Dym-Name",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "cosmoshub-3",
						Path:    "",
						Value:   cosmos3A,
					}},
				},
				{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-3",
							Path:    "",
							Value:   cosmos3A,
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "blumbus_111-1",
							Path:    "",
							Value:   addr2a,
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "froopyland_100-1",
							Path:    "",
							Value:   addr2a,
						},
					},
				},
			},
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "blumbus_111-1",
					NewChainId:      "blumbus_111-2",
				},
			},
			wantErr: false,
			wantDymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "cosmoshub-4",
						Path:    "",
						Value:   cosmos3A,
					}},
				},
				{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmos3A,
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "blumbus_111-2",
							Path:    "",
							Value:   addr2a,
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "froopyland_100-1",
							Path:    "",
							Value:   addr2a,
						},
					},
				},
			},
		},
		{
			name: "pass - ignore expired Dym-Name",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "cosmoshub-3",
						Path:    "",
						Value:   cosmos3A,
					}},
				},
				{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() - 1,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "cosmoshub-3",
						Path:    "",
						Value:   cosmos3A,
					}},
				},
			},
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
			},
			wantErr: false,
			wantDymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "cosmoshub-4",
						Path:    "",
						Value:   cosmos3A,
					}},
				},
				{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() - 1,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "cosmoshub-3", // keep
						Path:    "",
						Value:   cosmos3A,
					}},
				},
			},
		},
		{
			name: "fail - should stop if can not migrate params",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "blumbus_111-1",
						Path:    "",
						Value:   addr1a,
					}},
				},
			},
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "blumbus_111-1",
					NewChainId:      "dym", // collision with alias
				},
			},
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: s.chainId,
					Aliases: []string{"dym"}, // collision with new chain-id
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"bb"},
				},
			},
			wantErr:         true,
			wantErrContains: "chains params: alias: chain ID and alias must unique among all",
			wantDymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "blumbus_111-1", // not updated
						Path:    "",
						Value:   addr1a,
					}},
				},
			},
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				// not changed
				{
					ChainId: s.chainId,
					Aliases: []string{"dym"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"bb"},
				},
			},
		},
		{
			name: "fail - should stop if new params does not valid",
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "blumbus_111-1",
					NewChainId:      "dym", // collision with alias
				},
			},
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: s.chainId,
					Aliases: []string{"dym"}, // collision with new chain-id
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"bb"},
				},
			},
			wantErr:         true,
			wantErrContains: "chains params: alias: chain ID and alias must unique among all",
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				// not changed
				{
					ChainId: s.chainId,
					Aliases: []string{"dym"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"bb"},
				},
			},
		},
		{
			name:     "pass - should complete even tho nothing to update",
			dymNames: nil,
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
			},
			wantErr: false,
		},
		{
			name: "pass - skip migrate alias if new chain-id present, just remove",
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
			},
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: s.chainId,
					Aliases: []string{"dym"},
				},
				{
					ChainId: "cosmoshub-3",
					Aliases: []string{"cosmos3"},
				},
				{
					ChainId: "cosmoshub-4",
					Aliases: []string{"cosmos4"},
				},
			},
			wantErr: false,
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: s.chainId,
					Aliases: []string{"dym"},
				},
				{
					ChainId: "cosmoshub-4",
					Aliases: []string{"cosmos4"},
				},
			},
		},
		{
			name: "pass - skip migrate Dym-Name if new record does not pass validation",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						// migrate this will cause non-unique config
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-3",
							Path:    "",
							Value:   cosmos3A,
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmos3A,
						},
					},
				},
				{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-3",
							Path:    "",
							Value:   cosmos3A,
						},
					},
				},
				{
					Name:       "c",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmos3A,
						},
					},
				},
			},
			replacement: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
			},
			wantErr: false,
			wantDymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-3", // keep
							Path:    "",
							Value:   cosmos3A,
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmos3A,
						},
					},
				},
				{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4", // migrated
							Path:    "",
							Value:   cosmos3A,
						},
					},
				},
				{
					Name:       "c",
					Owner:      addr1a,
					Controller: addr1a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmos3A,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
				moduleParams.Chains.AliasesOfChainIds = tt.chainsAliasParams
				return moduleParams
			})

			for _, dymName := range tt.dymNames {
				s.setDymNameWithFunctionsAfter(dymName)
			}

			if tt.additionalSetup != nil {
				tt.additionalSetup(s)
			}

			err := s.dymNsKeeper.MigrateChainIds(s.ctx, tt.replacement)

			defer func() {
				laterModuleParams := s.moduleParams()
				if len(tt.wantChainsAliasParams) > 0 || len(laterModuleParams.Chains.AliasesOfChainIds) > 0 {
					if !reflect.DeepEqual(tt.wantChainsAliasParams, laterModuleParams.Chains.AliasesOfChainIds) {
						s.T().Errorf("alias: want %v, got %v", tt.wantChainsAliasParams, laterModuleParams.Chains.AliasesOfChainIds)
					}
				}
			}()

			defer func() {
				for _, wantDymName := range tt.wantDymNames {
					laterDymName := s.dymNsKeeper.GetDymName(s.ctx, wantDymName.Name)
					s.Require().NotNil(laterDymName)
					if !reflect.DeepEqual(wantDymName.Configs, laterDymName.Configs) {
						s.T().Errorf("dym name config: want %v, got %v", wantDymName.Configs, laterDymName.Configs)
					}
					if !reflect.DeepEqual(wantDymName, *laterDymName) {
						s.T().Errorf("dym name: want %v, got %v", wantDymName, *laterDymName)
					}
				}
			}()

			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_UpdateAliases() {
	tests := []struct {
		name                  string
		chainsAliasParams     []dymnstypes.AliasesOfChainId
		add                   []dymnstypes.UpdateAlias
		remove                []dymnstypes.UpdateAlias
		wantErr               bool
		wantErrContains       string
		wantChainsAliasParams []dymnstypes.AliasesOfChainId
	}{
		{
			name: "pass - can migrate",
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"dymension"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"blumbus"},
				},
				{
					ChainId: "froopyland_100-1",
					Aliases: []string{"fl"},
				},
			},
			add: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "bb",
				},
				{
					ChainId: "froopyland_100-1",
					Alias:   "frl",
				},
			},
			remove: []dymnstypes.UpdateAlias{
				{
					ChainId: "blumbus_111-1",
					Alias:   "blumbus",
				},
				{
					ChainId: "froopyland_100-1",
					Alias:   "fl",
				},
			},
			wantErr: false,
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"bb"},
				},
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"dym", "dymension"},
				},
				{
					ChainId: "froopyland_100-1",
					Aliases: []string{"frl"},
				},
			},
		},
		{
			name: "pass - records are sorted asc by chain-id",
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"blumbus"},
				},
				{
					ChainId: "froopyland_100-1",
					Aliases: []string{"fl"},
				},
			},
			add: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
				{
					ChainId: "froopyland_100-1",
					Alias:   "frl",
				},
			},
			wantErr: false,
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"blumbus"},
				},
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"dym"},
				},
				{
					ChainId: "froopyland_100-1",
					Aliases: []string{"fl", "frl"},
				},
			},
		},
		{
			name: "pass - aliases are sorted asc",
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"d5", "d3"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"b2"},
				},
			},
			add: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "d4",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "d1",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "d2",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "b4",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "b1",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "b3",
				},
			},
			remove: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "d5",
				},
			},
			wantErr: false,
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"b1", "b2", "b3", "b4"},
				},
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"d1", "d2", "d3", "d4"},
				},
			},
		},
		{
			name: "fail - adding existing alias of same chain-id",
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"d1", "d2"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"b1"},
				},
			},
			add: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "d1",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "d3",
				},
			},
			wantErr:         true,
			wantErrContains: "alias: d1 for dymension_1100-1: already exists",
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"d1", "d2"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"b1"},
				},
			},
		},
		{
			name: "fail - removing non-existing chain-id",
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"d1"},
				},
			},
			add: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "d2",
				},
			},
			remove: []dymnstypes.UpdateAlias{
				{
					ChainId: "blumbus_111-1",
					Alias:   "b1",
				},
			},
			wantErr:         true,
			wantErrContains: "not found to remove",
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"d1"},
				},
			},
		},
		{
			name: "fail - removing non-existing alias of chain-id",
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"d1", "d2"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"b1"},
				},
			},
			remove: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "d3",
				},
			},
			wantErr:         true,
			wantErrContains: "not found to remove",
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"d1", "d2"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"b1"},
				},
			},
		},
		{
			name: "fail - do not update if params validation failed",
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"dym"},
				},
			},
			add: []dymnstypes.UpdateAlias{
				{
					ChainId: "blumbus_111-1",
					Alias:   "dym", // collision with existing alias
				},
			},
			wantErr:         true,
			wantErrContains: "chain ID and alias must unique among all",
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"dym"},
				},
			},
		},
		{
			name: "pass - remove records that no more alias",
			chainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"dym", "dymension"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"blumbus"},
				},
			},
			remove: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dymension",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "blumbus",
				},
			},
			wantErr: false,
			wantChainsAliasParams: []dymnstypes.AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"dym"},
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
				moduleParams.Chains.AliasesOfChainIds = tt.chainsAliasParams
				return moduleParams
			})

			err := s.dymNsKeeper.UpdateAliases(s.ctx, tt.add, tt.remove)

			laterModuleParams := s.moduleParams()
			defer func() {
				if s.T().Failed() {
					return
				}

				if len(tt.wantChainsAliasParams) == 0 {
					s.Require().Empty(laterModuleParams.Chains.AliasesOfChainIds)
				} else {
					s.Require().Equal(tt.wantChainsAliasParams, laterModuleParams.Chains.AliasesOfChainIds)
				}
			}()

			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
		})
	}
}

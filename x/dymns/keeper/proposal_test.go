package keeper_test

import (
	"reflect"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/stretchr/testify/require"
)

func TestKeeper_MigrateChainIds(t *testing.T) {
	now := time.Now().UTC()
	const chainId = "dymension_1100-1"

	addr1a := testAddr(1).bech32()
	addr2a := testAddr(2).bech32()
	cosmos3A := testAddr(3).bech32C("cosmos")

	tests := []struct {
		name                  string
		dymNames              []dymnstypes.DymName
		replacement           []dymnstypes.MigrateChainId
		chainsAliasParams     []dymnstypes.AliasesOfChainId
		additionalSetup       func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper)
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ChainId: chainId,
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
					ChainId: chainId,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() - 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() - 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ChainId: chainId,
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
					ExpireAt:   now.Unix() + 1,
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
					ChainId: chainId,
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
					ChainId: chainId,
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
					ChainId: chainId,
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
					ChainId: chainId,
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
					ChainId: chainId,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
					ExpireAt:   now.Unix() + 1,
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
		t.Run(tt.name, func(t *testing.T) {
			dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

			ctx = ctx.WithBlockTime(now).WithChainID(chainId)

			moduleParams := dk.GetParams(ctx)
			moduleParams.Chains.AliasesOfChainIds = tt.chainsAliasParams
			require.NoError(t, dk.SetParams(ctx, moduleParams))

			for _, dymName := range tt.dymNames {
				setDymNameWithFunctionsAfter(ctx, dymName, t, dk)
			}

			if tt.additionalSetup != nil {
				tt.additionalSetup(ctx, dk, rk)
			}

			err := dk.MigrateChainIds(ctx, tt.replacement)

			defer func() {
				laterModuleParams := dk.GetParams(ctx)
				if len(tt.wantChainsAliasParams) > 0 || len(laterModuleParams.Chains.AliasesOfChainIds) > 0 {
					if !reflect.DeepEqual(tt.wantChainsAliasParams, laterModuleParams.Chains.AliasesOfChainIds) {
						t.Errorf("alias: want %v, got %v", tt.wantChainsAliasParams, laterModuleParams.Chains.AliasesOfChainIds)
					}
				}
			}()

			defer func() {
				for _, wantDymName := range tt.wantDymNames {
					laterDymName := dk.GetDymName(ctx, wantDymName.Name)
					require.NotNil(t, laterDymName)
					if !reflect.DeepEqual(wantDymName.Configs, laterDymName.Configs) {
						t.Errorf("dym name config: want %v, got %v", wantDymName.Configs, laterDymName.Configs)
					}
					if !reflect.DeepEqual(wantDymName, *laterDymName) {
						t.Errorf("dym name: want %v, got %v", wantDymName, *laterDymName)
					}
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestKeeper_UpdateAliases(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			moduleParams := dk.GetParams(ctx)
			moduleParams.Chains.AliasesOfChainIds = tt.chainsAliasParams
			require.NoError(t, dk.SetParams(ctx, moduleParams))

			err := dk.UpdateAliases(ctx, tt.add, tt.remove)

			laterModuleParams := dk.GetParams(ctx)
			defer func() {
				if t.Failed() {
					return
				}

				if len(tt.wantChainsAliasParams) == 0 {
					require.Empty(t, laterModuleParams.Chains.AliasesOfChainIds)
				} else {
					require.Equal(t, tt.wantChainsAliasParams, laterModuleParams.Chains.AliasesOfChainIds)
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

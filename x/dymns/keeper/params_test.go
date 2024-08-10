package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func TestGetSetParams(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	params := dymnstypes.DefaultParams()

	err := dk.SetParams(ctx, params)
	require.NoError(t, err)

	require.Equal(t, params, dk.GetParams(ctx))

	t.Run("can not set invalid params", func(t *testing.T) {
		params := dymnstypes.DefaultParams()
		params.Misc.BeginEpochHookIdentifier = ""
		require.Error(t, dk.SetParams(ctx, params))
	})

	t.Run("can not set invalid params", func(t *testing.T) {
		params := dymnstypes.DefaultParams()
		params.Price.PriceDenom = ""
		require.Error(t, dk.SetParams(ctx, params))
	})

	t.Run("can not set invalid params", func(t *testing.T) {
		params := dymnstypes.DefaultParams()
		params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
			{
				ChainId: "@",
				Aliases: nil,
			},
		}
		require.Error(t, dk.SetParams(ctx, params))
	})

	t.Run("can not set invalid params", func(t *testing.T) {
		params := dymnstypes.DefaultParams()
		params.Misc.GracePeriodDuration = -999 * time.Hour
		require.Error(t, dk.SetParams(ctx, params))
	})
}

func (s *KeeperTestSuite) TestKeeper_CanUseAliasForNewRegistration() {
	tests := []struct {
		name            string
		alias           string
		preSetup        func(s *KeeperTestSuite)
		wantErr         bool
		wantErrContains string
		want            bool
	}{
		{
			name:    "pass - can check",
			alias:   "a",
			wantErr: false,
			want:    true,
		},
		{
			name:            "fail - reject bad alias",
			alias:           "@",
			wantErr:         true,
			wantErrContains: "alias candidate: invalid argument",
		},
		{
			name:  "pass - returns as free if neither in Params or Roll-App",
			alias: "free",
			preSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return params
				})

				s.persistRollApp(
					*newRollApp("rollapp_1-1").WithAlias("ra"),
				)

				s.requireRollApp("rollapp_1-1").HasAlias("ra")
			},
			wantErr: false,
			want:    true,
		},
		{
			name:    "pass - returns as free if no params, no Roll-App",
			alias:   "free",
			wantErr: false,
			want:    true,
		},
		{
			name:  "pass - returns as NOT free if reserved in Params",
			alias: "dymension",
			preSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym", "dymension"},
						},
					}
					return params
				})

				s.persistRollApp(
					*newRollApp("rollapp_1-1").WithAlias("ra"),
				)

				s.requireRollApp("rollapp_1-1").HasAlias("ra")
			},
			wantErr: false,
			want:    false,
		},
		{
			name:  "pass - returns as NOT free if reserved in Params as chain-id, without alias",
			alias: "zeta",
			preSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "zeta",
							Aliases: nil,
						},
					}
					return params
				})
			},
			wantErr: false,
			want:    false,
		},
		{
			name:  "pass - returns as NOT free if reserved in RollApp",
			alias: "ra",
			preSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym", "dymension"},
						},
					}
					return params
				})

				s.persistRollApp(
					*newRollApp("rollapp_1-1").WithAlias("ra"),
				)

				s.requireRollApp("rollapp_1-1").HasAlias("ra")
			},
			wantErr: false,
			want:    false,
		},
		{
			name:  "pass - returns as NOT free if reserved in RollApp, which owned multiple aliases",
			alias: "two",
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(
					*newRollApp("rollapp_1-1").WithAlias("one"),
				)

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_1-1", "two")
				s.Require().NoError(err)

				err = s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_1-1", "three")
				s.Require().NoError(err)

				s.requireRollApp("rollapp_1-1").HasAlias("one", "two", "three")
			},
			wantErr: false,
			want:    false,
		},
		{
			name:  "pass - returns as NOT free if reserved in both Params and RollApp",
			alias: "dym",
			preSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym", "dymension"},
						},
					}
					return params
				})

				s.persistRollApp(
					*newRollApp("dymension_1-1").WithAlias("dym"),
				)

				s.requireRollApp("dymension_1-1").HasAlias("dym")
			},
			wantErr: false,
			want:    false,
		},
		{
			name:  "pass - returns as NOT free if it is a Chain-ID in params mapping",
			alias: "bridge",
			preSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "bridge",
							Aliases: []string{"b"},
						},
					}
					return params
				})
			},
			wantErr: false,
			want:    false,
		},
		// TODO DymNS: FIXME * this test will panic because RollApp keeper now validate the RollApp-ID,
		//  must find a way to make a RollApp with chain-id compatible with alias format
		/*
			{
				name:  "pass - returns as NOT free if it is a RollApp-ID",
				alias: "bridge",
				preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
					rk.SetRollapp(ctx, rollapptypes.Rollapp{
						RollappId: "bridge",
						Owner:     testAddr(1).bech32(),
					})
					err := dk.SetAliasForRollAppId(ctx, "bridge", "b")
					require.NoError(t, err)

					require.True(t, dk.IsRollAppId(ctx, "bridge"))
				},
				wantErr: false,
				want:    false,
			},
		*/
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			if tt.preSetup != nil {
				tt.preSetup(s)
			}

			can, err := s.dymNsKeeper.CanUseAliasForNewRegistration(s.ctx, tt.alias)
			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)

				s.Require().False(can)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tt.want, can)
		})
	}
}

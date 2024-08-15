package keeper_test

import (
	"time"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestGetSetParams() {
	params := dymnstypes.DefaultParams()

	err := s.dymNsKeeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	s.Require().Equal(params, s.dymNsKeeper.GetParams(s.ctx))

	s.Run("can not set invalid params", func() {
		params := dymnstypes.DefaultParams()
		params.Misc.EndEpochHookIdentifier = ""
		s.Require().Error(s.dymNsKeeper.SetParams(s.ctx, params))
	})

	s.Run("can not set invalid params", func() {
		params := dymnstypes.DefaultParams()
		params.Price.PriceDenom = ""
		s.Require().Error(s.dymNsKeeper.SetParams(s.ctx, params))
	})

	s.Run("can not set invalid params", func() {
		params := dymnstypes.DefaultParams()
		params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
			{
				ChainId: "@",
				Aliases: nil,
			},
		}
		s.Require().Error(s.dymNsKeeper.SetParams(s.ctx, params))
	})

	s.Run("can not set invalid params", func() {
		params := dymnstypes.DefaultParams()
		params.Misc.GracePeriodDuration = -999 * time.Hour
		s.Require().Error(s.dymNsKeeper.SetParams(s.ctx, params))
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
		{
			name:  "pass - returns as NOT free if it is a RollApp-ID",
			alias: "bridge",
			preSetup: func(s *KeeperTestSuite) {
				s.pureSetRollApp(rollapptypes.Rollapp{
					RollappId: "bridge",
					Owner:     testAddr(1).bech32(),
				})
				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "bridge", "b")
				s.Require().NoError(err)
			},
			wantErr: false,
			want:    false,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

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

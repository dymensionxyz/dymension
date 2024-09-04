package keeper_test

import (
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (s *KeeperTestSuite) TestKeeper_IsRollAppId() {
	s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_1-1",
		Owner:     testAddr(1).bech32(),
	})

	s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
		RollappId: "rolling_2-2",
		Owner:     testAddr(2).bech32(),
	})

	tests := []struct {
		rollAppId     string
		wantIsRollApp bool
	}{
		{
			rollAppId:     "rollapp_1-1",
			wantIsRollApp: true,
		},
		{
			rollAppId:     "rolling_2-2",
			wantIsRollApp: true,
		},
		{
			rollAppId:     "rollapp_1-11",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rollapp_11-1",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rollapp_11-11",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rollapp_1-2",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rollapp_2-1",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rolling_1-1",
			wantIsRollApp: false,
		},
	}
	for _, tt := range tests {
		s.Run(tt.rollAppId, func() {
			gotIsRollApp := s.dymNsKeeper.IsRollAppId(s.ctx, tt.rollAppId)
			s.Require().Equal(tt.wantIsRollApp, gotIsRollApp)
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_IsRollAppCreator() {
	acc1 := testAddr(1)
	acc2 := testAddr(2)

	tests := []struct {
		name      string
		rollApp   *rollapptypes.Rollapp
		rollAppId string
		account   string
		want      bool
	}{
		{
			name: "pass - is creator",
			rollApp: &rollapptypes.Rollapp{
				RollappId: "rollapp_1-1",
				Owner:     acc1.bech32(),
			},
			rollAppId: "rollapp_1-1",
			account:   acc1.bech32(),
			want:      true,
		},
		{
			name: "fail - rollapp does not exists",
			rollApp: &rollapptypes.Rollapp{
				RollappId: "rollapp_1-1",
				Owner:     acc1.bech32(),
			},
			rollAppId: "nah_2-2",
			account:   acc1.bech32(),
			want:      false,
		},
		{
			name: "fail - is NOT creator",
			rollApp: &rollapptypes.Rollapp{
				RollappId: "rollapp_1-1",
				Owner:     acc1.bech32(),
			},
			rollAppId: "rollapp_1-1",
			account:   acc2.bech32(),
			want:      false,
		},
		{
			name: "fail - creator but in different bech32 format is not accepted",
			rollApp: &rollapptypes.Rollapp{
				RollappId: "rollapp_1-1",
				Owner:     acc1.bech32(),
			},
			rollAppId: "rollapp_1-1",
			account:   acc1.bech32C("nim"),
			want:      false,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.rollApp != nil {
				s.rollAppKeeper.SetRollapp(s.ctx, *tt.rollApp)
			}

			got := s.dymNsKeeper.IsRollAppCreator(s.ctx, tt.rollAppId, tt.account)
			s.Require().Equal(tt.want, got)
		})
	}

	s.Run("pass - can detect among multiple RollApps of same owned", func() {
		s.RefreshContext()

		rollAppABy1 := rollapptypes.Rollapp{
			RollappId: "rollapp_1-1",
			Owner:     acc1.bech32(),
		}
		rollAppBBy1 := rollapptypes.Rollapp{
			RollappId: "rollapp_2-2",
			Owner:     acc1.bech32(),
		}
		rollAppCBy2 := rollapptypes.Rollapp{
			RollappId: "rollapp_3-3",
			Owner:     acc2.bech32(),
		}
		rollAppDBy2 := rollapptypes.Rollapp{
			RollappId: "rollapp_4-4",
			Owner:     acc2.bech32(),
		}

		s.rollAppKeeper.SetRollapp(s.ctx, rollAppABy1)
		s.rollAppKeeper.SetRollapp(s.ctx, rollAppBBy1)
		s.rollAppKeeper.SetRollapp(s.ctx, rollAppCBy2)
		s.rollAppKeeper.SetRollapp(s.ctx, rollAppDBy2)

		s.Require().True(s.dymNsKeeper.IsRollAppCreator(s.ctx, rollAppABy1.RollappId, acc1.bech32()))
		s.Require().True(s.dymNsKeeper.IsRollAppCreator(s.ctx, rollAppBBy1.RollappId, acc1.bech32()))
		s.Require().True(s.dymNsKeeper.IsRollAppCreator(s.ctx, rollAppCBy2.RollappId, acc2.bech32()))
		s.Require().True(s.dymNsKeeper.IsRollAppCreator(s.ctx, rollAppDBy2.RollappId, acc2.bech32()))

		s.Require().False(s.dymNsKeeper.IsRollAppCreator(s.ctx, rollAppABy1.RollappId, acc2.bech32()))
		s.Require().False(s.dymNsKeeper.IsRollAppCreator(s.ctx, rollAppBBy1.RollappId, acc2.bech32()))
		s.Require().False(s.dymNsKeeper.IsRollAppCreator(s.ctx, rollAppCBy2.RollappId, acc1.bech32()))
		s.Require().False(s.dymNsKeeper.IsRollAppCreator(s.ctx, rollAppDBy2.RollappId, acc1.bech32()))
	})
}

func (s *KeeperTestSuite) TestKeeper_GetRollAppBech32Prefix() {
	rollApp1 := rollapptypes.Rollapp{
		RollappId: "rollapp_1-1",
		Owner:     testAddr(0).bech32(),
		GenesisInfo: rollapptypes.GenesisInfo{
			Bech32Prefix: "one",
		},
	}
	rollApp2 := rollapptypes.Rollapp{
		RollappId: "rolling_2-2",
		Owner:     testAddr(0).bech32(),
		GenesisInfo: rollapptypes.GenesisInfo{
			Bech32Prefix: "two",
		},
	}
	rollApp3NonExists := rollapptypes.Rollapp{
		RollappId: "nah_3-3",
		Owner:     testAddr(0).bech32(),
		GenesisInfo: rollapptypes.GenesisInfo{
			Bech32Prefix: "three",
		},
	}

	s.rollAppKeeper.SetRollapp(s.ctx, rollApp1)
	s.rollAppKeeper.SetRollapp(s.ctx, rollApp2)

	bech32, found := s.dymNsKeeper.GetRollAppBech32Prefix(s.ctx, rollApp1.RollappId)
	s.True(found)
	s.Equal("one", bech32)

	bech32, found = s.dymNsKeeper.GetRollAppBech32Prefix(s.ctx, rollApp2.RollappId)
	s.True(found)
	s.Equal("two", bech32)

	bech32, found = s.dymNsKeeper.GetRollAppBech32Prefix(s.ctx, rollApp3NonExists.RollappId)
	s.False(found)
	s.Empty(bech32)
}

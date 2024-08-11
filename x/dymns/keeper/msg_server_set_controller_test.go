package keeper_test

import (
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_msgServer_SetController() {
	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).SetController(s.ctx, &dymnstypes.MsgSetController{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	ownerA := testAddr(1).bech32()
	controllerA := testAddr(2).bech32()
	notOwnerA := testAddr(3).bech32()

	tests := []struct {
		name            string
		dymName         *dymnstypes.DymName
		recordName      string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "fail - reject if Dym-Name not found",
			recordName:      "a",
			wantErr:         true,
			wantErrContains: "Dym-Name: a: not found",
		},
		{
			name: "fail - reject if not owned",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      notOwnerA,
				Controller: notOwnerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			recordName:      "a",
			wantErr:         true,
			wantErrContains: "not the owner of the Dym-Name",
		},
		{
			name: "fail - reject if not new controller is the same as previous controller",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			recordName:      "a",
			wantErr:         true,
			wantErrContains: "controller already set",
		},
		{
			name: "fail - reject if Dym-Name is already expired",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() - 1,
			},
			recordName:      "a",
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
		},
		{
			name: "pass - accept if new controller is different from previous controller",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			recordName: "a",
		},
		{
			name: "pass - changing controller will not change configs",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: ownerA,
				}},
			},
			recordName: "a",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.dymName != nil {
				err := s.dymNsKeeper.SetDymName(s.ctx, *tt.dymName)
				s.Require().NoError(err)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).SetController(s.ctx, &dymnstypes.MsgSetController{
				Name:       tt.recordName,
				Controller: controllerA,
				Owner:      ownerA,
			})
			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)

				s.Require().Nil(resp)

				laterDymName := s.dymNsKeeper.GetDymName(s.ctx, tt.recordName)

				if tt.dymName != nil {
					s.Require().Equal(*tt.dymName, *laterDymName)
				} else {
					s.Require().Nil(laterDymName)
				}

				return
			}

			s.Require().NoError(err)

			s.Require().NotNil(resp)

			s.Require().NotNil(tt.dymName, "mis-configured test case")

			laterDymName := s.dymNsKeeper.GetDymName(s.ctx, tt.recordName)
			s.Require().NotNil(laterDymName)

			s.Require().Equal(controllerA, laterDymName.Controller)
			s.Require().Equal(ownerA, laterDymName.Owner)

			s.Require().Equal(tt.dymName.ExpireAt, laterDymName.ExpireAt)
			s.Require().Equal(tt.dymName.Configs, laterDymName.Configs)
		})
	}
}

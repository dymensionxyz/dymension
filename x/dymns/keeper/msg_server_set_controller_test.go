package keeper_test

import (
	"testing"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_SetController(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, ctx := setupTest()

		_, err := dymnskeeper.NewMsgServerImpl(dk).SetController(ctx, &dymnstypes.MsgSetController{})
		require.ErrorContains(t, err, gerrc.ErrInvalidArgument.Error())
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() - 1,
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
				ExpireAt:   now.Unix() + 1,
			},
			recordName: "a",
		},
		{
			name: "pass - changing controller will not change configs",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: ownerA,
				}},
			},
			recordName: "a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			if tt.dymName != nil {
				err := dk.SetDymName(ctx, *tt.dymName)
				require.NoError(t, err)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).SetController(ctx, &dymnstypes.MsgSetController{
				Name:       tt.recordName,
				Controller: controllerA,
				Owner:      ownerA,
			})
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)

				require.Nil(t, resp)

				laterDymName := dk.GetDymName(ctx, tt.recordName)

				if tt.dymName != nil {
					require.Equal(t, *tt.dymName, *laterDymName)
				} else {
					require.Nil(t, laterDymName)
				}

				return
			}

			require.NoError(t, err)

			require.NotNil(t, resp)

			require.NotNil(t, tt.dymName, "mis-configured test case")

			laterDymName := dk.GetDymName(ctx, tt.recordName)
			require.NotNil(t, laterDymName)

			require.Equal(t, controllerA, laterDymName.Controller)
			require.Equal(t, ownerA, laterDymName.Owner)

			require.Equal(t, tt.dymName.ExpireAt, laterDymName.ExpireAt)
			require.Equal(t, tt.dymName.Configs, laterDymName.Configs)
		})
	}
}

package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func Test_msgServer_SetController(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).SetController(ctx, &dymnstypes.MsgSetController{})
			return err
		}, dymnstypes.ErrValidationFailed.Error())
	})

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const controller = "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"
	const recordName = "bonded-pool"

	tests := []struct {
		name            string
		dymName         *dymnstypes.DymName
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "reject if Dym-Name not found",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrDymNameNotFound.Error(),
		},
		{
			name: "reject if not owned",
			dymName: &dymnstypes.DymName{
				Owner:      "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
				Controller: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
				ExpireAt:   now.Unix() + 1,
			},
			wantErr:         true,
			wantErrContains: sdkerrors.ErrUnauthorized.Error(),
		},
		{
			name: "reject if not new controller is the same as previous controller",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			wantErr:         true,
			wantErrContains: "controller already set",
		},
		{
			name: "reject if Dym-Name is already expired",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() - 1,
			},
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
		},
		{
			name: "accept if new controller is different from previous controller",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
			},
		},
		{
			name: "changing controller will not change configs",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: owner,
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			if tt.dymName != nil {
				tt.dymName.Name = recordName
				err := dk.SetDymName(ctx, *tt.dymName)
				require.NoError(t, err)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).SetController(ctx, &dymnstypes.MsgSetController{
				Name:       "bonded-pool",
				Controller: controller,
				Owner:      owner,
			})
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)

				require.Nil(t, resp)

				laterDymName := dk.GetDymName(ctx, recordName)

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

			laterDymName := dk.GetDymName(ctx, recordName)
			require.NotNil(t, laterDymName)

			require.Equal(t, controller, laterDymName.Controller)
			require.Equal(t, owner, laterDymName.Owner)

			require.Equal(t, tt.dymName.ExpireAt, laterDymName.ExpireAt)
			require.Equal(t, tt.dymName.Configs, laterDymName.Configs)
		})
	}
}

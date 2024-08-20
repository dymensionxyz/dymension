package post_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/post"
	"github.com/stretchr/testify/require"
)

var (
	testClientState = ibctm.NewClientState("chain-id",
		ibctm.DefaultTrustLevel, time.Hour*24*7*2, time.Hour*24*7*2, time.Second*10,
		ibcclienttypes.MustParseHeight("1-1"), commitmenttypes.GetSDKSpecs(), []string{},
	)
)

type testInput struct {
	msg     *ibcclienttypes.MsgCreateClient
	success bool
}

func TestHandleMsgCreateClient(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func(ctx sdk.Context, k keeper.Keeper) testInput
		assert  func(ctx sdk.Context, k keeper.Keeper)
	}{
		{
			name: "Could not unpack light client state to tendermint state",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				return testInput{
					msg:     &ibcclienttypes.MsgCreateClient{},
					success: true,
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {},
		},
		{
			name: "Canonical client registration not in progress",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				testClientState.ChainId = "not-a-rollapp"
				cs, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState: cs,
					},
					success: true,
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {},
		},
		{
			name: "Canonical client registration in progress - tx was failure",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				testClientState.ChainId = "rollapp-client-registration-in-progress"
				cs, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				expectedClientID := "new-client-1"
				k.BeginCanonicalLightClientRegistration(
					ctx,
					"rollapp-client-registration-in-progress",
					expectedClientID,
				)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState: cs,
					},
					success: false,
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				_, found := k.GetCanonicalLightClientRegistration(ctx, "rollapp-client-registration-in-progress")
				require.False(t, found)
			},
		},
		{
			name: "Canonical client registration in progress - client was found with name",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				testClientState.ChainId = "rollapp-client-registration-in-progress"
				cs, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				expectedClientID := "canon-client-id"
				k.BeginCanonicalLightClientRegistration(
					ctx,
					"rollapp-client-registration-in-progress",
					expectedClientID,
				)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState: cs,
					},
					success: true,
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				clientID, found := k.GetCanonicalClient(ctx, "rollapp-client-registration-in-progress")
				require.True(t, found)
				require.Equal(t, "canon-client-id", clientID)
				_, found = k.GetCanonicalLightClientRegistration(ctx, "rollapp-client-registration-in-progress")
				require.False(t, found)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := keepertest.LightClientKeeper(t)
			ibcclientKeeper := NewMockIBCClientKeeper()
			ibcMsgDecorator := post.NewIBCMessagesDecorator(*keeper, ibcclientKeeper)

			input := tc.prepare(ctx, *keeper)
			ibcMsgDecorator.HandleMsgCreateClient(ctx, input.msg, input.success)
			tc.assert(ctx, *keeper)
		})
	}

}

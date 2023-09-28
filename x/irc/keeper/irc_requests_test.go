package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/irc/keeper"
	"github.com/dymensionxyz/dymension/x/irc/types"
	"github.com/stretchr/testify/require"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNIRCRequest(t *testing.T, keeper *keeper.Keeper, ctx sdk.Context, n int) []types.IRCRequest {
	items := make([]types.IRCRequest, n)
	for i := range items {
		req, err := types.NewIRCRequest(uint64(i), &channeltypes.MsgRecvPacket{
			Packet:          channeltypes.Packet{},
			ProofCommitment: []byte{1, 2, 3, 4, 5, 6, byte(i)},
			ProofHeight: clienttypes.Height{
				RevisionNumber: uint64(i),
				RevisionHeight: uint64(i + 1),
			},
			Signer: "",
		})
		require.NoError(t, err)
		items[i] = *req
		keeper.SetIRCRequest(ctx, items[i])
	}
	return items
}

func TestIRCRequestGet(t *testing.T) {
	keeper, _, ctx := keepertest.IRCKeeper(t)

	items := createNIRCRequest(t, keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetIRCRequest(ctx,
			item.ReqId,
		)
		require.True(t, found)
		require.Equal(t, &item, &rst)
		msg := rst.GetMsg()
		require.NotNil(t, msg)
	}
}
func TestIRCRequestRemove(t *testing.T) {
	keeper, _, ctx := keepertest.IRCKeeper(t)
	items := createNIRCRequest(t, keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveIRCRequest(ctx,
			item.ReqId,
		)
		_, found := keeper.GetIRCRequest(ctx,
			item.ReqId,
		)
		require.False(t, found)
	}
}

func TestIRCRequestGetAll(t *testing.T) {
	keeper, _, ctx := keepertest.IRCKeeper(t)

	items := createNIRCRequest(t, keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllIRCRequest(ctx)),
	)
}

package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNStateIndex(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.StateIndex {
	items := make([]types.StateIndex, n)
	for i := range items {
		items[i].RollappId = strconv.Itoa(i)

		keeper.SetStateIndex(ctx, items[i])
	}
	return items
}

func TestStateIndexGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNStateIndex(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetStateIndex(ctx,
			item.RollappId,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestStateIndexRemove(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNStateIndex(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveStateIndex(ctx,
			item.RollappId,
		)
		_, found := keeper.GetStateIndex(ctx,
			item.RollappId,
		)
		require.False(t, found)
	}
}

func TestStateIndexGetAll(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNStateIndex(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllStateIndex(ctx)),
	)
}

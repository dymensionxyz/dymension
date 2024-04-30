package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNLatestStateInfoIndex(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.StateInfoIndex {
	items := make([]types.StateInfoIndex, n)
	for i := range items {
		items[i].RollappId = strconv.Itoa(i)

		keeper.SetLatestStateInfoIndex(ctx, items[i])
	}
	return items
}

func TestLatestStateInfoIndexGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNLatestStateInfoIndex(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetLatestStateInfoIndex(ctx,
			item.RollappId,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestLatestStateInfoIndexRemove(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNLatestStateInfoIndex(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLatestStateInfoIndex(ctx,
			item.RollappId,
		)
		_, found := keeper.GetLatestStateInfoIndex(ctx,
			item.RollappId,
		)
		require.False(t, found)
	}
}

func TestLatestStateInfoIndexGetAll(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNLatestStateInfoIndex(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllLatestStateInfoIndex(ctx)),
	)
}

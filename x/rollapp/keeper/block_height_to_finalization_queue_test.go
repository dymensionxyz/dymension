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

func createNBlockHeightToFinalizationQueue(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.BlockHeightToFinalizationQueue {
	items := make([]types.BlockHeightToFinalizationQueue, n)
	for i := range items {
		items[i].CreationHeight = uint64(i)

		keeper.SetBlockHeightToFinalizationQueue(ctx, items[i])
	}
	return items
}

func TestBlockHeightToFinalizationQueueGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetBlockHeightToFinalizationQueue(ctx,
			item.CreationHeight,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestBlockHeightToFinalizationQueueRemove(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveBlockHeightToFinalizationQueue(ctx,
			item.CreationHeight,
		)
		_, found := keeper.GetBlockHeightToFinalizationQueue(ctx,
			item.CreationHeight,
		)
		require.False(t, found)
	}
}

func TestBlockHeightToFinalizationQueueGetAll(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllBlockHeightToFinalizationQueue(ctx)),
	)
}

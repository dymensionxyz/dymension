package keeper_test

import (
	"strconv"
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestRollappGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items, _ := createNRollapp(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetRollapp(ctx,
			item.RollappId,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestRollappRemove(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items, _ := createNRollapp(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveRollapp(ctx,
			item.RollappId,
		)
		_, found := keeper.GetRollapp(ctx,
			item.RollappId,
		)
		require.False(t, found)
	}
}

func TestRollappGetAll(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items, _ := createNRollapp(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllRollapps(ctx)),
	)
}

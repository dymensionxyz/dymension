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

func createNRollapp(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Rollapp {
	items := make([]types.Rollapp, n)
	for i := range items {
		items[i].RollappId = strconv.Itoa(i)

		keeper.SetRollapp(ctx, items[i])
	}
	return items
}

func TestRollappGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNRollapp(keeper, ctx, 10)
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
	items := createNRollapp(keeper, ctx, 10)
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
	items := createNRollapp(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllRollapp(ctx)),
	)
}

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

func createNRollappStateInfo(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.RollappStateInfo {
	items := make([]types.RollappStateInfo, n)
	for i := range items {
		items[i].RollappId = strconv.Itoa(i)
		items[i].StateIndex = uint64(i)

		keeper.SetRollappStateInfo(ctx, items[i])
	}
	return items
}

func TestRollappStateInfoGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNRollappStateInfo(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetRollappStateInfo(ctx,
			item.RollappId,
			item.StateIndex,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestRollappStateInfoRemove(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNRollappStateInfo(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveRollappStateInfo(ctx,
			item.RollappId,
			item.StateIndex,
		)
		_, found := keeper.GetRollappStateInfo(ctx,
			item.RollappId,
			item.StateIndex,
		)
		require.False(t, found)
	}
}

func TestRollappStateInfoGetAll(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNRollappStateInfo(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllRollappStateInfo(ctx)),
	)
}

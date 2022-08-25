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

func createNStateInfo(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.StateInfo {
	items := make([]types.StateInfo, n)
	for i := range items {
		items[i].RollappId = strconv.Itoa(i)
		items[i].StateIndex = uint64(i)

		keeper.SetStateInfo(ctx, items[i])
	}
	return items
}

func TestStateInfoGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNStateInfo(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetStateInfo(ctx,
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
func TestStateInfoRemove(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNStateInfo(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveStateInfo(ctx,
			item.RollappId,
			item.StateIndex,
		)
		_, found := keeper.GetStateInfo(ctx,
			item.RollappId,
			item.StateIndex,
		)
		require.False(t, found)
	}
}

func TestStateInfoGetAll(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNStateInfo(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllStateInfo(ctx)),
	)
}

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

func createNStateInfo(keeper *keeper.Keeper, ctx sdk.Context, n int) ([]types.StateInfo, []types.StateInfoSummary) {
	items := make([]types.StateInfo, n)
	for i := range items {
		items[i].StateInfoIndex.RollappId = strconv.Itoa(i)
		items[i].StateInfoIndex.Index = 1 + uint64(i)

		keeper.SetStateInfo(ctx, items[i])
	}

	var stateInfoSummaries []types.StateInfoSummary
	for _, item := range items {
		stateInfoSummary := types.StateInfoSummary{
			StateInfoIndex: item.StateInfoIndex,
			Status:         item.Status,
			CreationHeight: item.CreationHeight,
		}
		stateInfoSummaries = append(stateInfoSummaries, stateInfoSummary)
	}

	return items, stateInfoSummaries
}

func TestStateInfoGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items, _ := createNStateInfo(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetStateInfo(ctx,
			item.StateInfoIndex.RollappId,
			item.StateInfoIndex.Index,
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
	items, _ := createNStateInfo(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveStateInfo(ctx,
			item.StateInfoIndex.RollappId,
			item.StateInfoIndex.Index,
		)
		_, found := keeper.GetStateInfo(ctx,
			item.StateInfoIndex.RollappId,
			item.StateInfoIndex.Index,
		)
		require.False(t, found)
	}
}

func TestStateInfoGetAll(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items, _ := createNStateInfo(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllStateInfo(ctx)),
	)
}

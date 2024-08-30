package keeper_test

import (
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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
	k, ctx := keepertest.RollappKeeper(t)
	items, _ := createNStateInfo(k, ctx, 10)
	for _, item := range items {
		item := item
		rst, found := k.GetStateInfo(ctx,
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
	k, ctx := keepertest.RollappKeeper(t)
	items, _ := createNStateInfo(k, ctx, 10)
	for _, item := range items {
		k.RemoveStateInfo(ctx,
			item.StateInfoIndex.RollappId,
			item.StateInfoIndex.Index,
		)
		_, found := k.GetStateInfo(ctx,
			item.StateInfoIndex.RollappId,
			item.StateInfoIndex.Index,
		)
		require.False(t, found)
	}
}

func TestStateInfoGetAll(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	items, _ := createNStateInfo(k, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(k.GetAllStateInfo(ctx)),
	)
}

func TestKeeper_DeleteStateInfoUntilTimestamp(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)

	ts1 := time.Date(2020, time.May, 1, 10, 22, 0, 0, time.UTC)
	ts2 := ts1.Add(9 * time.Second)
	ts3 := ts2.Add(11 * time.Second)
	ts4 := ts3.Add(13 * time.Second)

	items := []types.StateInfo{
		{CreatedAt: ts1},
		{CreatedAt: ts2},
		{CreatedAt: ts3},
		{CreatedAt: ts4},
	}
	for i := range items {
		items[i].StateInfoIndex.RollappId = strconv.Itoa(i + 1)
		items[i].StateInfoIndex.Index = 1 + uint64(i)

		k.SetStateInfo(ctx, items[i])
	}

	// delete all before ts3: only ts3 and ts4 should be found
	k.DeleteStateInfoUntilTimestamp(ctx, ts2.Add(time.Second))

	for _, item := range items {
		_, found := k.GetStateInfo(ctx,
			item.StateInfoIndex.RollappId,
			item.StateInfoIndex.Index,
		)

		foundTSKey := k.HasStateInfoTimestampKey(ctx, item)

		if item.CreatedAt.After(ts2) {
			assert.True(t, found)
			assert.True(t, foundTSKey)
			continue
		}
		assert.Falsef(t, found, "item %v", item)
		assert.False(t, foundTSKey)
	}
}

package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNSequencer(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Sequencer {
	items := make([]types.Sequencer, n)
	for i := range items {
		items[i].SequencerAddress = strconv.Itoa(i)

		keeper.SetSequencer(ctx, items[i])

		scheduler := types.Scheduler{
			SequencerAddress: items[i].SequencerAddress,
			Status:           types.Unspecified,
		}

		keeper.SetScheduler(ctx, scheduler)
	}
	return items
}

func TestSequencerGet(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetSequencer(ctx,
			item.SequencerAddress,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestSequencerRemove(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveSequencer(ctx,
			item.SequencerAddress,
		)
		_, found := keeper.GetSequencer(ctx,
			item.SequencerAddress,
		)
		require.False(t, found)

		keeper.RemoveScheduler(ctx,
			item.SequencerAddress,
		)
		_, found = keeper.GetScheduler(ctx,
			item.SequencerAddress,
		)
		require.False(t, found)
	}
}

func TestSequencerGetAll(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllSequencer(ctx)),
	)
}

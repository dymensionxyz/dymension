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

func createNSequencersByRollapp(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.SequencersByRollapp {
	items := make([]types.SequencersByRollapp, n)
	for i := range items {
		items[i].RollappId = strconv.Itoa(i)
		sequencers := createNSequencer(keeper, ctx, n)
		for _, sequencer := range sequencers {
			items[i].Sequencers.Addresses = append(items[i].Sequencers.Addresses, sequencer.SequencerAddress)
		}

		keeper.SetSequencersByRollapp(ctx, items[i])
	}
	return items
}

func TestSequencersByRollappGet(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencersByRollapp(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetSequencersByRollapp(ctx,
			item.RollappId,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestSequencersByRollappRemove(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencersByRollapp(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveSequencersByRollapp(ctx,
			item.RollappId,
		)
		_, found := keeper.GetSequencersByRollapp(ctx,
			item.RollappId,
		)
		require.False(t, found)
	}
}

func TestSequencersByRollappGetAll(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencersByRollapp(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllSequencerssByRollapp(ctx)),
	)
}

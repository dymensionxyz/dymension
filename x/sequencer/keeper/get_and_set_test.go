package keeper_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/stretchr/testify/require"
)

func TestSequencerGet(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(k, ctx, 10)
	for _, item := range items {
		item := item
		rst, found := k.GetSequencer(ctx,
			item.Address,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestSequencerGetAll(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(k, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(k.GetAllSequencers(ctx)),
	)
}

func TestSequencersByRollappGet(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(k, ctx, 10)
	rst := k.RollappSequencers(ctx,
		items[0].RollappId,
	)

	require.Equal(t, len(rst), len(items))
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(rst),
	)
}

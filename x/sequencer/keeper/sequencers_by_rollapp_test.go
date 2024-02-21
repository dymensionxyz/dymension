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

func TestSequencersByRollappGet(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(keeper, ctx, 10)
	rst := keeper.GetSequencersByRollapp(ctx,
		items[0].RollappId,
	)

	require.Equal(t, len(rst), len(items))
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(rst),
	)
}

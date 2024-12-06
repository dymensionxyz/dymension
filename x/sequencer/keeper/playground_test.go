package keeper_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~
// PLAYGROUND ONLY
// ~~~~~~~~~~~~~~~

// Storage and query operations work for the event queue
func TestPlaygroundPrune(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)

	k.PruneTestUtilSetSentinel(ctx)

	_, err := k.RealSequencer(ctx, types.SentinelSeqAddr)
	require.NoError(t, err)

	k.Prune(ctx)
	_, err = k.RealSequencer(ctx, types.SentinelSeqAddr)
	require.Error(t, err)
}

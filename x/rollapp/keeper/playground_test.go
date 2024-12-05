package keeper_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~
// PLAYGROUND ONLY
// ~~~~~~~~~~~~~~~

// Storage and query operations work for the event queue
func TestPlaygroundPrune(t *testing.T) {

	k, ctx := keepertest.RollappKeeper(t)

	for h := range 100 {
		e := types.LivenessEvent{
			RollappId: "foo",
			HubHeight: int64(h),
		}
		k.PutLivenessEvent(ctx, e)
	}
	k.Prune(ctx.WithBlockHeight(50))
	all := k.GetLivenessEvents(ctx, nil)
	for _, e := range all {
		require.True(t, e.HubHeight >= 50)
	}
}

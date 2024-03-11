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

// TODO: Test FinalizeQueue function
// TODO: Test FinalizeQueue function with failed states

func (suite *RollappTestSuite) TestGetPendingFinalizationQueue() {
	suite.SetupTest()

	initialheight := uint64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight))
	ctx := &suite.Ctx

	keeper := suite.App.RollappKeeper

	// Create a rollapp
	rollapp := suite.CreateDefaultRollapp()

	// Create a sequencer
	proposer := suite.CreateDefaultSequencer(*ctx, rollapp)

	// Create a state update
	_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	suite.Require().Nil(err)

	expectedFinalizedHeight := initialheight + keeper.DisputePeriodInBlocks(*ctx)

	// Get the pending finalization queue
	suite.Require().Len(keeper.GetPendingFinalizationQueue(*ctx, expectedFinalizedHeight-1), 0)
	suite.Require().Len(keeper.GetPendingFinalizationQueue(*ctx, expectedFinalizedHeight), 1)
	suite.Require().Len(keeper.GetPendingFinalizationQueue(*ctx, expectedFinalizedHeight+5), 1)

}

func TestBlockHeightToFinalizationQueueGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetBlockHeightToFinalizationQueue(ctx,
			item.CreationHeight,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestBlockHeightToFinalizationQueueRemove(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveBlockHeightToFinalizationQueue(ctx,
			item.CreationHeight,
		)
		_, found := keeper.GetBlockHeightToFinalizationQueue(ctx,
			item.CreationHeight,
		)
		require.False(t, found)
	}
}

func TestBlockHeightToFinalizationQueueGetAll(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllBlockHeightToFinalizationQueue(ctx)),
	)
}

/* ---------------------------------- utils --------------------------------- */

func createNBlockHeightToFinalizationQueue(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.BlockHeightToFinalizationQueue {
	items := make([]types.BlockHeightToFinalizationQueue, n)
	for i := range items {
		items[i].CreationHeight = uint64(i)

		keeper.SetBlockHeightToFinalizationQueue(ctx, items[i])
	}
	return items
}

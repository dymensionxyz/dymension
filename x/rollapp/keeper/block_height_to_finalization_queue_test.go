package keeper_test

import (
	"slices"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func (suite *RollappTestSuite) TestGetAllFinalizationQueueUntilHeight() {
	suite.SetupTest()
	initialheight := uint64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight))
	ctx := &suite.Ctx
	keeper := suite.App.RollappKeeper

	rollapp := suite.CreateDefaultRollapp()
	proposer := suite.CreateDefaultSequencer(*ctx, rollapp)
	// Create 2 state updates
	_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	suite.Require().Nil(err)
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + 1))
	_, err = suite.PostStateUpdate(*ctx, rollapp, proposer, 11, uint64(10))
	suite.Require().Nil(err)

	// Get the pending finalization queue
	suite.Assert().Len(keeper.GetAllFinalizationQueueUntilHeight(*ctx, initialheight-1), 0)
	suite.Assert().Len(keeper.GetAllFinalizationQueueUntilHeight(*ctx, initialheight), 1)
	suite.Assert().Len(keeper.GetAllFinalizationQueueUntilHeight(*ctx, initialheight+1), 2)
	suite.Assert().Len(keeper.GetAllFinalizationQueueUntilHeight(*ctx, initialheight+100), 2)
}

func TestBlockHeightToFinalizationQueueGet(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(keeper, ctx, 10)
	for _, item := range items {
		item := item
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

// TODO: Test FinalizeQueue function with failed states
func (suite *RollappTestSuite) TestFinalize() {
	suite.SetupTest()

	initialheight := uint64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight))
	ctx := &suite.Ctx

	keeper := suite.App.RollappKeeper

	// Create a rollapp
	rollapp := suite.CreateDefaultRollapp()
	proposer := suite.CreateDefaultSequencer(*ctx, rollapp)

	// Create 2 state updates
	_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	suite.Require().Nil(err)

	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + 1))
	_, err = suite.PostStateUpdate(*ctx, rollapp, proposer, 11, uint64(10))
	suite.Require().Nil(err)

	// Finalize pending queues and check
	response := suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})
	suite.Require().Len(keeper.GetAllBlockHeightToFinalizationQueue(*ctx), 2)
	suite.False(findEvent(response, types.EventTypeStatusChange))

	// Finalize pending queues and check
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + keeper.DisputePeriodInBlocks(*ctx)))
	response = suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})
	suite.Require().Len(keeper.GetAllBlockHeightToFinalizationQueue(*ctx), 1)
	suite.True(findEvent(response, types.EventTypeStatusChange))

	// Finalize pending queues and check
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + keeper.DisputePeriodInBlocks(*ctx) + 1))
	response = suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})
	suite.Require().Len(keeper.GetAllBlockHeightToFinalizationQueue(*ctx), 0)
	suite.True(findEvent(response, types.EventTypeStatusChange))
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

func findEvent(response abci.ResponseEndBlock, eventType string) bool {
	return slices.ContainsFunc(response.Events, func(e abci.Event) bool { return e.Type == eventType })
}

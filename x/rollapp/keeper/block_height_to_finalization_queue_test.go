package keeper_test

import (
	"errors"
	"slices"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func (suite *RollappTestSuite) TestGetAllFinalizationQueueUntilHeight() {
	suite.SetupTest()
	initialHeight := uint64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialHeight))
	ctx := &suite.Ctx
	keeper := suite.App.RollappKeeper

	rollapp := suite.CreateDefaultRollapp()
	proposer := suite.CreateDefaultSequencer(*ctx, rollapp)
	// Create 2 state updates
	_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	suite.Require().Nil(err)
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialHeight + 1))
	_, err = suite.PostStateUpdate(*ctx, rollapp, proposer, 11, uint64(10))
	suite.Require().Nil(err)

	// Get the pending finalization queue
	suite.Len(keeper.GetAllFinalizationQueueUntilHeightInclusive(*ctx, initialHeight-1), 0)
	suite.Len(keeper.GetAllFinalizationQueueUntilHeightInclusive(*ctx, initialHeight), 1)
	suite.Len(keeper.GetAllFinalizationQueueUntilHeightInclusive(*ctx, initialHeight+1), 2)
	suite.Len(keeper.GetAllFinalizationQueueUntilHeightInclusive(*ctx, initialHeight+100), 2)
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
func (suite *RollappTestSuite) TestFinalizeRollapps() {
	type stateUpdate struct {
		blockHeight int64
		startHeight uint64
		numBlocks   uint64
	}
	type queue struct {
		rollappsLeft int
	}
	type blockEnd struct {
		blockHeight      func() int64
		wantNumFinalized int
		wantQueue        []queue
	}
	type rollappStateUpdate struct {
		stateUpdate             stateUpdate
		malleatePostStateUpdate func(rollappId string)
	}
	type fields struct {
		rollappStateUpdates []rollappStateUpdate
		blockEnd            blockEnd
	}

	const initialHeight int64 = 10

	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "finalize two rollapps successfully",
			fields: fields{
				rollappStateUpdates: []rollappStateUpdate{
					{
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
					}, {
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
					},
				},
				blockEnd: blockEnd{
					blockHeight:      func() int64 { return initialHeight + int64(suite.App.RollappKeeper.DisputePeriodInBlocks(suite.Ctx)) },
					wantQueue:        nil,
					wantNumFinalized: 2,
				},
			},
		}, {
			name: "finalize two rollapps, one with failed state",
			fields: fields{
				rollappStateUpdates: []rollappStateUpdate{
					{
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
					}, {
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
						malleatePostStateUpdate: func(rollappId string) {
							suite.App.RollappKeeper.RemoveStateInfo(suite.Ctx, rollappId, 1)
						},
					},
				},
				blockEnd: blockEnd{
					blockHeight:      func() int64 { return initialHeight + int64(suite.App.RollappKeeper.DisputePeriodInBlocks(suite.Ctx)) },
					wantQueue:        []queue{{rollappsLeft: 1}},
					wantNumFinalized: 1,
				},
			},
		}, {
			name: "finalize five rollapps, two with failed state",
			fields: fields{
				rollappStateUpdates: []rollappStateUpdate{
					{
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
					}, {
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
					}, {
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
					}, {
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
						malleatePostStateUpdate: func(rollappId string) {
							suite.App.RollappKeeper.RemoveStateInfo(suite.Ctx, rollappId, 1)
						},
					}, {
						stateUpdate: stateUpdate{
							blockHeight: initialHeight,
							startHeight: 1,
							numBlocks:   10,
						},
						malleatePostStateUpdate: func(rollappId string) {
							suite.App.RollappKeeper.RemoveStateInfo(suite.Ctx, rollappId, 1)
						},
					},
				},
				blockEnd: blockEnd{
					blockHeight: func() int64 { return initialHeight + int64(suite.App.RollappKeeper.DisputePeriodInBlocks(suite.Ctx)) },
					wantQueue: []queue{
						{
							rollappsLeft: 2,
						},
					},
					wantNumFinalized: 3,
				},
			},
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			suite.SetupTest()
			ctx := &suite.Ctx

			for _, rf := range tt.fields.rollappStateUpdates {
				// Create a rollapp
				rollapp := suite.CreateDefaultRollapp()
				proposer := suite.CreateDefaultSequencer(*ctx, rollapp)

				// Create state update
				su := rf.stateUpdate
				suite.Ctx = suite.Ctx.WithBlockHeight(su.blockHeight)
				_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, su.startHeight, su.numBlocks)
				suite.Require().Nil(err)

				if rf.malleatePostStateUpdate != nil {
					rf.malleatePostStateUpdate(rollapp)
				}
			}
			// End block and check if finalized
			be := tt.fields.blockEnd
			suite.Ctx = suite.Ctx.WithBlockHeight(be.blockHeight())
			response := suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})

			heightQueue := suite.App.RollappKeeper.GetAllBlockHeightToFinalizationQueue(*ctx)
			suite.Require().Len(heightQueue, len(be.wantQueue))

			for i, q := range be.wantQueue {
				suite.Require().Len(heightQueue[i].FinalizationQueue, q.rollappsLeft)
			}

			numFinalized := countFinalized(response)
			suite.Assert().Equal(be.wantNumFinalized, numFinalized)
		})
	}
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

func countFinalized(response abci.ResponseEndBlock) int {
	count := 0
	for _, event := range response.Events {
		if event.Type == types.EventTypeStatusChange {
			count++
		}
	}
	return count
}

func findEvent(response abci.ResponseEndBlock, eventType string) bool {
	return slices.ContainsFunc(response.Events, func(e abci.Event) bool { return e.Type == eventType })
}

//nolint:govet
func (suite *RollappTestSuite) TestKeeperFinalizePending() {
	tests := []struct {
		name                     string
		pendingFinalizationQueue []types.BlockHeightToFinalizationQueue
		errFinalizeIndices       []types.StateInfoIndex
		expectQueueAfter         []types.BlockHeightToFinalizationQueue
	}{
		{
			name: "finalize pending: all rollapps successfully finalized",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 1},
						{RollappId: "rollapp2", Index: 2},
						{RollappId: "rollapp1", Index: 2},
						{RollappId: "rollapp3", Index: 1},
						{RollappId: "rollapp2", Index: 3},
						{RollappId: "rollapp3", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 3},
						{RollappId: "rollapp2", Index: 4},
						{RollappId: "rollapp1", Index: 4},
						{RollappId: "rollapp3", Index: 3},
						{RollappId: "rollapp2", Index: 5},
						{RollappId: "rollapp3", Index: 4},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{},
			expectQueueAfter:   nil,
		}, {
			name: "finalize pending: 2 rollapps failed at 1 height",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 1},
						{RollappId: "rollapp2", Index: 2},
						{RollappId: "rollapp1", Index: 2},
						{RollappId: "rollapp3", Index: 1},
						{RollappId: "rollapp1", Index: 3},
						{RollappId: "rollapp3", Index: 2},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{{"rollapp1", 2}, {"rollapp3", 2}},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 2},
						{RollappId: "rollapp1", Index: 3},
						{RollappId: "rollapp3", Index: 2},
					},
				},
			},
		}, {
			name: "finalize pending: 2 rollapps failed at 2 heights",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 1},
						{RollappId: "rollapp2", Index: 1},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 2},
						{RollappId: "rollapp2", Index: 2},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{{"rollapp1", 1}, {"rollapp2", 2}},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 1},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp2", Index: 2},
					},
				},
			},
		}, {
			name: "finalize pending: 2 rollapps failed at 2 heights, one in each height",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 1},
						{RollappId: "rollapp2", Index: 2},
						{RollappId: "rollapp1", Index: 2},
						{RollappId: "rollapp3", Index: 1},
						{RollappId: "rollapp2", Index: 3},
						{RollappId: "rollapp3", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 3},
						{RollappId: "rollapp2", Index: 4},
						{RollappId: "rollapp1", Index: 4},
						{RollappId: "rollapp3", Index: 3},
						{RollappId: "rollapp2", Index: 5},
						{RollappId: "rollapp3", Index: 4},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{{"rollapp1", 2}, {"rollapp3", 4}},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp3", Index: 4},
					},
				},
			},
		}, {
			name: "finalize pending: all rollapps failed to finalize",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 1},
						{RollappId: "rollapp2", Index: 2},
						{RollappId: "rollapp1", Index: 2},
						{RollappId: "rollapp3", Index: 1},
						{RollappId: "rollapp2", Index: 3},
						{RollappId: "rollapp3", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 3},
						{RollappId: "rollapp2", Index: 4},
						{RollappId: "rollapp1", Index: 4},
						{RollappId: "rollapp3", Index: 3},
						{RollappId: "rollapp2", Index: 5},
						{RollappId: "rollapp3", Index: 4},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{
				{RollappId: "rollapp1", Index: 1},
				{RollappId: "rollapp2", Index: 2},
				{RollappId: "rollapp3", Index: 1},
				{RollappId: "rollapp1", Index: 3},
				{RollappId: "rollapp2", Index: 4},
				{RollappId: "rollapp3", Index: 3},
			},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 1},
						{RollappId: "rollapp2", Index: 2},
						{RollappId: "rollapp1", Index: 2},
						{RollappId: "rollapp3", Index: 1},
						{RollappId: "rollapp2", Index: 3},
						{RollappId: "rollapp3", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp1", Index: 3},
						{RollappId: "rollapp2", Index: 4},
						{RollappId: "rollapp1", Index: 4},
						{RollappId: "rollapp3", Index: 3},
						{RollappId: "rollapp2", Index: 5},
						{RollappId: "rollapp3", Index: 4},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			suite.SetupTest()

			k := suite.App.RollappKeeper
			k.SetFinalizePendingFn(MockFinalizePending(tt.errFinalizeIndices))
			k.FinalizePending(suite.Ctx, tt.pendingFinalizationQueue)

			suite.Require().Equal(tt.expectQueueAfter, k.GetAllBlockHeightToFinalizationQueue(suite.Ctx))
		})
	}
}

func MockFinalizePending(errFinalizedIndices []types.StateInfoIndex) func(ctx sdk.Context, stateInfoIndex types.StateInfoIndex) error {
	return func(ctx sdk.Context, stateInfoIndex types.StateInfoIndex) error {
		if slices.Contains(errFinalizedIndices, stateInfoIndex) {
			return errors.New("error")
		}
		return nil
	}
}

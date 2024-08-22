package keeper_test

import (
	"errors"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"unsafe"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

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
	k := suite.App.RollappKeeper

	rollapp, proposer := suite.CreateDefaultRollappAndProposer()
	// Create 2 state updates
	_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	suite.Require().Nil(err)
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialHeight + 1))
	_, err = suite.PostStateUpdate(*ctx, rollapp, proposer, 11, uint64(10))
	suite.Require().Nil(err)

	// Get the pending finalization queue
	suite.Len(k.GetAllFinalizationQueueUntilHeightInclusive(*ctx, initialHeight-1), 0)
	suite.Len(k.GetAllFinalizationQueueUntilHeightInclusive(*ctx, initialHeight), 1)
	suite.Len(k.GetAllFinalizationQueueUntilHeightInclusive(*ctx, initialHeight+1), 2)
	suite.Len(k.GetAllFinalizationQueueUntilHeightInclusive(*ctx, initialHeight+100), 2)
}

func TestBlockHeightToFinalizationQueueGet(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(k, ctx, 10)
	for _, item := range items {
		item := item
		rst, found := k.GetBlockHeightToFinalizationQueue(ctx,
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
	k, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(k, ctx, 10)
	for _, item := range items {
		k.RemoveBlockHeightToFinalizationQueue(ctx,
			item.CreationHeight,
		)
		_, found := k.GetBlockHeightToFinalizationQueue(ctx,
			item.CreationHeight,
		)
		require.False(t, found)
	}
}

func TestBlockHeightToFinalizationQueueGetAll(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(k, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(k.GetAllBlockHeightToFinalizationQueue(ctx)),
	)
}

//nolint:gofumpt
func (suite *RollappTestSuite) TestFinalizeRollapps() {
	suite.SetupTest()

	type rollappQueue struct {
		rollappId string
		index     uint64
	}
	type queue struct {
		rollappsLeft []rollappQueue
	}
	type blockEnd struct {
		wantNumFinalized int
		wantQueue        []queue
		recovers         map[types.StateInfoIndex]struct{}
	}
	type stateUpdate struct {
		blockHeight int64
		startHeight uint64
		numOfBlocks uint64
		fail        bool
	}
	type rollappStateUpdate struct {
		rollappId    string
		stateUpdates []stateUpdate
	}
	type fields struct {
		rollappStateUpdates []rollappStateUpdate
		finalizations       []blockEnd
	}

	const initialHeight int64 = 10
	getDisputePeriod := func() int64 {
		return int64(suite.App.RollappKeeper.DisputePeriodInBlocks(suite.Ctx))
	}

	getFinalizationHeight := func(n int64) int64 {
		return initialHeight + getDisputePeriod()*n
	}

	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "finalize two rollapps in one finalization successfully",
			fields: fields{
				rollappStateUpdates: []rollappStateUpdate{
					{
						rollappId: "rollapp_1234-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10,
						}},
					}, {
						rollappId: "rollappa_2345-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10,
						}},
					},
				},
				finalizations: []blockEnd{
					{
						wantNumFinalized: 2,
						wantQueue:        nil,
					},
				},
			},
		}, {
			name: "finalize two rollapps, one with failed state",
			fields: fields{
				rollappStateUpdates: []rollappStateUpdate{
					{
						rollappId: "rollapp_1234-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10,
						}},
					}, {
						rollappId: "rollappa_2345-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10, fail: true,
						}},
					},
				},
				finalizations: []blockEnd{
					{
						wantNumFinalized: 1,
						wantQueue: []queue{{
							rollappsLeft: []rollappQueue{{
								rollappId: "rollappa_2345-1",
								index:     1,
							}},
						}},
					},
				},
			},
		}, {
			name: "finalize five rollapps, three with failed state",
			fields: fields{
				rollappStateUpdates: []rollappStateUpdate{
					{
						rollappId: "rollapp_1234-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10,
						}, {
							blockHeight: initialHeight, startHeight: 11, numOfBlocks: 20,
						}},
					}, {
						rollappId: "rollappa_2345-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10,
						}},
					}, {
						rollappId: "rollappe_3456-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10,
						}},
					}, {
						rollappId: "rollappi_4567-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10,
							fail: true,
						}, {
							blockHeight: initialHeight + getDisputePeriod(), startHeight: 11, numOfBlocks: 20,
							fail: true,
						}},
					}, {
						rollappId: "rollappo_5678-1",
						stateUpdates: []stateUpdate{{
							blockHeight: initialHeight, startHeight: 1, numOfBlocks: 10,
							fail: true,
						}},
					},
				},
				finalizations: []blockEnd{
					{
						// first finalization: 4 states finalized, 3 states left
						wantNumFinalized: 4,
						wantQueue: []queue{
							{
								rollappsLeft: []rollappQueue{
									{
										rollappId: "rollappi_4567-1",
										index:     1,
									}, {
										rollappId: "rollappo_5678-1",
										index:     1,
									},
								},
							}, {
								rollappsLeft: []rollappQueue{
									{
										rollappId: "rollappi_4567-1",
										index:     2,
									},
								},
							},
						},
					}, {
						// second finalization: 1 state finalized from first finalization, 2 states left
						wantNumFinalized: 1,
						recovers: map[types.StateInfoIndex]struct{}{
							{RollappId: "rollappi_4567-1", Index: 1}: {},
						},
						wantQueue: []queue{
							{
								rollappsLeft: []rollappQueue{
									{
										rollappId: "rollappo_5678-1",
										index:     1,
									},
								},
							}, {
								rollappsLeft: []rollappQueue{
									{
										rollappId: "rollappi_4567-1",
										index:     2,
									},
								},
							},
						},
					}, {
						// third finalization: 1 state finalized from first finalization, 1 state left
						wantNumFinalized: 1,
						recovers: map[types.StateInfoIndex]struct{}{
							{RollappId: "rollappo_5678-1", Index: 1}: {},
						},
						wantQueue: []queue{
							{
								rollappsLeft: []rollappQueue{
									{
										rollappId: "rollappi_4567-1",
										index:     2,
									},
								},
							},
						},
					}, {
						// fourth finalization: 1 state finalized from first finalization, 0 states left
						wantNumFinalized: 1,
						recovers: map[types.StateInfoIndex]struct{}{
							{RollappId: "rollappi_4567-1", Index: 2}: {},
						},
						wantQueue: nil,
					},
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
				suite.CreateRollappByName(rf.rollappId)
				proposer := suite.CreateDefaultSequencer(suite.Ctx, rf.rollappId)

				// Create state update
				for _, su := range rf.stateUpdates {
					suite.Ctx = suite.Ctx.WithBlockHeight(su.blockHeight)
					_, err := suite.PostStateUpdate(*ctx, rf.rollappId, proposer, su.startHeight, su.numOfBlocks)
					suite.Require().Nil(err)
				}
			}

			// prepare hooks for failed state updates
			var errFinalizeIndexes []types.StateInfoIndex
			for _, rf := range tt.fields.rollappStateUpdates {
				for i, su := range rf.stateUpdates {
					if su.fail {
						errFinalizeIndexes = append(errFinalizeIndexes, types.StateInfoIndex{
							RollappId: rf.rollappId,
							Index:     uint64(i + 1),
						})
					}
				}
			}
			suite.setMockErrRollappKeeperHooks(errFinalizeIndexes)
			// run finalizations and check finalized state updates
			for i, be := range tt.fields.finalizations {
				errFinalizeIndexes = slices.DeleteFunc(errFinalizeIndexes, func(e types.StateInfoIndex) bool {
					_, ok := be.recovers[e]
					return ok
				})

				suite.Ctx = suite.Ctx.WithBlockHeight(getFinalizationHeight(int64(i + 1)))
				response := suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})

				numFinalized := countFinalized(response)
				suite.Require().Equalf(be.wantNumFinalized, numFinalized, "finalization %d", i+1)

				heightQueue := suite.App.RollappKeeper.GetAllBlockHeightToFinalizationQueue(*ctx)
				suite.Require().Lenf(heightQueue, len(be.wantQueue), "finalization %d", i+1)

				for i, q := range be.wantQueue {
					suite.Require().Lenf(heightQueue[i].FinalizationQueue, len(q.rollappsLeft), "finalization %d", i+1)

					for j, r := range q.rollappsLeft {
						suite.Require().Equalf(heightQueue[i].FinalizationQueue[j].RollappId, r.rollappId, "finalization %d, rollappLeft: %d", i+1, j+1)
						suite.Require().Equalf(heightQueue[i].FinalizationQueue[j].Index, r.index, "finalization %d, rollappLeft: %d", i+1, j+1)
					}
				}
			}
		})
	}
}

// TODO: Test FinalizeQueue function with failed states
func (suite *RollappTestSuite) TestFinalize() {
	suite.SetupTest()

	initialheight := uint64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight))
	ctx := &suite.Ctx

	k := suite.App.RollappKeeper

	// Create a rollapp
	rollapp, proposer := suite.CreateDefaultRollappAndProposer()

	// Create 2 state updates
	_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	suite.Require().Nil(err)

	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + 1))
	_, err = suite.PostStateUpdate(*ctx, rollapp, proposer, 11, uint64(10))
	suite.Require().Nil(err)

	// Finalize pending queues and check
	response := suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})
	suite.Require().Len(k.GetAllBlockHeightToFinalizationQueue(*ctx), 2)
	suite.False(findEvent(response, types.EventTypeStatusChange))

	// Finalize pending queues and check
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + k.DisputePeriodInBlocks(*ctx)))
	response = suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})
	suite.Require().Len(k.GetAllBlockHeightToFinalizationQueue(*ctx), 1)
	suite.True(findEvent(response, types.EventTypeStatusChange))

	// Finalize pending queues and check
	suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + k.DisputePeriodInBlocks(*ctx) + 1))
	response = suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})
	suite.Require().Len(k.GetAllBlockHeightToFinalizationQueue(*ctx), 0)
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
						{RollappId: "rollapp_1234-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollappa_2345-1", Index: 4},
						{RollappId: "rollapp_1234-1", Index: 4},
						{RollappId: "rollappe_3456-1", Index: 3},
						{RollappId: "rollappa_2345-1", Index: 5},
						{RollappId: "rollappe_3456-1", Index: 4},
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
						{RollappId: "rollapp_1234-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{{"rollapp_1234-1", 2}, {"rollappe_3456-1", 2}},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
				},
			},
		}, {
			name: "finalize pending: 2 rollapps failed at 2 heights",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 1},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollappa_2345-1", Index: 2},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{{"rollapp_1234-1", 1}, {"rollappa_2345-1", 2}},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollappa_2345-1", Index: 2},
					},
				},
			},
		}, {
			name: "finalize pending: 2 rollapps failed at 2 heights, one in each height",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollappa_2345-1", Index: 4},
						{RollappId: "rollapp_1234-1", Index: 4},
						{RollappId: "rollappe_3456-1", Index: 3},
						{RollappId: "rollappa_2345-1", Index: 5},
						{RollappId: "rollappe_3456-1", Index: 4},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{{"rollapp_1234-1", 2}, {"rollappe_3456-1", 4}},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollapp_1234-1", Index: 4},
						{RollappId: "rollappe_3456-1", Index: 4},
					},
				},
			},
		}, {
			name: "finalize pending: all rollapps failed to finalize",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollappa_2345-1", Index: 4},
						{RollappId: "rollapp_1234-1", Index: 4},
						{RollappId: "rollappe_3456-1", Index: 3},
						{RollappId: "rollappa_2345-1", Index: 5},
						{RollappId: "rollappe_3456-1", Index: 4},
					},
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{
				{RollappId: "rollapp_1234-1", Index: 1},
				{RollappId: "rollappa_2345-1", Index: 2},
				{RollappId: "rollappe_3456-1", Index: 1},
			},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappa_2345-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollappa_2345-1", Index: 4},
						{RollappId: "rollapp_1234-1", Index: 4},
						{RollappId: "rollappe_3456-1", Index: 3},
						{RollappId: "rollappa_2345-1", Index: 5},
						{RollappId: "rollappe_3456-1", Index: 4},
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
			k.FinalizeAllPending(suite.Ctx, tt.pendingFinalizationQueue)

			finalizationQueue := k.GetAllBlockHeightToFinalizationQueue(suite.Ctx)
			suite.Require().Equal(tt.expectQueueAfter, finalizationQueue)
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

// black-ops: don't do this at home
// nolint:gosec
func (suite *RollappTestSuite) setMockErrRollappKeeperHooks(failIndexes []types.StateInfoIndex) {
	k := suite.App.RollappKeeper
	v := reflect.ValueOf(k).Elem()
	f := v.FieldByName("hooks")
	hooks := mockRollappHooks{failIndexes: failIndexes}

	if f.CanSet() {
		f.Set(reflect.ValueOf(types.MultiRollappHooks{hooks}))
	} else {
		reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(types.MultiRollappHooks{hooks}))
	}
}

var _ types.RollappHooks = mockRollappHooks{}

type mockRollappHooks struct {
	types.StubRollappCreatedHooks
	failIndexes []types.StateInfoIndex
}

func (m mockRollappHooks) AfterStateFinalized(_ sdk.Context, _ string, stateInfo *types.StateInfo) (err error) {
	if slices.Contains(m.failIndexes, stateInfo.StateInfoIndex) {
		return errors.New("error")
	}
	return
}

package keeper_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"unsafe"

	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func (s *RollappTestSuite) TestGetAllFinalizationQueueUntilHeight() {
	s.SetupTest()
	initialHeight := uint64(10)
	s.Ctx = s.Ctx.WithBlockHeight(int64(initialHeight))
	ctx := &s.Ctx
	k := s.k()

	rollapp, proposer := s.CreateDefaultRollappAndProposer()
	// Create 2 state updates
	_, err := s.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	s.Require().Nil(err)
	s.Ctx = s.Ctx.WithBlockHeight(int64(initialHeight + 1))
	_, err = s.PostStateUpdate(*ctx, rollapp, proposer, 11, uint64(10))
	s.Require().Nil(err)

	// Get the pending finalization queue
	testCases := []struct {
		height      uint64
		expectedLen int
	}{
		{height: initialHeight - 1, expectedLen: 0},
		{height: initialHeight, expectedLen: 1},
		{height: initialHeight + 1, expectedLen: 2},
		{height: initialHeight + 100, expectedLen: 2},
	}
	for _, tc := range testCases {
		actual, err := k.GetFinalizationQueueUntilHeightInclusive(*ctx, tc.height)
		s.Require().NoError(err)
		s.Require().Len(actual, tc.expectedLen)
	}
}

func TestBlockHeightToFinalizationQueueGet(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(k, ctx, 10)
	for _, item := range items {
		item := item
		rst, found := k.GetFinalizationQueue(
			ctx,
			item.CreationHeight,
			item.RollappId,
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
		err := k.RemoveFinalizationQueue(
			ctx,
			item.CreationHeight,
			item.RollappId,
		)
		require.NoError(t, err)
		_, found := k.GetFinalizationQueue(
			ctx,
			item.CreationHeight,
			item.RollappId,
		)
		require.False(t, found)
	}
}

func TestBlockHeightToFinalizationQueueGetAll(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	items := createNBlockHeightToFinalizationQueue(k, ctx, 10)
	queue, err := k.GetEntireFinalizationQueue(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(queue),
	)
}

func TestGetFinalizationQueueByRollapp(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)

	q1 := types.BlockHeightToFinalizationQueue{CreationHeight: 1, RollappId: "rollapp_1234-1"}
	q2 := types.BlockHeightToFinalizationQueue{CreationHeight: 2, RollappId: "rollapp_1234-1"}
	q3 := types.BlockHeightToFinalizationQueue{CreationHeight: 3, RollappId: "rollapp_1234-1"}

	k.MustSetFinalizationQueue(ctx, q1)
	k.MustSetFinalizationQueue(ctx, q2)
	k.MustSetFinalizationQueue(ctx, q3)

	// Check all queues
	q, err := k.GetEntireFinalizationQueue(ctx)
	require.NoError(t, err)
	require.Equal(t, []types.BlockHeightToFinalizationQueue{q1, q2, q3}, q)

	// Get all queues from different heights associated with a given rollapp
	q, err = k.GetFinalizationQueueByRollapp(ctx, "rollapp_1234-1")
	require.NoError(t, err)
	require.Equal(t, []types.BlockHeightToFinalizationQueue{q1, q2, q3}, q)

	// Remove one of the queues
	k.MustRemoveFinalizationQueue(ctx, 2, "rollapp_1234-1")

	// Verify the index is updated
	q, err = k.GetFinalizationQueueByRollapp(ctx, "rollapp_1234-1")
	require.NoError(t, err)
	require.Equal(t, []types.BlockHeightToFinalizationQueue{q1, q3}, q)

	// Verify height 2 is empty

	// Check all queues until height 3
	q, err = k.GetFinalizationQueueUntilHeightInclusive(ctx, 3)
	require.NoError(t, err)
	require.Equal(t, []types.BlockHeightToFinalizationQueue{q1, q3}, q)

	// Check all queues
	q, err = k.GetEntireFinalizationQueue(ctx)
	require.NoError(t, err)
	require.Equal(t, []types.BlockHeightToFinalizationQueue{q1, q3}, q)
}

//nolint:gofumpt
func (s *RollappTestSuite) TestFinalizeRollapps() {
	s.SetupTest()

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
		return int64(s.k().DisputePeriodInBlocks(s.Ctx))
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
									},
								},
							}, {
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
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()
			ctx := &s.Ctx

			for _, rf := range tt.fields.rollappStateUpdates {
				// Create a rollapp
				s.CreateRollappByName(rf.rollappId)
				proposer := s.CreateDefaultSequencer(s.Ctx, rf.rollappId)

				// Create state update
				for _, su := range rf.stateUpdates {
					s.Ctx = s.Ctx.WithBlockHeight(su.blockHeight)
					_, err := s.PostStateUpdate(*ctx, rf.rollappId, proposer, su.startHeight, su.numOfBlocks)
					s.Require().Nil(err)
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
			s.setMockErrRollappKeeperHooks(errFinalizeIndexes)
			// run finalizations and check finalized state updates
			for i, be := range tt.fields.finalizations {
				errFinalizeIndexes = slices.DeleteFunc(errFinalizeIndexes, func(e types.StateInfoIndex) bool {
					_, ok := be.recovers[e]
					return ok
				})

				s.Ctx = s.Ctx.WithBlockHeight(getFinalizationHeight(int64(i + 1)))
				response := s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})

				numFinalized := countFinalized(response)
				s.Require().Equalf(be.wantNumFinalized, numFinalized, "finalization %d", i+1)

				heightQueue, err := s.k().GetEntireFinalizationQueue(*ctx)
				s.Require().NoError(err)
				s.Require().Lenf(heightQueue, len(be.wantQueue), "finalization %d", i+1)

				for i, q := range be.wantQueue {
					s.Require().Lenf(heightQueue[i].FinalizationQueue, len(q.rollappsLeft), "finalization %d", i+1)

					for j, r := range q.rollappsLeft {
						s.Require().Equalf(heightQueue[i].FinalizationQueue[j].RollappId, r.rollappId, "finalization %d, rollappLeft: %d", i+1, j+1)
						s.Require().Equalf(heightQueue[i].FinalizationQueue[j].Index, r.index, "finalization %d, rollappLeft: %d", i+1, j+1)
					}
				}
			}
		})
	}
}

// TODO: Test FinalizeQueue function with failed states
func (s *RollappTestSuite) TestFinalize() {
	s.SetupTest()

	initialheight := uint64(10)
	s.Ctx = s.Ctx.WithBlockHeight(int64(initialheight))
	ctx := &s.Ctx

	k := s.k()

	// Create a rollapp
	rollapp, proposer := s.CreateDefaultRollappAndProposer()

	// Create 2 state updates
	_, err := s.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	s.Require().Nil(err)

	s.Ctx = s.Ctx.WithBlockHeight(int64(initialheight + 1))
	_, err = s.PostStateUpdate(*ctx, rollapp, proposer, 11, uint64(10))
	s.Require().Nil(err)

	// Finalize pending queues and check
	response := s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})
	actualQueue, err := k.GetEntireFinalizationQueue(*ctx)
	s.Require().NoError(err)
	s.Require().Len(actualQueue, 2)
	s.False(findEvent(response, types.EventTypeStatusChange))

	// Finalize pending queues and check
	s.Ctx = s.Ctx.WithBlockHeight(int64(initialheight + k.DisputePeriodInBlocks(*ctx)))
	response = s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})
	actualQueue, err = k.GetEntireFinalizationQueue(*ctx)
	s.Require().NoError(err)
	s.Require().Len(actualQueue, 1)
	s.True(findEvent(response, types.EventTypeStatusChange))

	// Finalize pending queues and check
	s.Ctx = s.Ctx.WithBlockHeight(int64(initialheight + k.DisputePeriodInBlocks(*ctx) + 1))
	response = s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})
	actualQueue, err = k.GetEntireFinalizationQueue(*ctx)
	s.Require().NoError(err)
	s.Require().Len(actualQueue, 0)
	s.True(findEvent(response, types.EventTypeStatusChange))
}

/* ---------------------------------- utils --------------------------------- */
func createNBlockHeightToFinalizationQueue(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.BlockHeightToFinalizationQueue {
	items := make([]types.BlockHeightToFinalizationQueue, n)
	for i := range items {
		items[i].CreationHeight = uint64(i)
		items[i].RollappId = fmt.Sprintf("rollapp_%d-1", i)
		keeper.MustSetFinalizationQueue(ctx, items[i])
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
func (s *RollappTestSuite) TestKeeperFinalizePending() {
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
						{RollappId: "rollapp_1234-1", Index: 2},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollappa_2345-1", Index: 3},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
					RollappId: "rollappe_3456-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollapp_1234-1", Index: 4},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 4},
						{RollappId: "rollappa_2345-1", Index: 5},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 4},
					},
					RollappId: "rollappe_3456-1",
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
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollapp_1234-1", Index: 3},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 2},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
					RollappId: "rollappe_3456-1",
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{
				{RollappId: "rollapp_1234-1", Index: 2},
				{RollappId: "rollappe_3456-1", Index: 2},
			},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 2},
						{RollappId: "rollapp_1234-1", Index: 3},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 2},
					},
					RollappId: "rollappe_3456-1",
				},
			},
		}, {
			name: "finalize pending: 2 rollapps failed at 2 heights",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 1},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 2},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 2},
					},
					RollappId: "rollappa_2345-1",
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{
				{RollappId: "rollapp_1234-1", Index: 1},
				{RollappId: "rollappa_2345-1", Index: 2},
			},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 2},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 2},
					},
					RollappId: "rollappa_2345-1",
				},
			},
		}, {
			name: "finalize pending: 2 rollapps failed at 2 heights, one in each height",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
						{RollappId: "rollapp_1234-1", Index: 2},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollappa_2345-1", Index: 3},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
					RollappId: "rollappe_3456-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollapp_1234-1", Index: 4},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 4},
						{RollappId: "rollappa_2345-1", Index: 5},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 4},
					},
					RollappId: "rollappe_3456-1",
				},
			},
			errFinalizeIndices: []types.StateInfoIndex{
				{RollappId: "rollapp_1234-1", Index: 2},
				{RollappId: "rollappe_3456-1", Index: 4},
			},
			expectQueueAfter: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 2},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollapp_1234-1", Index: 4},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 4},
					},
					RollappId: "rollappe_3456-1",
				},
			},
		}, {
			name: "finalize pending: all rollapps failed to finalize",
			pendingFinalizationQueue: []types.BlockHeightToFinalizationQueue{
				{
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 1},
						{RollappId: "rollapp_1234-1", Index: 2},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollappa_2345-1", Index: 3},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
					RollappId: "rollappe_3456-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollapp_1234-1", Index: 4},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 4},
						{RollappId: "rollappa_2345-1", Index: 5},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 4},
					},
					RollappId: "rollappe_3456-1",
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
						{RollappId: "rollapp_1234-1", Index: 2},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 2},
						{RollappId: "rollappa_2345-1", Index: 3},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 1,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 1},
						{RollappId: "rollappe_3456-1", Index: 2},
					},
					RollappId: "rollappe_3456-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollapp_1234-1", Index: 3},
						{RollappId: "rollapp_1234-1", Index: 4},
					},
					RollappId: "rollapp_1234-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappa_2345-1", Index: 4},
						{RollappId: "rollappa_2345-1", Index: 5},
					},
					RollappId: "rollappa_2345-1",
				}, {
					CreationHeight: 2,
					FinalizationQueue: []types.StateInfoIndex{
						{RollappId: "rollappe_3456-1", Index: 3},
						{RollappId: "rollappe_3456-1", Index: 4},
					},
					RollappId: "rollappe_3456-1",
				},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			k := s.k()
			for _, item := range tt.pendingFinalizationQueue {
				k.MustSetFinalizationQueue(s.Ctx, item)
			}
			k.SetFinalizePendingFn(MockFinalizePending(tt.errFinalizeIndices))
			k.FinalizeAllPending(s.Ctx, tt.pendingFinalizationQueue)

			finalizationQueue, err := k.GetEntireFinalizationQueue(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tt.expectQueueAfter, finalizationQueue)
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
func (s *RollappTestSuite) setMockErrRollappKeeperHooks(failIndexes []types.StateInfoIndex) {
	k := s.k()
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

func TestUnbondConditionFlow(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)

	seq := keepertest.Alice

	err := k.CanUnbond(ctx, seq)
	require.NoError(t, err)

	for h := range 10 {
		err := k.SaveSequencerHeight(ctx, seq.Address, uint64(h))
		require.NoError(t, err)
	}

	pairs, err := k.AllSequencerHeightPairs(ctx)
	require.NoError(t, err)
	require.Len(t, pairs, 10)

	err = k.CanUnbond(ctx, seq)
	require.True(t, errorsmod.IsOf(err, sequencertypes.ErrUnbondNotAllowed))

	err = k.PruneSequencerHeights(ctx, []string{seq.Address}, 6)
	require.NoError(t, err)
	pairs, err = k.AllSequencerHeightPairs(ctx)
	require.NoError(t, err)
	require.Len(t, pairs, 7) // removed heights above 6

	err = k.CanUnbond(ctx, seq)
	require.True(t, errorsmod.IsOf(err, sequencertypes.ErrUnbondNotAllowed))

	for h := range 7 {
		err := k.DelSequencerHeight(ctx, seq.Address, uint64(h))
		require.NoError(t, err)
	}

	err = k.CanUnbond(ctx, seq)
	require.NoError(t, err)
}

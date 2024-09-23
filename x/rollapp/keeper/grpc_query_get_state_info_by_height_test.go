package keeper_test

import (
	"math/rand"
	"strconv"
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

const (
	maxNumOfBlocks = 1000
)

func createNStateInfoAndIndex(keeper *keeper.Keeper, ctx sdk.Context, n int, rollappId string) []types.StateInfo {
	keeper.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappId,
	})
	items := make([]types.StateInfo, n)
	StartHeight := uint64(1)
	for i := range items {
		numBlocks := uint64(rand.Intn(maxNumOfBlocks) + 1) //nolint:gosec // this is for a test
		stateInfo := types.StateInfo{
			StateInfoIndex: types.StateInfoIndex{
				RollappId: rollappId,
				Index:     uint64(i + 1),
			},
			StartHeight: StartHeight,
			BDs:         types.BlockDescriptors{BD: make([]types.BlockDescriptor, numBlocks)},
		}
		StartHeight += stateInfo.NumBlocks()

		keeper.SetStateInfo(ctx, stateInfo)
		keeper.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
			RollappId: rollappId,
			Index:     stateInfo.StateInfoIndex.Index,
		})

		items[i] = stateInfo
	}
	keeper.SetLatestFinalizedStateIndex(ctx, types.StateInfoIndex{
		RollappId: rollappId,
		Index:     items[n-1].StateInfoIndex.Index,
	})
	return items
}

func TestStateInfoByHeightLatestStateInfoIndex(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	rollappId := "rollappid_1234-1"
	k.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappId,
	})
	request := &types.QueryGetStateInfoRequest{
		RollappId: rollappId,
		Height:    100,
	}
	_, err := k.StateInfo(wctx, request)
	require.EqualError(t, err, errorsmod.Wrapf(types.ErrNotFound, "LatestStateInfoIndex wasn't found for rollappId=%s", rollappId).Error())
}

func TestStateInfoByHeightMissingStateInfo(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	rollappId := urand.RollappID()
	k.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappId,
	})
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: rollappId,
		Index:     uint64(85),
	})
	request := &types.QueryGetStateInfoRequest{
		RollappId: rollappId,
		Height:    100,
	}
	_, err := k.StateInfo(wctx, request)
	errIndex := 1 + (85-1)/2 // Using binary search, the middle index is lookedup first and is missing.
	require.EqualError(t, err, errorsmod.Wrapf(types.ErrNotFound,
		"StateInfo wasn't found for rollappId=%s, index=%d",
		rollappId, errIndex).Error())
}

func TestStateInfoByHeightMissingStateInfo1(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	rollappId := urand.RollappID()
	k.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappId,
	})
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: rollappId,
		Index:     uint64(60),
	})
	request := &types.QueryGetStateInfoRequest{
		RollappId: rollappId,
		Height:    70,
	}
	k.SetStateInfo(ctx, types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: rollappId, Index: 60},
		StartHeight:    71,
		NumBlocks:      1,
	})
	_, err := k.StateInfo(wctx, request)
	errIndex := 1 + (60-1)/2 // Using binary search, the middle index is lookedup first and is missing.
	require.EqualError(t, err, errorsmod.Wrapf(types.ErrNotFound,
		"StateInfo wasn't found for rollappId=%s, index=%d",
		rollappId, errIndex).Error())
}

func TestStateInfoByHeightErr(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	rollappID := urand.RollappID()
	msgs := createNStateInfoAndIndex(k, ctx, 4, rollappID)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetStateInfoRequest
		response *types.QueryGetStateInfoResponse
		err      error
	}{
		{
			desc: "StateInfoByHeight",
			request: &types.QueryGetStateInfoRequest{
				RollappId: rollappID,
				Height:    msgs[3].StartHeight + 1,
			},
			response: &types.QueryGetStateInfoResponse{StateInfo: types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{RollappId: rollappID, Index: 4},
				StartHeight:    msgs[3].StartHeight,
				NumBlocks:      msgs[3].NumBlocks,
			}},
		},
		{
			desc: "StateInfoByHeight_firstBlockInBatch",
			request: &types.QueryGetStateInfoRequest{
				RollappId: rollappID,
				Height:    msgs[2].StartHeight,
			},
			response: &types.QueryGetStateInfoResponse{StateInfo: types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{RollappId: rollappID, Index: 3},
				StartHeight:    msgs[2].StartHeight,
				NumBlocks:      msgs[2].NumBlocks,
			}},
		},
		{
			desc: "StateInfoByHeight_lastBlockInBatch",
			request: &types.QueryGetStateInfoRequest{
				RollappId: rollappID,
				Height:    msgs[2].StartHeight + msgs[2].NumBlocks - 1,
			},
			response: &types.QueryGetStateInfoResponse{StateInfo: types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{RollappId: rollappID, Index: 3},
				StartHeight:    msgs[2].StartHeight,
				NumBlocks:      msgs[2].NumBlocks,
			}},
		},
		{
			desc: "StateInfoByHeight_unknownRollappId",
			request: &types.QueryGetStateInfoRequest{
				RollappId: "UnknownRollappId",
				Height:    5,
			},
			err: types.ErrUnknownRollappID,
		},
		{
			desc: "StateInfoByHeight_invalidHeight",
			request: &types.QueryGetStateInfoRequest{
				RollappId: rollappID,
				Height:    10000000,
			},
			err: types.ErrStateNotExists,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := k.StateInfo(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestStateInfoByHeightValidIncreasingBlockBatches(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	numOfMsg := 20
	rollappID := urand.RollappID()
	msgs := createNStateInfoAndIndex(k, ctx, numOfMsg, rollappID)

	for i := 0; i < numOfMsg; i += 1 {
		for height := msgs[i].StartHeight; height < msgs[i].StartHeight+msgs[i].NumBlocks; height += 1 {
			request := &types.QueryGetStateInfoRequest{
				RollappId: rollappID,
				Height:    height,
			}
			response, err := k.StateInfo(wctx, request)
			require.NoError(t, err)
			require.Equal(t,
				nullify.Fill(&types.QueryGetStateInfoResponse{StateInfo: msgs[i]}),
				nullify.Fill(response),
			)
		}
	}
}

func TestStateInfoByHeightValidDecreasingBlockBatches(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	numOfMsg := 20
	rollappID := urand.RollappID()
	msgs := createNStateInfoAndIndex(k, ctx, numOfMsg, rollappID)

	for i := 0; i < numOfMsg; i += 1 {
		for height := msgs[i].StartHeight; height < msgs[i].StartHeight+msgs[i].NumBlocks; height += 1 {
			request := &types.QueryGetStateInfoRequest{
				RollappId: rollappID,
				Height:    height,
			}
			response, err := k.StateInfo(wctx, request)
			require.NoError(t, err)
			require.Equal(t,
				nullify.Fill(&types.QueryGetStateInfoResponse{StateInfo: msgs[i]}),
				nullify.Fill(response),
			)
		}
	}
}

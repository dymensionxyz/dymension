package keeper_test

import (
	"math/rand"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
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
			NumBlocks:   numBlocks,
		}
		StartHeight += stateInfo.NumBlocks

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
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	rollappId := "rollappId"
	keeper.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappId,
	})
	request := &types.QueryGetStateInfoRequest{
		RollappId: rollappId,
		Height:    100,
	}
	_, err := keeper.StateInfo(wctx, request)
	require.EqualError(t, err, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "LatestStateInfoIndex wasn't found for rollappId=%s", rollappId).Error())
}

func TestStateInfoByHeightMissingStateInfo(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	rollappId := "rollappId"
	keeper.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappId,
	})
	keeper.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: rollappId,
		Index:     uint64(85),
	})
	request := &types.QueryGetStateInfoRequest{
		RollappId: rollappId,
		Height:    100,
	}
	_, err := keeper.StateInfo(wctx, request)
	require.EqualError(t, err, sdkerrors.Wrapf(sdkerrors.ErrNotFound,
		"StateInfo wasn't found for rollappId=%s, index=%d",
		rollappId, 85).Error())
}

func TestStateInfoByHeightMissingStateInfo1(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	rollappId := "rollappId"
	keeper.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappId,
	})
	keeper.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: rollappId,
		Index:     uint64(60),
	})
	request := &types.QueryGetStateInfoRequest{
		RollappId: rollappId,
		Height:    70,
	}
	keeper.SetStateInfo(ctx, types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: rollappId, Index: 60},
		StartHeight:    71,
		NumBlocks:      1,
	})
	_, err := keeper.StateInfo(wctx, request)
	require.EqualError(t, err, sdkerrors.Wrapf(sdkerrors.ErrNotFound,
		"StateInfo wasn't found for rollappId=%s, index=%d",
		rollappId, 1).Error())
}

func TestStateInfoByHeightErr(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNStateInfoAndIndex(keeper, ctx, 4, "rollappId")
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetStateInfoRequest
		response *types.QueryGetStateInfoResponse
		err      error
	}{
		{
			desc: "StateInfoByHeight",
			request: &types.QueryGetStateInfoRequest{
				RollappId: "rollappId",
				Height:    msgs[3].StartHeight + 1,
			},
			response: &types.QueryGetStateInfoResponse{StateInfo: types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{RollappId: "rollappId", Index: 4},
				StartHeight:    msgs[3].StartHeight,
				NumBlocks:      msgs[3].NumBlocks,
			}},
		},
		{
			desc: "StateInfoByHeight_firstBlockInBatch",
			request: &types.QueryGetStateInfoRequest{
				RollappId: "rollappId",
				Height:    msgs[2].StartHeight,
			},
			response: &types.QueryGetStateInfoResponse{StateInfo: types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{RollappId: "rollappId", Index: 3},
				StartHeight:    msgs[2].StartHeight,
				NumBlocks:      msgs[2].NumBlocks,
			}},
		},
		{
			desc: "StateInfoByHeight_lastBlockInBatch",
			request: &types.QueryGetStateInfoRequest{
				RollappId: "rollappId",
				Height:    msgs[2].StartHeight + msgs[2].NumBlocks - 1,
			},
			response: &types.QueryGetStateInfoResponse{StateInfo: types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{RollappId: "rollappId", Index: 3},
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
				RollappId: "rollappId",
				Height:    10000000,
			},
			err: types.ErrStateNotExists,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.StateInfo(wctx, tc.request)
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
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	numOfMsg := 20
	msgs := createNStateInfoAndIndex(keeper, ctx, numOfMsg, "rollappId")

	for i := 0; i < numOfMsg; i += 1 {
		for height := msgs[i].StartHeight; height < msgs[i].StartHeight+msgs[i].NumBlocks; height += 1 {
			request := &types.QueryGetStateInfoRequest{
				RollappId: "rollappId",
				Height:    height,
			}
			response, err := keeper.StateInfo(wctx, request)
			require.NoError(t, err)
			require.Equal(t,
				nullify.Fill(&types.QueryGetStateInfoResponse{StateInfo: msgs[i]}),
				nullify.Fill(response),
			)
		}
	}
}

func TestStateInfoByHeightValidDecreasingBlockBatches(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	numOfMsg := 20
	msgs := createNStateInfoAndIndex(keeper, ctx, numOfMsg, "rollappId")

	for i := 0; i < numOfMsg; i += 1 {
		for height := msgs[i].StartHeight; height < msgs[i].StartHeight+msgs[i].NumBlocks; height += 1 {
			request := &types.QueryGetStateInfoRequest{
				RollappId: "rollappId",
				Height:    height,
			}
			response, err := keeper.StateInfo(wctx, request)
			require.NoError(t, err)
			require.Equal(t,
				nullify.Fill(&types.QueryGetStateInfoResponse{StateInfo: msgs[i]}),
				nullify.Fill(response),
			)
		}
	}
}

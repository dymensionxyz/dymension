package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestStateInfoQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs, _ := createNStateInfo(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetStateInfoRequest
		response *types.QueryGetStateInfoResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetStateInfoRequest{
				RollappId: msgs[0].StateInfoIndex.RollappId,
				Index:     msgs[0].StateInfoIndex.Index,
			},
			response: &types.QueryGetStateInfoResponse{StateInfo: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetStateInfoRequest{
				RollappId: msgs[1].StateInfoIndex.RollappId,
				Index:     msgs[1].StateInfoIndex.Index,
			},
			response: &types.QueryGetStateInfoResponse{StateInfo: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetStateInfoRequest{
				RollappId: strconv.Itoa(100000),
				Index:     msgs[0].StateInfoIndex.Index,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetStateInfoRequest{
				RollappId: msgs[1].StateInfoIndex.RollappId,
				Index:     100000,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
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

func TestFindStateInfoByHeight(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	rollappID := urand.RollappID()
	keeper.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappID,
	})
	keeper.SetStateInfo(ctx, types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: rollappID, Index: 1},
		StartHeight:    1,
	}.WithNumBlocks(2))
	keeper.SetStateInfo(ctx, types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: rollappID, Index: 2},
		StartHeight:    3,
	}.WithNumBlocks(3))
	keeper.SetStateInfo(ctx, types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: rollappID, Index: 3},
		StartHeight:    6,
	}.WithNumBlocks(4))
	keeper.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: rollappID,
		Index:     3,
	})

	type testInput struct {
		rollappId string
		height    uint64
	}

	testCase := []struct {
		name           string
		input          testInput
		stateInfoIndex uint64
		err            error
	}{
		{
			name: "Zero height",
			input: testInput{
				rollappId: "1",
				height:    0,
			},
			err: types.ErrInvalidHeight,
		},
		{
			name: "Rollapp not found",
			input: testInput{
				rollappId: "unknown",
				height:    1,
			},
			err: types.ErrUnknownRollappID,
		},
		{
			name: "First height",
			input: testInput{
				rollappId: rollappID,
				height:    1,
			},
			stateInfoIndex: 1,
		},
		{
			name: "Last height",
			input: testInput{
				rollappId: rollappID,
				height:    9,
			},
			stateInfoIndex: 3,
		},
		{
			name: "Height in between",
			input: testInput{
				rollappId: rollappID,
				height:    4,
			},
			stateInfoIndex: 2,
		},
		{
			name: "Height not found",
			input: testInput{
				rollappId: rollappID,
				height:    10,
			},
			err: types.ErrStateNotExists,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			response, err := keeper.FindStateInfoByHeight(ctx, tc.input.rollappId, tc.input.height)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.stateInfoIndex, response.StateInfoIndex.Index)
			}
		})
	}
}

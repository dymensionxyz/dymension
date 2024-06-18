package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pgregory.net/rapid"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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

func TestFindStateInfoByHeightBinary(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)

	f := func(r *rapid.T) {
		heights := rapid.IntRange(1, 100)

		rollapp := "foo"
		lastHeight := uint64(0)
		ix := uint64(1)

		ops := map[string]func(*rapid.T){
			"insert": func(t *rapid.T) {
				height := uint64(heights.Draw(t, "k"))
				if height <= lastHeight {
					return
				}
				info := types.StateInfo{
					StateInfoIndex: types.StateInfoIndex{
						RollappId: rollapp,
						Index:     ix,
					},
					StartHeight: lastHeight + 1,
					NumBlocks:   height - lastHeight,
				}
				keeper.SetStateInfo(ctx, info)
				keeper.SetLatestStateInfoIndex(ctx, info.StateInfoIndex)
				lastHeight = height
				ix++
			},
			"find": func(t *rapid.T) {
				height := uint64(heights.Draw(t, "k"))
				state, err := keeper.FindStateInfoByHeightBinary(ctx, rollapp, height)
				shouldFind := 0 < lastHeight && height <= lastHeight
				if shouldFind && err != nil {
					t.Fatalf("err: %v", err)
				}
				if shouldFind && !state.ContainsHeight(height) {
					t.Fatal("does not contain height")
				}
			},
		}
		r.Repeat(ops)
	}

	rapid.Check(t, f)
}

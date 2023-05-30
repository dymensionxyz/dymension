package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
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

func TestStateInfoQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	_, msgs := createNStateInfo(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllStateInfoRequest {
		return &types.QueryAllStateInfoRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.StateInfoAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.StateInfo), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.StateInfo),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.StateInfoAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.StateInfo), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.StateInfo),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.StateInfoAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.StateInfoAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

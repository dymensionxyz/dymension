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

func TestBlockHeightToFinalizationQueueQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNBlockHeightToFinalizationQueue(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetBlockHeightToFinalizationQueueRequest
		response *types.QueryGetBlockHeightToFinalizationQueueResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetBlockHeightToFinalizationQueueRequest{
				FinalizationHeight: msgs[0].FinalizationHeight,
			},
			response: &types.QueryGetBlockHeightToFinalizationQueueResponse{BlockHeightToFinalizationQueue: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetBlockHeightToFinalizationQueueRequest{
				FinalizationHeight: msgs[1].FinalizationHeight,
			},
			response: &types.QueryGetBlockHeightToFinalizationQueueResponse{BlockHeightToFinalizationQueue: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetBlockHeightToFinalizationQueueRequest{
				FinalizationHeight: 100000,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.BlockHeightToFinalizationQueue(wctx, tc.request)
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

func TestBlockHeightToFinalizationQueueQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNBlockHeightToFinalizationQueue(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllBlockHeightToFinalizationQueueRequest {
		return &types.QueryAllBlockHeightToFinalizationQueueRequest{
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
			resp, err := keeper.BlockHeightToFinalizationQueueAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.BlockHeightToFinalizationQueue), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.BlockHeightToFinalizationQueue),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.BlockHeightToFinalizationQueueAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.BlockHeightToFinalizationQueue), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.BlockHeightToFinalizationQueue),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.BlockHeightToFinalizationQueueAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.BlockHeightToFinalizationQueue),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.BlockHeightToFinalizationQueueAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

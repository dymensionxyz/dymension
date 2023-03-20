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
	"github.com/dymensionxyz/dymension/x/irc/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestIRCRequestQuerySingle(t *testing.T) {
	keeper, _, ctx := keepertest.IRCKeeper(t)

	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNIRCRequest(t, keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetIRCRequestRequest
		response *types.QueryGetIRCRequestResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetIRCRequestRequest{
				ReqId: msgs[0].ReqId,
			},
			response: &types.QueryGetIRCRequestResponse{IrcRequest: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetIRCRequestRequest{
				ReqId: msgs[1].ReqId,
			},
			response: &types.QueryGetIRCRequestResponse{IrcRequest: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetIRCRequestRequest{
				ReqId: 100000,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.IRCRequest(wctx, tc.request)
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

func TestIRCRequestQueryPaginated(t *testing.T) {
	keeper, _, ctx := keepertest.IRCKeeper(t)

	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNIRCRequest(t, keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllIRCRequestRequest {
		return &types.QueryAllIRCRequestRequest{
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
			resp, err := keeper.IRCRequestAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.IrcRequest), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.IrcRequest),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.IRCRequestAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.IrcRequest), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.IrcRequest),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.IRCRequestAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.IrcRequest),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.IRCRequestAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

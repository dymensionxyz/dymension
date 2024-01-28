package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestSchedulerQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNScheduler(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetSchedulerRequest
		response *types.QueryGetSchedulerResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetSchedulerRequest{
				SequencerAddress: msgs[0].SequencerAddress,
			},
			response: &types.QueryGetSchedulerResponse{Scheduler: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetSchedulerRequest{
				SequencerAddress: msgs[1].SequencerAddress,
			},
			response: &types.QueryGetSchedulerResponse{Scheduler: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetSchedulerRequest{
				SequencerAddress: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Scheduler(wctx, tc.request)
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

func TestSchedulerQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNScheduler(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllSchedulerRequest {
		return &types.QueryAllSchedulerRequest{
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
			resp, err := keeper.SchedulerAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Scheduler), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Scheduler),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.SchedulerAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Scheduler), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Scheduler),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.SchedulerAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Scheduler),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.SchedulerAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

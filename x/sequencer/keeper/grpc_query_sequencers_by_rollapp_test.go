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
	"github.com/dymensionxyz/dymension/x/sequencer/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestSequencersByRollappQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	sequencersByRollappList := createNSequencersByRollapp(keeper, ctx, 2)
	var SequencersByRollappResponseList []types.QueryGetSequencersByRollappResponse
	for _, sequencerByRollapp := range sequencersByRollappList {
		var sequencerInfoList []types.SequencerInfo
		for _, sequencerAddr := range sequencerByRollapp.Sequencers.Addresses {
			sequencer, found := keeper.GetSequencer(ctx, sequencerAddr)
			require.True(t, found)
			sequencerInfoList = append(sequencerInfoList, types.SequencerInfo{
				Sequencer: sequencer,
				Status:    types.Unspecified,
			})
		}
		SequencersByRollappResponseList = append(SequencersByRollappResponseList,
			types.QueryGetSequencersByRollappResponse{RollappId: sequencerByRollapp.RollappId, SequencerInfoList: sequencerInfoList})
	}
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetSequencersByRollappRequest
		response *types.QueryGetSequencersByRollappResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetSequencersByRollappRequest{
				RollappId: sequencersByRollappList[0].RollappId,
			},
			response: &SequencersByRollappResponseList[0],
		},
		{
			desc: "Second",
			request: &types.QueryGetSequencersByRollappRequest{
				RollappId: sequencersByRollappList[1].RollappId,
			},
			response: &SequencersByRollappResponseList[1],
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetSequencersByRollappRequest{
				RollappId: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.SequencersByRollapp(wctx, tc.request)
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

func TestSequencersByRollappQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	sequencersByRollappList := createNSequencersByRollapp(keeper, ctx, 5)
	var SequencersByRollappResponseList []types.QueryGetSequencersByRollappResponse
	for _, sequencerByRollapp := range sequencersByRollappList {
		var sequencerInfoList []types.SequencerInfo
		for _, sequencerAddr := range sequencerByRollapp.Sequencers.Addresses {
			sequencer, found := keeper.GetSequencer(ctx, sequencerAddr)
			require.True(t, found)
			sequencerInfoList = append(sequencerInfoList, types.SequencerInfo{
				Sequencer: sequencer,
				Status:    types.Unspecified,
			})
		}
		SequencersByRollappResponseList = append(SequencersByRollappResponseList,
			types.QueryGetSequencersByRollappResponse{RollappId: sequencerByRollapp.RollappId, SequencerInfoList: sequencerInfoList})
	}
	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllSequencersByRollappRequest {
		return &types.QueryAllSequencersByRollappRequest{
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
		for i := 0; i < len(sequencersByRollappList); i += step {
			resp, err := keeper.SequencersByRollappAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SequencersByRollapp), step)
			require.Subset(t,
				nullify.Fill(SequencersByRollappResponseList),
				nullify.Fill(resp.SequencersByRollapp),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(sequencersByRollappList); i += step {
			resp, err := keeper.SequencersByRollappAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SequencersByRollapp), step)
			require.Subset(t,
				nullify.Fill(SequencersByRollappResponseList),
				nullify.Fill(resp.SequencersByRollapp),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.SequencersByRollappAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(sequencersByRollappList), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(SequencersByRollappResponseList),
			nullify.Fill(resp.SequencersByRollapp),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.SequencersByRollappAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestLatestStateIndexQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNLatestStateInfoIndex(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetLatestStateIndexRequest
		response *types.QueryGetLatestStateIndexResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetLatestStateIndexRequest{
				RollappId: msgs[0].RollappId,
			},
			response: &types.QueryGetLatestStateIndexResponse{StateIndex: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetLatestStateIndexRequest{
				RollappId: msgs[1].RollappId,
			},
			response: &types.QueryGetLatestStateIndexResponse{StateIndex: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetLatestStateIndexRequest{
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
			response, err := keeper.LatestStateIndex(wctx, tc.request)
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

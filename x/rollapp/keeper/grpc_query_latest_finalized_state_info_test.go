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

func TestLatestFinalizedStateInfoQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNLatestFinalizedStateIndex(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetLatestFinalizedStateInfoRequest
		response *types.QueryGetLatestFinalizedStateInfoResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetLatestFinalizedStateInfoRequest{
				RollappId: msgs[0].RollappId,
			},
			response: &types.QueryGetLatestFinalizedStateInfoResponse{
				StateInfo: types.StateInfo{
					StateInfoIndex: msgs[0],
					Status:         types.STATE_STATUS_FINALIZED,
				},
			},
		},
		{
			desc: "Second",
			request: &types.QueryGetLatestFinalizedStateInfoRequest{
				RollappId: msgs[1].RollappId,
			},
			response: &types.QueryGetLatestFinalizedStateInfoResponse{StateInfo: types.StateInfo{
				StateInfoIndex: msgs[1],
				Status:         types.STATE_STATUS_FINALIZED,
			}},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetLatestFinalizedStateInfoRequest{
				RollappId: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "LatestFinalizedStateIndex not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.LatestFinalizedStateInfo(wctx, tc.request)
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

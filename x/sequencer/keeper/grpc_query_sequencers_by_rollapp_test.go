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
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestSequencersByRollappQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	sequencersByRollappList := createNSequencer(keeper, ctx, 2)
	var SequencersByRollappResponseList []types.QueryGetSequencersByRollappResponse

	for _, sequencerByRollapp := range sequencersByRollappList {
		sequencer, found := keeper.GetSequencer(ctx, sequencerByRollapp.SequencerAddress)
		require.True(t, found)
		SequencersByRollappResponseList = append(SequencersByRollappResponseList,
			types.QueryGetSequencersByRollappResponse{
				SequencerInfoList: []types.Sequencer{sequencer},
			})
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

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

// go test -run=TestPropertyBased -rapid.checks=10000 -rapid.steps=50
func TestPropertyBased(t *testing.T) {
	/*
	  -rapid.checks int
	    	rapid: number of checks to perform (default 100)
	  -rapid.debug
	    	rapid: debugging output
	  -rapid.debugvis
	    	rapid: debugging visualization
	  -rapid.failfile string
	    	rapid: fail file to use to reproduce test failure
	  -rapid.log
	    	rapid: eager verbose output to stdout (to aid with unrecoverable test failures)
	  -rapid.nofailfile
	    	rapid: do not write fail files on test failures
	  -rapid.seed uint
	    	rapid: PRNG seed to start with (0 to use a random one)
	  -rapid.shrinktime duration
	    	rapid: maximum time to spend on test case minimization (default 30s)
	  -rapid.steps int
	    	rapid: average number of Repeat actions to execute (default 30)
	  -rapid.v
	    	rapid: verbose output
	*/
	rapid.Check(t, testWithRapid)
}

func testWithRapid(t *rapid.T) {
	key := rapid.IntRange(-12, 12)

	set := map[int]struct{}{}

	ops := map[string]func(*rapid.T){
		"insert": func(t *rapid.T) {
			k := key.Draw(t, "k")
			set[k] = struct{}{}
		},
		"find": func(t *rapid.T) {
			k := key.Draw(t, "k")
			_, ok := set[k]
		},
	}
	t.Repeat(ops)
}

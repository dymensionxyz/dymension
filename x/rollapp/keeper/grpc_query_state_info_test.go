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

// go test ./x/rollapp/keeper/... -run=TestFindStateInfoByHeightBinary -rapid.checks=10000 -rapid.steps=50
func TestFindStateInfoByHeightBinary(t *testing.T) {
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
	keeper, ctx := keepertest.RollappKeeper(t)

	f := func(r *rapid.T) {
		heights := rapid.IntRange(1, 100)

		rollapp := "foo"
		lastHeight := uint64(0)
		ix := uint64(1)

		ops := map[string]func(*rapid.T){
			"insert": func(t *rapid.T) {
				height := uint64(heights.Draw(t, "k"))
				t.Logf("inserting: [%d, %d]", lastHeight+1, height)
				if height <= lastHeight {
					return
				}
				ixKey := types.StateInfoIndex{
					RollappId: rollapp,
					Index:     ix,
				}
				keeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: ixKey,
					Sequencer:      "",
					StartHeight:    lastHeight + 1,
					NumBlocks:      height - lastHeight,
					DAPath:         "",
					Version:        0,
					CreationHeight: 0,
					Status:         0,
					BDs:            types.BlockDescriptors{},
				})
				keeper.SetLatestStateInfoIndex(ctx, ixKey)
				lastHeight = height
				ix++
			},
			"sanity": func(t *rapid.T) {
				if 1 < ix {
					_, ok := keeper.GetStateInfo(ctx, rollapp, ix-1)
					require.True(t, ok)
				}
			},
			"find": func(t *rapid.T) {
				height := uint64(heights.Draw(t, "k"))
				t.Logf("searching for: %d", height)
				state, err := keeper.FindStateInfoByHeightBinary(ctx, rollapp, height)
				shouldFind := 1 < ix && height <= lastHeight
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

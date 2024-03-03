package keeper_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestSequencersByRollappQuery3() {
	suite.SetupTest()

	rollappId := suite.CreateDefaultRollapp()
	rollappId2 := suite.CreateDefaultRollapp()

	// create 2 sequencer
	addr1_1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	addr2_1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	seq1, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1_1)
	require.True(suite.T(), found)
	seq2, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2_1)
	require.True(suite.T(), found)
	seq1Response := types.QueryGetSequencersByRollappResponse{
		Sequencers: []types.Sequencer{seq1, seq2},
	}

	addr1_2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId2)
	addr2_2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId2)
	seq3, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1_2)
	require.True(suite.T(), found)
	seq4, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2_2)
	require.True(suite.T(), found)
	seq2Response := types.QueryGetSequencersByRollappResponse{
		Sequencers: []types.Sequencer{seq3, seq4},
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
				RollappId: rollappId,
			},
			response: &seq1Response,
		},
		{
			desc: "Second",
			request: &types.QueryGetSequencersByRollappRequest{
				RollappId: rollappId2,
			},
			response: &seq2Response,
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetSequencersByRollappRequest{
				RollappId: strconv.Itoa(100000),
			},
			err: types.ErrUnknownRollappID,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		suite.T().Run(tc.desc, func(t *testing.T) {
			response, err := suite.App.SequencerKeeper.SequencersByRollapp(suite.Ctx, tc.request)
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

func (suite *SequencerTestSuite) TestSequencersByRollappByStatusQuery() {
	suite.SetupTest()

	msgserver := keeper.NewMsgServerImpl(suite.App.SequencerKeeper)

	rollappId := suite.CreateDefaultRollapp()
	// create 2 sequencers on rollapp1
	addr1_1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	addr2_1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	_, err := msgserver.Unbond(suite.Ctx, &types.MsgUnbond{
		Creator: addr2_1,
	})
	require.NoError(suite.T(), err)

	// create 2 sequencers on rollapp2
	rollappId2 := suite.CreateDefaultRollapp()
	addr1_2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId2)
	addr2_2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId2)

	for _, tc := range []struct {
		desc          string
		request       *types.QueryGetSequencersByRollappByStatusRequest
		response_addr []string
		err           error
	}{
		{
			desc: "First - Bonded",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: rollappId,
				Status:    types.Bonded,
			},
			response_addr: []string{addr1_1},
		},
		{
			desc: "First - Unbonding",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: rollappId,
				Status:    types.Unbonding,
			},
			response_addr: []string{addr2_1},
		},
		{
			desc: "First - Unbonded",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: rollappId,
				Status:    types.Unbonded,
			},
			response_addr: []string{},
		},
		{
			desc: "Second",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: rollappId2,
				Status:    types.Bonded,
			},
			response_addr: []string{addr1_2, addr2_2},
		},
		{
			desc: "Second - prposer and bonded the same",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: rollappId2,
				Status:    types.Proposer,
			},
			response_addr: []string{addr1_2, addr2_2},
		},
		{
			desc: "Unspecified Status",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: rollappId2,
				Status:    types.Unspecified,
			},
			response_addr: []string{addr1_2, addr2_2},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: strconv.Itoa(100000),
			},
			err: types.ErrUnknownRollappID,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		suite.T().Run(tc.desc, func(t *testing.T) {
			response, err := suite.App.SequencerKeeper.SequencersByRollappByStatus(suite.Ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Len(t, response.Sequencers, len(tc.response_addr))

				for _, seqAddr := range tc.response_addr {
					seq, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, seqAddr)
					require.True(t, found)
					require.Contains(t, response.Sequencers, seq)
				}
			}
		})
	}
}

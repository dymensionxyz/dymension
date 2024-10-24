package keeper_test

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (s *SequencerTestSuite) TestSequencersByRollappQuery3() {
	rollappId, pk11 := s.CreateDefaultRollapp()
	pk12 := ed25519.GenPrivKey().PubKey()
	rollappId2, pk21 := s.CreateDefaultRollapp()
	pk22 := ed25519.GenPrivKey().PubKey()

	// create 2 sequencer
	addr11 := s.CreateSequencer(s.Ctx, rollappId, pk11)
	addr21 := s.CreateSequencer(s.Ctx, rollappId, pk12)
	seq1, found := s.App.SequencerKeeper.GetSequencer(s.Ctx, addr11)
	require.True(s.T(), found)
	seq2, found := s.App.SequencerKeeper.GetSequencer(s.Ctx, addr21)
	require.True(s.T(), found)
	seq1Response := types.QueryGetSequencersByRollappResponse{
		Sequencers: []types.Sequencer{seq1, seq2},
	}

	addr12 := s.CreateSequencer(s.Ctx, rollappId2, pk21)
	addr22 := s.CreateSequencer(s.Ctx, rollappId2, pk22)
	seq3, found := s.App.SequencerKeeper.GetSequencer(s.Ctx, addr12)
	require.True(s.T(), found)
	seq4, found := s.App.SequencerKeeper.GetSequencer(s.Ctx, addr22)
	require.True(s.T(), found)
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
			err: types.ErrRollappNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  gerrc.ErrInvalidArgument,
		},
	} {
		s.T().Run(tc.desc, func(t *testing.T) {
			response, err := s.App.SequencerKeeper.SequencersByRollapp(s.Ctx, tc.request)
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

func (s *SequencerTestSuite) TestSequencersByRollappByStatusQuery() {
	msgserver := keeper.NewMsgServerImpl(s.App.SequencerKeeper)

	rollappId, pk11 := s.CreateDefaultRollapp()
	pk12 := ed25519.GenPrivKey().PubKey()
	// create 2 sequencers on rollapp1
	addr11 := s.CreateSequencer(s.Ctx, rollappId, pk11)
	addr21 := s.CreateSequencer(s.Ctx, rollappId, pk12)
	_, err := msgserver.Unbond(s.Ctx, &types.MsgUnbond{
		Creator: addr21,
	})
	require.NoError(s.T(), err)

	// create 2 sequencers on rollapp2
	rollappId2, pk21 := s.CreateDefaultRollapp()
	pk22 := ed25519.GenPrivKey().PubKey()
	addr12 := s.CreateSequencer(s.Ctx, rollappId2, pk21)
	addr22 := s.CreateSequencer(s.Ctx, rollappId2, pk22)

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
			response_addr: []string{addr11},
		},
		{
			desc: "First - Unbonding",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: rollappId,
				Status:    types.Unbonding,
			},
			response_addr: []string{addr21},
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
			response_addr: []string{addr12, addr22},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: strconv.Itoa(100000),
			},
			err: types.ErrRollappNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  gerrc.ErrInvalidArgument,
		},
	} {
		s.T().Run(tc.desc, func(t *testing.T) {
			response, err := s.App.SequencerKeeper.SequencersByRollappByStatus(s.Ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Len(t, response.Sequencers, len(tc.response_addr))

				for _, seqAddr := range tc.response_addr {
					seq, found := s.App.SequencerKeeper.GetSequencer(s.Ctx, seqAddr)
					require.True(t, found)
					require.Contains(t, response.Sequencers, seq)
				}
			}
		})
	}
}

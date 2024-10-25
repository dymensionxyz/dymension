package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (s *SequencerTestSuite) TestSequencersByRollappQuery() {
	rollappId, pk11 := s.createRollappWithInitialSequencer()
	pk12 := ed25519.GenPrivKey().PubKey()
	rollappId2, pk21 := s.createRollappWithInitialSequencer()
	pk22 := ed25519.GenPrivKey().PubKey()

	// create 2 sequencer
	addr11 := s.createSequencerWithPk(s.Ctx, rollappId, pk11)
	addr21 := s.createSequencerWithPk(s.Ctx, rollappId, pk12)
	seq1, err := s.k().GetRealSequencer(s.Ctx, addr11)
	require.NoError(s.T(), err)
	seq2, err := s.k().GetRealSequencer(s.Ctx, addr21)
	require.NoError(s.T(), err)
	seq1Response := types.QueryGetSequencersByRollappResponse{
		Sequencers: []types.Sequencer{seq1, seq2},
	}

	addr12 := s.createSequencerWithPk(s.Ctx, rollappId2, pk21)
	addr22 := s.createSequencerWithPk(s.Ctx, rollappId2, pk22)
	seq3, err := s.k().GetRealSequencer(s.Ctx, addr12)
	require.NoError(s.T(), err)
	seq4, err := s.k().GetRealSequencer(s.Ctx, addr22)
	require.NoError(s.T(), err)
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
			desc: "InvalidRequest",
			err:  gerrc.ErrInvalidArgument,
		},
	} {
		s.T().Run(tc.desc, func(t *testing.T) {
			response, err := s.k().SequencersByRollapp(s.Ctx, tc.request)
			if tc.err != nil {
				utest.IsErr(require.New(t), err, tc.err)
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
	rollappId, pk11 := s.createRollappWithInitialSequencer()
	pk12 := ed25519.GenPrivKey().PubKey()
	// create 2 sequencers on rollapp1
	addr11 := s.createSequencerWithPk(s.Ctx, rollappId, pk11)
	addr21 := s.createSequencerWithPk(s.Ctx, rollappId, pk12)
	_, err := s.msgServer.Unbond(s.Ctx, &types.MsgUnbond{
		Creator: addr21,
	})
	require.NoError(s.T(), err)

	// create 2 sequencers on rollapp2
	rollappId2, pk21 := s.createRollappWithInitialSequencer()
	pk22 := ed25519.GenPrivKey().PubKey()
	addr12 := s.createSequencerWithPk(s.Ctx, rollappId2, pk21)
	addr22 := s.createSequencerWithPk(s.Ctx, rollappId2, pk22)

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
			desc: "First - Unbonded",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: rollappId,
				Status:    types.Unbonded,
			},
			response_addr: []string{addr21},
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
			desc: "InvalidRequest",
			err:  gerrc.ErrInvalidArgument,
		},
	} {
		s.T().Run(tc.desc, func(t *testing.T) {
			response, err := s.k().SequencersByRollappByStatus(s.Ctx, tc.request)
			if tc.err != nil {
				utest.IsErr(require.New(t), err, tc.err)
			} else {
				require.NoError(t, err)
				require.Len(t, response.Sequencers, len(tc.response_addr))

				for _, seqAddr := range tc.response_addr {
					seq, err := s.k().GetRealSequencer(s.Ctx, seqAddr)
					require.NoError(t, err)
					require.Contains(t, response.Sequencers, seq)
				}
			}
		})
	}
}

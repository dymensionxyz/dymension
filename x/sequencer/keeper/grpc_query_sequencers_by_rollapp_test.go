package keeper_test

import (
	"testing"

	"github.com/dymensionxyz/sdk-utils/utils/utest"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (s *SequencerTestSuite) TestSequencersByRollappQuery() {
	ra1 := s.createRollapp()
	ra2 := s.createRollapp()
	pk11 := pks[0]
	pk12 := pks[1]
	pk21 := pks[2]
	pk22 := pks[3]
	seq1 := s.createSequencerWithBond(s.Ctx, ra1.RollappId, pk11, bond)
	seq2 := s.createSequencerWithBond(s.Ctx, ra1.RollappId, pk12, bond)
	seq3 := s.createSequencerWithBond(s.Ctx, ra2.RollappId, pk21, bond)
	seq4 := s.createSequencerWithBond(s.Ctx, ra2.RollappId, pk22, bond)

	seq1Response := types.QueryGetSequencersByRollappResponse{
		Sequencers: []types.Sequencer{seq1, seq2},
	}

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
				RollappId: ra1.RollappId,
			},
			response: &seq1Response,
		},
		{
			desc: "Second",
			request: &types.QueryGetSequencersByRollappRequest{
				RollappId: ra2.RollappId,
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
	ra1 := s.createRollapp()
	ra2 := s.createRollapp()
	pk11 := pks[0]
	pk12 := pks[1]
	pk21 := pks[2]
	pk22 := pks[3]
	addr11 := s.createSequencerWithBond(s.Ctx, ra1.RollappId, pk11, bond).Address
	addr12 := s.createSequencerWithBond(s.Ctx, ra1.RollappId, pk12, bond).Address
	addr21 := s.createSequencerWithBond(s.Ctx, ra2.RollappId, pk21, bond).Address
	addr22 := s.createSequencerWithBond(s.Ctx, ra2.RollappId, pk22, bond).Address

	_, err := s.msgServer.Unbond(s.Ctx, &types.MsgUnbond{Creator: addr12})
	s.Require().NoError(err)

	for _, tc := range []struct {
		desc         string
		request      *types.QueryGetSequencersByRollappByStatusRequest
		responseAddr []string
		err          error
	}{
		{
			desc: "First - Bonded",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: ra1.RollappId,
				Status:    types.Bonded,
			},
			responseAddr: []string{addr11},
		},
		{
			desc: "First - Unbonded",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: ra1.RollappId,
				Status:    types.Unbonded,
			},
			responseAddr: []string{addr12},
		},
		{
			desc: "Second",
			request: &types.QueryGetSequencersByRollappByStatusRequest{
				RollappId: ra2.RollappId,
				Status:    types.Bonded,
			},
			responseAddr: []string{addr21, addr22},
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
				require.Len(t, response.Sequencers, len(tc.responseAddr))

				for _, seqAddr := range tc.responseAddr {
					seq, err := s.k().RealSequencer(s.Ctx, seqAddr)
					require.NoError(t, err)
					require.Contains(t, response.Sequencers, seq)
				}
			}
		})
	}
}

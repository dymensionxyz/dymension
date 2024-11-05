package keeper_test

import (
	"testing"

	_ "github.com/cosmos/cosmos-sdk/crypto/codec"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/stretchr/testify/require"
)

func TestSequencerGet(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencers(k, ctx, 10)
	for _, item := range items {
		item := item
		rst, err := k.RealSequencer(ctx,
			item.Address,
		)
		require.NoError(t, err)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestSequencerGetAll(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencers(k, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(k.AllSequencers(ctx)),
	)
}

func TestSequencersByRollappGet(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencers(k, ctx, 10)
	rst := k.RollappSequencers(ctx,
		items[0].RollappId,
	)

	require.Equal(t, len(rst), len(items))
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(rst),
	)
}

func (s *SequencerTestSuite) TestByProposerAddr() {
	ra := s.createRollapp()
	seqExp := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	seqGot, err := s.k().SequencerByDymintAddr(s.Ctx, seqExp.MustProposerAddr())
	s.Require().NoError(err)
	s.Require().Equal(seqExp.Address, seqGot.Address)
}

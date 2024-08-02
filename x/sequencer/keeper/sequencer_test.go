package keeper_test

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNSequencer(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Sequencer {
	items := make([]types.Sequencer, n)
	for i := range items {
		seq := types.Sequencer{
			Address: strconv.Itoa(i),
			Status:  types.Bonded,
		}
		items[i] = seq

		keeper.SetSequencer(ctx, items[i])
	}
	return items
}

func TestSequencerGet(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(keeper, ctx, 10)
	for _, item := range items {
		item := item
		rst, found := keeper.GetSequencer(ctx,
			item.Address,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestSequencerGetAll(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(k, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(k.GetAllSequencers(ctx)),
	)
}

func TestSequencersByRollappGet(t *testing.T) {
	k, ctx := keepertest.SequencerKeeper(t)
	items := createNSequencer(k, ctx, 10)
	rst := k.GetSequencersByRollapp(ctx,
		items[0].RollappId,
	)

	require.Equal(t, len(rst), len(items))
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(rst),
	)
}

func (suite *SequencerTestSuite) TestRotatingSequencerByBond() {
	suite.SetupTest()
	rollappId, pk := suite.CreateDefaultRollapp()

	numOfSequencers := 5

	// create sequencers
	seqAddrs := make([]string, numOfSequencers)
	seqAddrs[0] = suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk)
	for j := 1; j < len(seqAddrs)-1; j++ {
		seqAddrs[j] = suite.CreateDefaultSequencer(suite.Ctx, rollappId, ed25519.GenPrivKey().PubKey())
	}
	// last one with high bond is the expected new proposer
	pkx := ed25519.GenPrivKey().PubKey()
	seqAddrs[len(seqAddrs)-1] = suite.CreateSequencerWithBond(suite.Ctx, rollappId, sdk.NewCoin(bond.Denom, bond.Amount.MulRaw(2)), pkx)
	expecetedProposer := seqAddrs[len(seqAddrs)-1]

	// check starting proposer and unbond
	sequencer, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, seqAddrs[0])
	suite.Require().True(found)
	suite.Require().True(sequencer.Proposer)

	suite.App.SequencerKeeper.RotateProposer(suite.Ctx, rollappId)

	// check proposer rotation
	newProposer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, expecetedProposer)
	suite.Equal(types.Bonded, newProposer.Status)
	suite.True(newProposer.Proposer)
}

package sequencer_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	genesisState := types.GenesisState{Params: types.DefaultParams()}
	k, ctx := keepertest.SequencerKeeper(t)
	sequencer.InitGenesis(ctx, *k, genesisState)
	require.Equal(t, genesisState.Params, k.GetParams(ctx))
}

func TestExportGenesis(t *testing.T) {
	params := types.Params{
		MinBond:       sdk.NewCoin("dym", sdk.NewInt(100)),
		UnbondingTime: 100,
	}
	sequencerList := []types.Sequencer{
		{
			SequencerAddress: "0",
			Status:           types.Bonded,
			Proposer:         true,
		},
		{
			SequencerAddress: "1",
			Status:           types.Bonded,
		},
	}
	k, ctx := keepertest.SequencerKeeper(t)
	k.SetParams(ctx, params)
	for _, sequencer := range sequencerList {
		k.SetSequencer(ctx, sequencer)
	}
	got := sequencer.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, params, got.Params)
	require.ElementsMatch(t, sequencerList, got.SequencerList)
}

package sequencer_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/sequencer"
	"github.com/dymensionxyz/dymension/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		SequencerList: []types.Sequencer{
			{
				SequencerAddress: "0",
			},
			{
				SequencerAddress: "1",
			},
		},
		SequencersByRollappList: []types.SequencersByRollapp{
			{
				RollappId: "0",
			},
			{
				RollappId: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.SequencerKeeper(t)
	sequencer.InitGenesis(ctx, *k, genesisState)
	got := sequencer.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.SequencerList, got.SequencerList)
	require.ElementsMatch(t, genesisState.SequencersByRollappList, got.SequencersByRollappList)
	// this line is used by starport scaffolding # genesis/test/assert
}

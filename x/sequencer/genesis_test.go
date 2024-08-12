package sequencer_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func TestInitGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		SequencerList: []types.Sequencer{
			{
				Address:  "0",
				Status:   types.Bonded,
				Proposer: true,
			},
			{
				Address: "1",
				Status:  types.Bonded,
			},
		},
		BondReductions: []types.BondReduction{
			{
				SequencerAddress:   "0",
				DecreaseBondAmount: sdk.NewCoin("dym", sdk.NewInt(100)),
				DecreaseBondTime:   time.Now().UTC(),
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
	require.ElementsMatch(t, genesisState.BondReductions, got.BondReductions)
	// this line is used by starport scaffolding # genesis/test/assert
}

func TestExportGenesis(t *testing.T) {
	params := types.Params{
		MinBond:                 sdk.NewCoin("dym", sdk.NewInt(100)),
		UnbondingTime:           100,
		LivenessSlashMultiplier: sdk.ZeroDec(),
	}
	sequencerList := []types.Sequencer{
		{
			Address:  "0",
			Status:   types.Bonded,
			Proposer: true,
		},
		{
			Address: "1",
			Status:  types.Bonded,
		},
	}
	bondReductions := []types.BondReduction{
		{
			SequencerAddress:   "0",
			DecreaseBondAmount: sdk.NewCoin("dym", sdk.NewInt(100)),
			DecreaseBondTime:   time.Now().UTC(),
		},
	}
	k, ctx := keepertest.SequencerKeeper(t)
	k.SetParams(ctx, params)
	for _, sequencer := range sequencerList {
		k.SetSequencer(ctx, sequencer)
	}
	for _, bondReduction := range bondReductions {
		k.SetDecreasingBondQueue(ctx, bondReduction)
	}

	got := sequencer.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, params, got.Params)
	require.ElementsMatch(t, sequencerList, got.SequencerList)
	require.ElementsMatch(t, bondReductions, got.BondReductions)
}

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
	tests := []struct {
		name       string
		params     types.Params
		sequencers []types.Sequencer
		expPanic   bool
	}{
		{
			name: "only params - success",
			params: types.Params{
				MinBond:       sdk.NewCoin("dym", sdk.NewInt(100)),
				UnbondingTime: 100,
			},
			sequencers: []types.Sequencer{},
			expPanic:   false,
		},
		{
			name: "params and demand order - panic",
			params: types.Params{
				MinBond:       sdk.NewCoin("dym", sdk.NewInt(100)),
				UnbondingTime: 100,
			},
			sequencers: []types.Sequencer{{SequencerAddress: "0"}},
			expPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesisState := types.GenesisState{Params: tt.params, SequencerList: tt.sequencers}
			k, ctx := keepertest.SequencerKeeper(t)
			if tt.expPanic {
				require.Panics(t, func() {
					sequencer.InitGenesis(ctx, *k, genesisState)
				})
			} else {
				sequencer.InitGenesis(ctx, *k, genesisState)
				params := k.GetParams(ctx)
				require.Equal(t, genesisState.Params, params)
			}
		})
	}
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

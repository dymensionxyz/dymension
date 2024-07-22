package sequencer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// InitGenesis initializes the sequencer module's state from a provided genesis
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the sequencer
	for _, elem := range genState.SequencerList {
		k.SetSequencer(ctx, elem)

		if elem.Status == types.Unbonding {
			k.SetUnbondingSequencerQueue(ctx, elem)
		}
	}
	k.SetParams(ctx, genState.Params)

	//FIXME: set proposer and nextProposer
}

// ExportGenesis returns the sequencer module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.GenesisState{}
	genesis.Params = k.GetParams(ctx)
	genesis.SequencerList = k.GetAllSequencers(ctx)
	return &genesis
}

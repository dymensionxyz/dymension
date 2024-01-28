package sequencer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the sequencer
	for _, elem := range genState.SequencerList {
		k.SetSequencer(ctx, elem)
	}
	// Set all the sequencersByRollapp
	for _, elem := range genState.SequencersByRollappList {
		k.SetSequencersByRollapp(ctx, elem)
	}
	// Set all the scheduler
	for _, elem := range genState.SchedulerList {
		k.SetScheduler(ctx, elem)
	}
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.SequencerList = k.GetAllSequencer(ctx)
	genesis.SequencersByRollappList = k.GetAllSequencersByRollapp(ctx)
	genesis.SchedulerList = k.GetAllScheduler(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}

package rollapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.RollappList = k.GetAllRollapps(ctx)
	genesis.StateInfoList = k.GetAllStateInfo(ctx)
	genesis.LatestStateInfoIndexList = k.GetAllLatestStateInfoIndex(ctx)
	genesis.LatestFinalizedStateIndexList = k.GetAllLatestFinalizedStateIndex(ctx)
	genesis.BlockHeightToFinalizationQueueList = k.GetAllBlockHeightToFinalizationQueue(ctx)

	return genesis
}

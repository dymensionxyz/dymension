package rollapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// InitGenesis initializes the capability module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the rollapp
	for _, elem := range genState.RollappList {
		k.SetRollapp(ctx, elem)
	}
	// Set all the stateInfo
	for _, elem := range genState.StateInfoList {
		k.SetStateInfo(ctx, elem)
	}
	// Set all the latestStateInfoIndex
	for _, elem := range genState.LatestStateInfoIndexList {
		k.SetLatestStateInfoIndex(ctx, elem)
	}
	// Set all the latestFinalizedStateIndex
	for _, elem := range genState.LatestFinalizedStateIndexList {
		k.SetLatestFinalizedStateIndex(ctx, elem)
	}
	// Set all the blockHeightToFinalizationQueue
	for _, elem := range genState.BlockHeightToFinalizationQueueList {
		k.SetBlockHeightToFinalizationQueue(ctx, elem)
	}
	for _, elem := range genState.LivenessEvents {
		k.PutLivenessEvent(ctx, elem)
	}
	// Set all the app
	for _, elem := range genState.AppList {
		k.SetApp(ctx, elem)
	}
	// Set rollapp registered denoms
	for _, elem := range genState.RegisteredDenoms {
		for _, denom := range elem.Denoms {
			if err := k.SetRegisteredDenom(ctx, elem.RollappId, denom); err != nil {
				panic(err)
			}
		}
	}

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
	genesis.LivenessEvents = k.GetLivenessEvents(ctx, nil)
	apps := k.GetRollappApps(ctx, "")
	var appList []types.App
	for _, app := range apps {
		appList = append(appList, *app)
	}
	genesis.AppList = appList

	var registeredRollappDenoms []types.RollappRegisteredDenoms
	for _, rollapp := range genesis.RollappList {
		denoms, err := k.GetAllRegisteredDenoms(ctx, rollapp.RollappId)
		if err != nil {
			panic(err)
		}
		registeredRollappDenoms = append(registeredRollappDenoms, types.RollappRegisteredDenoms{
			RollappId: rollapp.RollappId,
			Denoms:    denoms,
		})
	}
	genesis.RegisteredDenoms = registeredRollappDenoms

	return genesis
}

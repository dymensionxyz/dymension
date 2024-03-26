package rollapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the rollapp
	for _, elem := range genState.RollappList {

		// validate rollapp info
		err := elem.ValidateBasic()
		if err != nil {
			k.Logger(ctx).Error("error init genesis validating rollapp information: rollapp:%s", elem.RollappId)
			continue
		}
		// check to see if the RollappId has been registered before
		if _, isFound := k.GetRollapp(ctx, elem.RollappId); isFound {
			k.Logger(ctx).Error("error init genesis rollapp already exists: rollapp:%s", elem.RollappId)
			continue
		}
		// verify rollapp id
		rollappId, err := types.NewChainID(elem.RollappId)
		if err != nil {
			k.Logger(ctx).Error("error parsing Chain Id: rollapp:%s: Error:%s", elem.RollappId, err)
		}
		elem.RollappId = rollappId.ChainID
		k.SetRollapp(ctx, elem)
	}
	// Set all the stateInfo
	for _, elem := range genState.StateInfoList {

		stateInfo := types.NewMsgUpdateState(elem.Sequencer, elem.StateInfoIndex.RollappId, elem.StartHeight, elem.NumBlocks, elem.DAPath, elem.Version, &elem.BDs)
		err := stateInfo.ValidateBasic()
		if err != nil {
			k.Logger(ctx).Error("error init genesis validating state info: rollapp:%s: state info index: %d: Error:%s", elem.StateInfoIndex.RollappId, elem.StateInfoIndex.Index, err)
			continue
		}
		// load rollapp object for stateful validations
		rollapp, isFound := k.GetRollapp(ctx, elem.StateInfoIndex.RollappId)
		if !isFound {
			k.Logger(ctx).Error("error init genesis rollapp not found for state info: rollapp:%s: state info index: %s", elem.StateInfoIndex.RollappId, elem.StateInfoIndex.Index)
			continue
		}

		// check rollapp version
		if rollapp.Version != elem.Version {
			k.Logger(ctx).Error("error init genesis state info and rollapp version mismatch: rollapp:%s: state info index: %d: rollapp version:%d: state info version:%d", elem.StateInfoIndex.RollappId, elem.StateInfoIndex.Index, rollapp.Version, elem.Version)
			continue
		}

		k.SetStateInfo(ctx, elem)
	}
	// Set all the latestStateInfoIndex
	for _, elem := range genState.LatestStateInfoIndexList {

		// check the latest state info is found
		_, found := k.GetStateInfo(ctx, elem.RollappId, elem.Index)
		if !found {
			k.Logger(ctx).Error("error init genesis state info not found for latest state info index: rollapp:%s: latest state info index: %s:", elem.RollappId, elem.Index)
			continue
		}
		k.SetLatestStateInfoIndex(ctx, elem)

	}
	// Set all the latestFinalizedStateIndex
	for _, elem := range genState.LatestFinalizedStateIndexList {
		_, found := k.GetStateInfo(ctx, elem.RollappId, elem.Index)
		if !found {
			k.Logger(ctx).Error("error init genesis state info not found for latest finalized state info index: rollapp:%s: latest state info index: %s:", elem.RollappId, elem.Index)
			continue
		}
		latestStateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, elem.RollappId)
		if !found {
			k.Logger(ctx).Error("error init genesis latest state info index not found: rollapp:%s", elem.RollappId)
			continue
		}
		latestStateInfo, found := k.GetStateInfo(ctx, elem.RollappId, latestStateInfoIndex.Index)
		if !found {
			k.Logger(ctx).Error("error init genesis latest state info  not found: rollapp:%s: latest state info index: %s:", elem.RollappId, latestStateInfoIndex.Index)
			continue
		}
		if latestStateInfo.StateInfoIndex.Index < elem.Index {
			k.Logger(ctx).Error("error init genesis latest state info index lower than latest finalized state info: rollapp:%s: latest state info index: %s: latest finalized state info index:%s", elem.RollappId, latestStateInfoIndex.Index, elem.Index)
			continue
		}
		k.SetLatestFinalizedStateIndex(ctx, elem)

	}
	// Set all the blockHeightToFinalizationQueue
	for _, elem := range genState.BlockHeightToFinalizationQueueList {

		// set empty finalization queue
		queue := types.BlockHeightToFinalizationQueue{
			CreationHeight:    elem.CreationHeight,
			FinalizationQueue: []types.StateInfoIndex{},
		}
		for _, stateInfoIndex := range elem.FinalizationQueue {
			stateInfo, found := k.GetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			if !found || stateInfo.Status != common.Status_PENDING {
				// Invariant breaking
				k.Logger(ctx).Error("error init genesis failed to find state for finalization: rollappId %s, index %d, found %t, status %s",
					stateInfoIndex.RollappId, stateInfoIndex.Index, found, stateInfo.Status)
			} else {
				// adding state info to queue after validating state
				queue.FinalizationQueue = append(queue.FinalizationQueue, stateInfoIndex)
			}
		}
		k.SetBlockHeightToFinalizationQueue(ctx, queue)

	}
	// this line is used by starport scaffolding # genesis/module/init
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
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}

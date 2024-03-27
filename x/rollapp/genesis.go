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
		// verify rollapp id. err already checked in ValidateBasic
		rollappId, _ := types.NewChainID(elem.RollappId)

		elem.RollappId = rollappId.ChainID
		k.SetRollapp(ctx, elem)
	}
	// Set all the stateInfo
	for _, elem := range genState.StateInfoList {

		blockDescriptors := &types.BlockDescriptors{BD: elem.BDs.BD}

		stateInfo := types.NewMsgUpdateState(elem.Sequencer, elem.StateInfoIndex.RollappId, elem.StartHeight, elem.NumBlocks, elem.DAPath, elem.Version, blockDescriptors)
		err := stateInfo.ValidateBasic()
		if err != nil {
			k.Logger(ctx).Error("error init genesis validating state info: rollapp:%s: state info index: %d: Error:%s", elem.StateInfoIndex.RollappId, elem.StateInfoIndex.Index, err)
			continue
		}
		// load rollapp object for stateful validations
		rollapp, isFound := k.GetRollapp(ctx, elem.StateInfoIndex.RollappId)
		if !isFound {
			k.Logger(ctx).Error("error init genesis rollapp not found for state info: rollapp:%s: state info index: %d", elem.StateInfoIndex.RollappId, elem.StateInfoIndex.Index)
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
			k.Logger(ctx).Error("error init genesis state info not found for latest state info index: rollapp:%s: latest state info index: %d:", elem.RollappId, elem.Index)
			//invariant breaking. removing all state info for the rollapp
			removeAllStateInfo(ctx, k, elem.RollappId)
			continue
		}
		// check if there are state infos with higher index
		_, found = k.GetStateInfo(ctx, elem.RollappId, elem.Index+1)
		if found {
			k.Logger(ctx).Error("error init genesis state latest state info index is not the latest: rollapp:%s: latest state info index: %d:", elem.RollappId, elem.Index)
			//invariant breaking. removing all state info for the rollapp
			removeAllStateInfo(ctx, k, elem.RollappId)
			continue
		}
		k.SetLatestStateInfoIndex(ctx, elem)

	}
	// Set all the latestFinalizedStateIndex
	for _, elem := range genState.LatestFinalizedStateIndexList {

		// check the rollapp exists
		_, found := k.GetStateInfo(ctx, elem.RollappId, elem.Index)
		if !found {
			k.Logger(ctx).Error("error init genesis state info not found for latest finalized state info index: rollapp:%s: latest state info index: %d:", elem.RollappId, elem.Index)
			continue
		}

		// check there is a latest state info
		latestStateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, elem.RollappId)
		if !found {
			k.Logger(ctx).Error("error init genesis latest state info index not found: rollapp:%s", elem.RollappId)
			continue
		}
		latestStateInfo, found := k.GetStateInfo(ctx, elem.RollappId, latestStateInfoIndex.Index)
		if !found {
			k.Logger(ctx).Error("error init genesis latest state info not found: rollapp:%s: latest state info index: %d:", elem.RollappId, latestStateInfoIndex.Index)
			continue
		}

		// check the latest state info is not previous to the latest finalized state info
		if latestStateInfo.StateInfoIndex.Index < elem.Index {
			k.Logger(ctx).Error("error init genesis latest state info index lower than latest finalized state info: rollapp:%s: latest state info index: %d: latest finalized state info index:%d", elem.RollappId, latestStateInfoIndex.Index, elem.Index)
			//invariant breaking. removing all state info for the rollapp
			removeAllStateInfo(ctx, k, latestStateInfo.StateInfoIndex.RollappId)
			removeLatestStateInfo(ctx, k, latestStateInfo.StateInfoIndex.RollappId)
			continue
		}

		// check all previous state infos are finalized
		for i := uint64(1); i <= elem.Index; i++ {
			stateInfo, found := k.GetStateInfo(ctx, elem.RollappId, i)
			if !found || stateInfo.Status != common.Status_FINALIZED {
				k.Logger(ctx).Error("error init genesis there are non-finalized state infos previous to the finalized state info: rollapp:%s: state info index: %d: latest finalized state info index:%d", elem.RollappId, stateInfo.StateInfoIndex.Index, elem.Index)
				//invariant breaking. removing all state info for the rollapp
				removeAllStateInfo(ctx, k, latestStateInfo.StateInfoIndex.RollappId)
				removeLatestStateInfo(ctx, k, latestStateInfo.StateInfoIndex.RollappId)
				break
			}
		}

		k.SetLatestFinalizedStateIndex(ctx, elem)

	}
	// Set all the blockHeightToFinalizationQueue
	for _, elem := range genState.BlockHeightToFinalizationQueueList {

		// check all state infos from the queue and all only those that are found and not finalized
		queue := types.BlockHeightToFinalizationQueue{
			CreationHeight:    elem.CreationHeight,
			FinalizationQueue: []types.StateInfoIndex{},
		}
		for _, stateInfoIndex := range elem.FinalizationQueue {
			stateInfo, found := k.GetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			if !found {
				// Invariant breaking
				k.Logger(ctx).Error("error init genesis failed to find state in finalization queue: rollappId %s, index %d",
					stateInfoIndex.RollappId, stateInfoIndex.Index)
				continue
			}
			if stateInfo.Status != common.Status_PENDING {
				k.Logger(ctx).Error("error init genesis state info in finalization queue is not in pending state: rollappId %s, index %d",
					stateInfoIndex.RollappId, stateInfoIndex.Index)
				continue
			}
			latestFinalizedIndex, found := k.GetLatestFinalizedStateIndex(ctx, stateInfoIndex.RollappId)
			if !found {
				k.Logger(ctx).Error("error init genesis latest finalized index not found: rollappId %s, index %d",
					stateInfoIndex.RollappId, stateInfoIndex.Index)
				continue
			}
			latestFinalizedStateInfo, found := k.GetStateInfo(ctx, stateInfoIndex.RollappId, latestFinalizedIndex.Index)
			if !found {
				k.Logger(ctx).Error("error init genesis latest finalized state info not found: rollappId %s, index %d",
					stateInfoIndex.RollappId, stateInfoIndex.Index)
				continue
			}

			if latestFinalizedStateInfo.StateInfoIndex.Index >= stateInfo.StateInfoIndex.Index {
				// Invariant breaking
				k.Logger(ctx).Error("error init genesis state info in finalization queue should be finalized: rollappId %s, index: %d, finalized index:%d",
					stateInfoIndex.RollappId, stateInfoIndex.Index, latestFinalizedStateInfo.StateInfoIndex.Index)
			}
			// adding state info to queue after validating state
			queue.FinalizationQueue = append(queue.FinalizationQueue, stateInfoIndex)

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

func removeAllStateInfo(ctx sdk.Context, k keeper.Keeper, rollappId string) {

	index, found := k.GetLatestStateInfoIndex(ctx, rollappId)
	if found {
		for i := uint64(1); i <= index.Index; i++ {
			k.RemoveStateInfo(ctx, rollappId, i)
		}
	} else {
		i := uint64(1)
		for {
			_, found := k.GetStateInfo(ctx, rollappId, i)
			if !found {
				break
			}
			if found {
				k.RemoveStateInfo(ctx, rollappId, i)
			}
			i++
		}
	}
}
func removeLatestStateInfo(ctx sdk.Context, k keeper.Keeper, rollappId string) {
	k.RemoveLatestStateInfoIndex(ctx, rollappId)
}

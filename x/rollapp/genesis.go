package rollapp

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) error {
	// Set all the rollapp
	for _, elem := range genState.RollappList {

		// validate rollapp info
		err := elem.ValidateBasic()
		if err != nil {
			removeAllGenesisState(ctx, k)
			return errorsmod.Wrapf(sdkerrors.ErrLogic, "error init genesis validating rollapp information: rollapp:%s", elem.RollappId)
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
			removeAllGenesisState(ctx, k)
			return errorsmod.Wrapf(sdkerrors.ErrLogic, "error init genesis validating state info: rollapp:%s: state info index: %d: Error:%s", elem.StateInfoIndex.RollappId, elem.StateInfoIndex.Index, err)
		}
		// if rollapp is not found, state info is not added
		_, isFound := k.GetRollapp(ctx, elem.StateInfoIndex.RollappId)
		if !isFound {
			removeAllGenesisState(ctx, k)
			return errorsmod.Wrapf(sdkerrors.ErrLogic, "error init genesis rollapp not found for state info: rollapp:%s: state info index: %d", elem.StateInfoIndex.RollappId, elem.StateInfoIndex.Index)
		}

		k.SetStateInfo(ctx, elem)
	}

	err := checkAllRollapsStateInfo(ctx, k)
	if err != nil {
		return err
	}
	buildBlockHeightToFinalizationQueue(ctx, k)
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
	return nil
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.RollappList = k.GetAllRollapps(ctx)
	genesis.StateInfoList = k.GetAllStateInfo(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}

func removeAllGenesisState(ctx sdk.Context, k keeper.Keeper) {
	for _, rollapp := range k.GetAllRollapps(ctx) {
		k.RemoveRollapp(ctx, rollapp.RollappId)
	}
	for _, stateInfo := range k.GetAllStateInfo(ctx) {
		k.RemoveStateInfo(ctx, stateInfo.StateInfoIndex.RollappId, stateInfo.StateInfoIndex.Index)
	}
}

func checkAllRollapsStateInfo(ctx sdk.Context, k keeper.Keeper) error {
	rollappsList := k.GetAllRollapps(ctx)

	for _, rollapp := range rollappsList {
		stateInfoList := k.GetAllRollappStateInfo(ctx, rollapp.RollappId)
		previousIndex := uint64(0)
		lastFinalized := uint64(0)
		prevStateInfoLastBlock := uint64(0)
		stateInfoIsCorrect := true
		var errormsg string
		for _, stateInfo := range stateInfoList {
			if stateInfo.StateInfoIndex.Index != previousIndex+1 {
				errormsg = "missing state info"
				stateInfoIsCorrect = false
				break
			}
			if stateInfo.Status == common.Status_FINALIZED {
				if lastFinalized != stateInfo.StateInfoIndex.Index-1 {
					errormsg = "inconsistent finalized status"
					stateInfoIsCorrect = false
					break
				}
				lastFinalized = stateInfo.StateInfoIndex.Index
			}

			if stateInfo.StartHeight-1 != prevStateInfoLastBlock {
				errormsg = "missing block in state info"
				stateInfoIsCorrect = false
				break
			}
			prevStateInfoLastBlock = stateInfo.GetLatestHeight()
			previousIndex = stateInfo.StateInfoIndex.Index
		}
		if !stateInfoIsCorrect {
			removeAllGenesisState(ctx, k)
			return errorsmod.Wrapf(sdkerrors.ErrLogic, "error init genesis validating state info rollapp:%s: error:%s", rollapp.RollappId, errormsg)
		}
		if previousIndex != uint64(0) {
			k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{RollappId: rollapp.RollappId, Index: previousIndex})
		}
		if lastFinalized != uint64(0) {
			k.SetLatestFinalizedStateIndex(ctx, types.StateInfoIndex{RollappId: rollapp.RollappId, Index: lastFinalized})
		}
	}
	return nil
}

func buildBlockHeightToFinalizationQueue(ctx sdk.Context, k keeper.Keeper) {
	blockHeightToFinalizationQueue := make(map[uint64][]types.StateInfoIndex)
	rollappsList := k.GetAllRollapps(ctx)

	for _, rollapp := range rollappsList {
		stateInfoList := k.GetAllRollappStateInfo(ctx, rollapp.RollappId)

		for _, stateInfo := range stateInfoList {
			if stateInfo.Status == common.Status_PENDING {
				// load FinalizationQueue and update
				finalizationQueue, found := blockHeightToFinalizationQueue[stateInfo.CreationHeight]
				newFinalizationQueue := []types.StateInfoIndex{stateInfo.StateInfoIndex}
				if found {
					newFinalizationQueue = append(finalizationQueue, newFinalizationQueue...)
				}

				blockHeightToFinalizationQueue[stateInfo.CreationHeight] = newFinalizationQueue
			}
		}
	}

	for height, finalizationQueue := range blockHeightToFinalizationQueue {
		k.SetBlockHeightToFinalizationQueue(ctx, types.BlockHeightToFinalizationQueue{CreationHeight: height, FinalizationQueue: finalizationQueue})
	}
}

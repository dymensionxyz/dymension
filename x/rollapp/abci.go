package rollapp

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dymensionxyz/dymension/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called on every block.
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
}

// Called every block to finalize states that their dispute period over.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// check to see if there are pending  states to be finalized
	blockHeightToFinalizationQueue, found := k.GetBlockHeightToFinalizationQueue(ctx, uint64(ctx.BlockHeight()))

	if found {
		// finalize pending states
		for _, stateInfoIndex := range blockHeightToFinalizationQueue.FinalizationQueue {
			stateInfo, found := k.GetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			if !found {
				panic(fmt.Errorf("invariant broken: rollapp %s should have state to be finalizaed in height %d. update state index is %d \n",
					stateInfoIndex.RollappId, ctx.BlockHeight(), stateInfoIndex.Index))
			}
			stateInfo.Status = types.STATE_STATUS_FINALIZED
			// update the status of the stateInfo
			k.SetStateInfo(ctx, stateInfo)
			// uppdate the LatestStateInfoIndex of the rollapp
			k.SetLatestFinalizedStateIndex(ctx, stateInfoIndex)
			// call the after-update-state hook
			keeperHooks := k.GetHooks()
			err := keeperHooks.AfterStateFinalized(ctx, stateInfoIndex.RollappId, &stateInfo)
			if err != nil {
				ctx.Logger().Error("Error after state finalized", "rollappID", stateInfoIndex.RollappId, "error", err.Error())
			}

			// emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(types.EventTypeStatusChange,
					sdk.NewAttribute(types.AttributeKeyRollappId, stateInfoIndex.RollappId),
					sdk.NewAttribute(types.AttributeKeyStateInfoIndex, strconv.FormatUint(stateInfoIndex.Index, 10)),
					sdk.NewAttribute(types.AttributeKeyStartHeight, strconv.FormatUint(stateInfo.StartHeight, 10)),
					sdk.NewAttribute(types.AttributeKeyNumBlocks, strconv.FormatUint(stateInfo.NumBlocks, 10)),
					sdk.NewAttribute(types.AttributeKeyStatus, stateInfo.Status.String()),
				),
			)
		}
	}
}

package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

// AfterStateFinalized implements the RollappHooks interface
func (im IBCMiddleware) AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *rollapptypes.StateInfo) error {
	// Finalize the packets for the rollapp at the given height
	stateEndHeight := stateInfo.StartHeight + stateInfo.NumBlocks - 1
	im.FinalizeRollappPackets(ctx, rollappID, stateEndHeight)
	return nil
}

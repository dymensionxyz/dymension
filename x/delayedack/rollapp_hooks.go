package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ rollapptypes.RollappHooks = &IBCMiddleware{}

func (im IBCMiddleware) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error {
	return nil
}

// AfterStateFinalized implements the RollappHooks interface
func (im IBCMiddleware) AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *rollapptypes.StateInfo) error {
	// Finalize the packets for the rollapp at the given height
	stateEndHeight := stateInfo.StartHeight + stateInfo.NumBlocks - 1
	return im.Keeper.FinalizeRollappPackets(ctx, im.IBCModule, rollappID, stateEndHeight)
}

func (im IBCMiddleware) FraudSubmitted(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error {
	return im.Keeper.HandleFraud(ctx, rollappID, im.IBCModule)
}

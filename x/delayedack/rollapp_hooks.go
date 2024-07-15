package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ rollapptypes.RollappHooks = &IBCMiddleware{}

func (w IBCMiddleware) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error {
	return nil
}

// AfterStateFinalized implements the RollappHooks interface
func (w IBCMiddleware) AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *rollapptypes.StateInfo) error {
	// Finalize the packets for the rollapp at the given height
	stateEndHeight := stateInfo.StartHeight + stateInfo.NumBlocks - 1
	return w.FinalizeRollappPackets(ctx, w.IBCModule, rollappID, stateEndHeight)
}

func (w IBCMiddleware) FraudSubmitted(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error {
	return w.HandleFraud(ctx, rollappID, w.IBCModule)
}

// RollappCreated implements types.RollappHooks.
func (im IBCMiddleware) RollappCreated(ctx sdk.Context, rollappID string) error {
	return nil
}

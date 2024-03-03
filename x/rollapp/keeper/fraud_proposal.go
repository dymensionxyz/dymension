package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

//
//
//
//HandleFraud

// HandleFraud handles the fraud evidence submitted by the user.
func (k Keeper) HandleFraud(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error {
	// Get the rollapp from the store
	_, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidRollappID, "rollapp with ID %s not found", rollappID)
	}

	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, height)
	if err != nil {
		return err
	}

	if stateInfo.Sequencer != seqAddr {
		return sdkerrors.Wrapf(types.ErrInvalidSequencer, "sequencer address %s does not match the one in the state info", seqAddr)
	}

	// slash the sequencer
	err = k.hooks.FraudSubmitted(ctx, rollappID, height, seqAddr)
	if err != nil {
		return err
	}

	// freeze the ibc channel

	// Emit an event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraud,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
			sdk.NewAttribute(types.AttributeKeyFraudHeight, string(height)),
			sdk.NewAttribute(types.AttributeKeyFraudSequencer, seqAddr),
		),
	)

	return nil
}

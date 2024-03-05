package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	tmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// HandleFraud handles the fraud evidence submitted by the user.
func (k Keeper) HandleFraud(ctx sdk.Context, rollappID, clientId string, height uint64, seqAddr string) error {
	// Get the rollapp from the store
	_, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidRollappID, "rollapp with ID %s not found", rollappID)
	}

	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, height)
	if err != nil {
		return err
	}

	//TODO: mark the rollapp as frozen (if immutable) or mark the fraud height to allow overwriting

	if stateInfo.Sequencer != seqAddr {
		return sdkerrors.Wrapf(types.ErrInvalidSequencer, "sequencer address %s does not match the one in the state info", seqAddr)
	}

	// slash the sequencer
	err = k.hooks.FraudSubmitted(ctx, rollappID, height, seqAddr)
	if err != nil {
		return err
	}

	//FIXME: make sure the clientId corresponds to the rollappID

	clientState, ok := k.ibcclientkeeper.GetClientState(ctx, clientId)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidClientState, "client state for clientID %s not found", clientId)
	}

	// Set the client state to frozen
	tmClientState, ok := clientState.(*tmtypes.ClientState)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidClientState, "client state with ID %s is not a tendermint client state", clientId)
	}

	tmClientState.FrozenHeight = clienttypes.NewHeight(tmClientState.GetLatestHeight().GetRevisionHeight(), tmClientState.GetLatestHeight().GetRevisionNumber())
	k.ibcclientkeeper.SetClientState(ctx, clientId, tmClientState)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraud,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
			sdk.NewAttribute(types.AttributeKeyFraudHeight, fmt.Sprint(height)),
			sdk.NewAttribute(types.AttributeKeyFraudSequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyClientID, clientId),
		),
	)

	return nil
}

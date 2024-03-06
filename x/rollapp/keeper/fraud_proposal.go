package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	tmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// HandleFraud handles the fraud evidence submitted by the user.
func (k Keeper) HandleFraud(ctx sdk.Context, rollappID, clientId string, height uint64, seqAddr string) error {
	// Get the rollapp from the store
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidRollappID, "rollapp with ID %s not found", rollappID)
	}

	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, height)
	if err != nil {
		return err
	}

	//check height is not finalized
	if stateInfo.Status == common.Status_FINALIZED {
		return sdkerrors.Wrapf(types.ErrDisputeAlreadyFinalized, "state info for height %d is already finalized", height)
	}

	//check the sequencer for this height is the same as the one in the fraud evidence
	if stateInfo.Sequencer != seqAddr {
		return sdkerrors.Wrapf(types.ErrWrongProposerAddr, "sequencer address %s does not match the one in the state info", seqAddr)
	}

	// slash the sequencer, clean delayed packets
	err = k.hooks.FraudSubmitted(ctx, rollappID, height, seqAddr)
	if err != nil {
		return err
	}

	//mark the rollapp as frozen. revert all pending states to finalized
	rollapp.Frozen = true
	k.SetRollapp(ctx, rollapp)

	//TODO: go over all pending states, and set as disputed those beloning to the rollapp
	// {
	// 	stateInfo.Status = types.STATE_STATUS_DISPUTED
	// 	k.SetStateInfo(ctx, stateInfo)
	// }

	//TODO: get the clientId from rollapp object, instead of by proposal
	clientState, ok := k.ibcclientkeeper.GetClientState(ctx, clientId)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidClientState, "client state for clientID %s not found", clientId)
	}

	tmClientState, ok := clientState.(*tmtypes.ClientState)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidClientState, "client state with ID %s is not a tendermint client state", clientId)
	}

	//validate the clientId related to the disputed rollapp
	if tmClientState.ChainId != rollappID {
		return sdkerrors.Wrapf(types.ErrWrongClientId, "client state with ID %s is not related to rollapp with ID %s", clientId, rollappID)
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

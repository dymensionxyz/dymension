package keeper

import (
	"bytes"
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/sequencer/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.DymintPubKey == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "sequencer pubkey can not be empty")
	}
	// load rollapp object for stateful validations
	rollapp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	// check to see if the rollapp has been registered before
	if !found {
		return nil, types.ErrUnknownRollappID
	}
	// check if there are permissionedAddresses.
	// if the list is not empty, it means that only premissioned sequencers can be added
	permissionedAddresses := rollapp.PermissionedAddresses
	if len(permissionedAddresses) > 0 {
		bPermissioned := false
		// check to see if the sequencer is in the permissioned list
		for _, addr := range permissionedAddresses {
			if addr == msg.Creator {
				// Found!
				bPermissioned = true
				break
			}
		}
		// Err: only permissioned sequencers allowed and this one is not in the list
		if !bPermissioned {
			return nil, types.ErrSequencerNotPermissioned
		}
	}

	// check to see if the sequencer has been registered before
	sequencer, found := k.GetSequencer(ctx, msg.Creator)
	if !found {
		sequencer = types.Sequencer{
			SequencerAddress: msg.Creator,
			DymintPubKey:     msg.DymintPubKey,
			Description:      msg.Description,
			RollappIDs:       []string{msg.RollappId},
		}

		k.SetSequencer(ctx, sequencer)
	} else {
		//validate same data of the sequencer
		if !bytes.Equal(sequencer.DymintPubKey.GetValue(), msg.DymintPubKey.GetValue()) {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "sequencer pubkey does not match")
		}
		//ignore new description

		// check to see if the rollappId matches the one of the sequencer
		for _, rollapp := range sequencer.RollappIDs {
			if rollapp == msg.RollappId {
				return nil, types.ErrSequencerAlreadyRegistered
			}
		}
		// add rollappId to sequencer
		sequencer.RollappIDs = append(sequencer.RollappIDs, msg.RollappId)
		k.SetSequencer(ctx, sequencer)
	}

	// update sequencers list
	sequencersByRollapp, found := k.GetSequencersByRollapp(ctx, msg.RollappId)
	if found {
		// check to see if we reached maxsimum number of sequeners
		maxSequencers := int(rollapp.MaxSequencers)
		activeSequencers := sequencersByRollapp.Sequencers
		currentNumOfSequencers := len(activeSequencers)
		if maxSequencers < currentNumOfSequencers {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "rollapp id: %s cannot have more than %d sequencers but got: %d", msg.RollappId, maxSequencers, currentNumOfSequencers)
		}
		if maxSequencers == currentNumOfSequencers {
			return nil, types.ErrMaxSequencersLimit
		}
		// add sequencer to list
		sequencersByRollapp.Sequencers = append(sequencersByRollapp.Sequencers, sequencer.SequencerAddress)
		// it's not the first sequencer, make it INACTIVE
		scheduler := types.Scheduler{
			SequencerAddress: msg.Creator,
			Status:           types.Inactive,
		}
		k.SetScheduler(ctx, scheduler)
	} else {
		// this is the first sequencer, make it a PROPOSER
		sequencersByRollapp.RollappId = msg.RollappId
		sequencersByRollapp.Sequencers = append(sequencersByRollapp.Sequencers, msg.Creator)
		scheduler := types.Scheduler{
			SequencerAddress: msg.Creator,
			Status:           types.Proposer,
		}
		k.SetScheduler(ctx, scheduler)
	}
	k.SetSequencersByRollapp(ctx, sequencersByRollapp)

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	return &types.MsgCreateSequencerResponse{}, nil
}

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/sequencer/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Pubkey can be nil only in simulation mode
	if !k.isSimulation {
		if msg.DymintPubKey == nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "sequencer pubkey can not be empty")
		}
	}

	// check to see if the sequencer has been registered before
	if _, found := k.GetSequencer(ctx, msg.Creator); found {
		return nil, types.ErrSequencerExists
	}

	// load rollapp object for stateful validations
	rollapp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	// check to see if the rollapp has been registered before
	if found {
		// check if there are permissionedAddresses.
		// if the list is not empty, it means that only premissioned sequencers can be added
		permissionedAddresses := rollapp.PermissionedAddresses.Addresses
		if len(permissionedAddresses) > 0 {
			bPermissioned := false
			// check to see if the sequencer is in the permissioned list
			for i := range permissionedAddresses {
				if permissionedAddresses[i] == msg.Creator {
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
	} else {
		return nil, types.ErrUnknownRollappID
	}

	// update sequencers list
	sequencersByRollapp, found := k.GetSequencersByRollapp(ctx, msg.RollappId)
	if found {
		// check to see if we reached maxsimum number of sequeners
		maxSequencers := int(rollapp.MaxSequencers)
		activeSequencers := sequencersByRollapp.Sequencers
		currentNumOfSequencers := len(activeSequencers.Addresses)
		if maxSequencers < currentNumOfSequencers {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "rollapp id: %s cannot have more than %d sequencers but got: %d", msg.RollappId, maxSequencers, currentNumOfSequencers)
		}
		if maxSequencers == currentNumOfSequencers {
			return nil, types.ErrMaxSequencersLimit
		}
		// add sequencer to list
		sequencersByRollapp.Sequencers.Addresses = append(sequencersByRollapp.Sequencers.Addresses, msg.Creator)
		// it's not the first sequencer, make it INACTIVE
		scheduler := types.Scheduler{
			SequencerAddress: msg.Creator,
			Status:           types.Inactive,
		}
		k.SetScheduler(ctx, scheduler)
	} else {
		// this is the first sequencer, make it a PROPOSER
		sequencersByRollapp.RollappId = msg.RollappId
		sequencersByRollapp.Sequencers.Addresses = append(sequencersByRollapp.Sequencers.Addresses, msg.Creator)
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

	sequencer := types.Sequencer{
		SequencerAddress: msg.Creator,
		DymintPubKey:     msg.DymintPubKey,
		Description:      msg.Description,
		RollappId:        msg.RollappId,
	}

	k.SetSequencer(ctx, sequencer)

	return &types.MsgCreateSequencerResponse{}, nil
}

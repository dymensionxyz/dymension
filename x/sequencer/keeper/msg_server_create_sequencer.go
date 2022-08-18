package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/sequencer/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	seqAddr, err := sdk.AccAddressFromBech32(msg.SequencerAddress)
	if err != nil {
		return nil, err
	}

	seqAddrStr := seqAddr.String()

	// check to see if the sequencer has been registered before
	if _, found := k.GetSequencer(ctx, seqAddrStr); found {
		return nil, types.ErrSequencerExists
	}

	// load rollapp object for stateful validations
	rollapp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	// check to see if the rollapp has been registered before
	if found {
		// check if there are permissionedAddresses.
		// if the list is not empty, it means that only premissioned sequencers can be added
		permissionedAddresses := rollapp.GetPermissionedAddresses().Addresses
		if len(permissionedAddresses) > 0 {
			bPermissioned := false
			// check to see if the sequencer is in the permissioned list
			for i := range permissionedAddresses {
				if permissionedAddresses[i] == seqAddrStr {
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
		return nil, types.ErrUnknownRollappId
	}

	// update sequencers list
	sequencersByRollapp, found := k.GetSequencersByRollapp(ctx, msg.RollappId)
	if found {
		// check to see if we reached maxsimum number of sequeners
		maxSequencers := int(rollapp.GetMaxSequencers())
		activeSequencers := sequencersByRollapp.GetSequencers()
		currentNumOfSequencers := len(activeSequencers.GetAddresses())
		if maxSequencers < currentNumOfSequencers {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "rollapp id: %s cannot have more than %d sequencers but got: %d", msg.RollappId, maxSequencers, currentNumOfSequencers)
		}
		if maxSequencers == currentNumOfSequencers {
			return nil, types.ErrMaxSequencersLimit
		}
		// add sequencer to list
		sequencersByRollapp.Sequencers.Addresses = append(sequencersByRollapp.Sequencers.Addresses, seqAddrStr)
	} else {
		// should be: return nil, types.XXXX
		sequencersByRollapp.RollappId = msg.RollappId
		sequencersByRollapp.Sequencers.Addresses = append(sequencersByRollapp.Sequencers.Addresses, seqAddrStr)
	}
	k.SetSequencersByRollapp(ctx, sequencersByRollapp)

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	pkAny, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return nil, err
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	sequencer := types.Sequencer{
		Creator:          creator.String(),
		SequencerAddress: seqAddrStr,
		Pubkey:           pkAny,
		Description:      msg.Description,
		RollappId:        msg.RollappId,
	}

	k.SetSequencer(ctx, sequencer)

	return &types.MsgCreateSequencerResponse{}, nil
}

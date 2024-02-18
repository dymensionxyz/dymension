package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.DymintPubKey == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "sequencer pubkey can not be empty")
	}

	// check to see if the sequencer has been registered before
	if _, found := k.GetSequencer(ctx, msg.Creator); found {
		return nil, types.ErrSequencerExists
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
		if !bPermissioned {
			return nil, types.ErrSequencerNotPermissioned
		}
	}

	// check to see if the sequencer has enough balance and deduct the bond
	seqAcc, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	//TODO: use custom error codes
	minBond := k.GetParams(ctx).MinBond
	if msg.Bond.Denom != minBond.Denom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Bond.Denom, minBond.Denom,
		)
	}

	if msg.Bond.Amount.LT(minBond.Amount) {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "insufficient bond: got %s, expected %s", msg.Bond.Amount, k.GetParams(ctx).MinBond,
		)
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, seqAcc, types.ModuleName, sdk.NewCoins(msg.Bond))
	if err != nil {
		return nil, err
	}
	sequencer := types.Sequencer{
		SequencerAddress: msg.Creator,
		DymintPubKey:     msg.DymintPubKey,
		RollappId:        msg.RollappId,
		Description:      msg.Description,
		Status:           types.Bonded,
		Tokens:           &msg.Bond,
	}

	// update sequencers list
	//FIXME: change to get only bonded sequencers
	sequencersByRollapp, found := k.GetSequencersByRollapp(ctx, msg.RollappId)
	// check to see if we reached the maximum number of sequeners for this rollapp
	maxSequencers := int(rollapp.MaxSequencers)
	currentNumOfSequencers := len(sequencersByRollapp.Sequencers.Addresses)
	if currentNumOfSequencers >= maxSequencers {
		return nil, types.ErrMaxSequencersLimit
	}
	if !found {
		// this is the first sequencer, make it a PROPOSER
		sequencersByRollapp.RollappId = msg.RollappId
		sequencer.Status = types.Proposer
	}
	sequencersByRollapp.Sequencers.Addresses = append(sequencersByRollapp.Sequencers.Addresses, msg.Creator)
	k.SetSequencersByRollapp(ctx, sequencersByRollapp)

	k.SetSequencer(ctx, sequencer)

	return &types.MsgCreateSequencerResponse{}, nil
}

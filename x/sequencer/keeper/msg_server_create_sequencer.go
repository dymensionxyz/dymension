package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	sdkerrors "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.DymintPubKey == nil {
		return nil, sdkerrors.Wrapf(errortypes.ErrInvalidPubKey, "sequencer pubkey can not be empty")
	}

	// check to see if the sequencer has been registered before
	if _, found := k.GetSequencer(ctx, msg.Creator); found {
		return nil, types.ErrSequencerExists
	}

	// check to see if the rollapp has been registered before
	rollapp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, types.ErrUnknownRollappID
	}
	if rollapp.Frozen {
		return nil, types.ErrRollappJailed
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

	bond := sdk.Coins{}
	minBond := k.GetParams(ctx).MinBond
	if !minBond.IsNil() && !minBond.IsZero() {
		if msg.Bond.Denom != minBond.Denom {
			return nil, sdkerrors.Wrapf(
				types.ErrInvalidCoinDenom, "got %s, expected %s", msg.Bond.Denom, minBond.Denom,
			)
		}

		if msg.Bond.Amount.LT(minBond.Amount) {
			return nil, sdkerrors.Wrapf(
				types.ErrInsufficientBond, "got %s, expected %s", msg.Bond.Amount, minBond,
			)
		}

		err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, seqAcc, types.ModuleName, sdk.NewCoins(msg.Bond))
		if err != nil {
			return nil, err
		}
		bond = sdk.NewCoins(msg.Bond)
	}

	sequencer := types.Sequencer{
		SequencerAddress: msg.Creator,
		DymintPubKey:     msg.DymintPubKey,
		RollappId:        msg.RollappId,
		Description:      msg.Description,
		Status:           types.Bonded,
		Proposer:         false,
		Tokens:           bond,
	}

	bondedSequencers := k.GetSequencersByRollappByStatus(ctx, msg.RollappId, types.Bonded)
	unbondingSequencers := k.GetSequencersByRollappByStatus(ctx, msg.RollappId, types.Unbonding)
	// check to see if we reached the maximum number of sequencers for this rollapp
	currentNumOfSequencers := len(bondedSequencers) + len(unbondingSequencers)
	if rollapp.MaxSequencers > 0 && uint64(currentNumOfSequencers) >= rollapp.MaxSequencers {
		return nil, types.ErrMaxSequencersLimit
	}
	// this is the first sequencer, make it a PROPOSER
	if len(bondedSequencers) == 0 {
		sequencer.Proposer = true
	}

	k.SetSequencer(ctx, sequencer)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateSequencer,
			sdk.NewAttribute(types.AttributeKeyRollappId, msg.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBond, msg.Bond.String()),
			sdk.NewAttribute(types.AttributeKeyProposer, strconv.FormatBool(sequencer.Proposer)),
		),
	)

	return &types.MsgCreateSequencerResponse{}, nil
}

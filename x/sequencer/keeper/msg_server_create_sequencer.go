package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidRequest, "validate basic: %v", err)
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

	// check to see if the sequencer has enough balance and deduct the bond
	seqAcc, _ := sdk.AccAddressFromBech32(msg.Creator)

	bond := sdk.Coins{}
	minBond := k.GetParams(ctx).MinBond
	if !minBond.IsNil() && !minBond.IsZero() {
		if msg.Bond.Denom != minBond.Denom {
			return nil, errorsmod.Wrapf(
				types.ErrInvalidCoinDenom, "got %s, expected %s", msg.Bond.Denom, minBond.Denom,
			)
		}

		if msg.Bond.Amount.LT(minBond.Amount) {
			return nil, errorsmod.Wrapf(
				types.ErrInsufficientBond, "got %s, expected %s", msg.Bond.Amount, minBond,
			)
		}

		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, seqAcc, types.ModuleName, sdk.NewCoins(msg.Bond))
		if err != nil {
			return nil, err
		}
		bond = sdk.NewCoins(msg.Bond)
	}

	sequencer := types.Sequencer{
		Address:      msg.Creator,
		DymintPubKey: msg.DymintPubKey,
		RollappId:    msg.RollappId,
		Metadata:     msg.Metadata,
		Status:       types.Bonded,
		Proposer:     false,
		Tokens:       bond,
	}

	bondedSequencers := k.GetSequencersByRollappByStatus(ctx, msg.RollappId, types.Bonded)

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

package keeper

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check to see if the rollapp has been registered before
	rollapp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, types.ErrRollappNotFound
	}

	if err := msg.VMSpecificValidate(rollapp.VmType); err != nil {
		return nil, errors.Join(types.ErrInvalidRequest, err)
	}

	// check to see if the seq has been registered before
	if _, err := k.tryGetSequencer(ctx, msg.Creator); err == nil {
		return nil, types.ErrSequencerAlreadyExists
	}

	// In case InitialSequencer is set to one or more bech32 addresses, only one of them can be the first to register,
	// and is automatically selected as the first proposer, allowing the Rollapp to be set to 'launched'
	// (provided that all the immutable fields are set in the Rollapp).
	// This limitation prevents scenarios such as:
	// a) any unintended initial seq getting registered before the immutable fields are set in the Rollapp.
	// b) situation when seq "X" is registered prior to the initial seq,
	// after which the initial seq's address is set to seq X's address, effectively preventing:
	// 	1. the initial seq from getting selected as the first proposer,
	// 	2. the rollapp from getting launched again
	// In case the InitialSequencer is set to the "*" wildcard, any seq can be the first to register.
	if !rollapp.Launched {
		isInitialOrAllAllowed := slices.Contains(strings.Split(rollapp.InitialSequencer, ","), msg.Creator) || rollapp.InitialSequencer == "*"
		if !isInitialOrAllAllowed {
			return nil, types.ErrNotInitialSequencer
		}

		// check pre launch time.
		// skipped if no pre launch time is set
		if rollapp.PreLaunchTime != nil && rollapp.PreLaunchTime.After(ctx.BlockTime()) {
			return nil, types.ErrBeforePreLaunchTime
		}

		if err := k.rollappKeeper.SetRollappAsLaunched(ctx, &rollapp); err != nil {
			return nil, err
		}
	}

	if err := k.sufficientBond(ctx, msg.Bond); err != nil {
		return nil, err
	}

	// send bond to module account
	seqAcc := sdk.MustAccAddressFromBech32(msg.Creator)
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, seqAcc, types.ModuleName, sdk.NewCoins(msg.Bond))
	if err != nil {
		return nil, err
	}

	bond := sdk.NewCoins(msg.Bond)
	seq := types.Sequencer{
		Address:      msg.Creator,
		DymintPubKey: msg.DymintPubKey,
		RollappId:    msg.RollappId,
		Metadata:     msg.Metadata,
		Status:       types.Bonded,
		Tokens:       bond,
	}

	/*
		TODO: need to stop registration when awaiting the last block of the proposer and the successor is sentinel
		because the proposer might already have produced the last block, which would turn out to be wrong
	*/

	if err := k.ChooseProposer(ctx, msg.RollappId); err != nil {
		return nil, err
	}

	k.SetSequencer(ctx, seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateSequencer,
			sdk.NewAttribute(types.AttributeKeyRollappId, msg.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBond, msg.Bond.String()),
			sdk.NewAttribute(types.AttributeKeyProposer, strconv.FormatBool(k.isProposer(ctx, seq))),
		),
	)

	return &types.MsgCreateSequencerResponse{}, nil
}

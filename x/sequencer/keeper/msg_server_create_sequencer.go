package keeper

import (
	"context"
	"slices"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check to see if the rollapp has been registered before
	rollapp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, types.ErrUnknownRollappID
	}

	if rollapp.Frozen {
		return nil, types.ErrRollappJailed
	}

	if err := msg.VMSpecificValidate(rollapp.VmType); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidRequest, err.Error())
	}

	// check to see if the sequencer has been registered before
	if _, found = k.GetSequencer(ctx, msg.Creator); found {
		return nil, types.ErrSequencerExists
	}

	// In case InitialSequencer is set to one or more bech32 addresses, only one of them can be the first to register,
	// and is automatically selected as the first proposer, allowing the Rollapp to be sealed
	// (provided that all the immutable fields are set in the Rollapp).
	// This limitation prevents scenarios such as:
	// a) any unintended initial sequencer getting registered before the immutable fields are set in the Rollapp.
	// b) situation when sequencer "X" is registered prior to the initial sequencer,
	// after which the initial sequencer's address is set to sequencer X's address, effectively preventing:
	// 	1. the initial sequencer from getting selected as the first proposer,
	// 	2. the rollapp from getting sealed
	// In case the InitialSequencer is set to the "*" wildcard, any sequencer can be the first to register.
	if !rollapp.Sealed {
		isInitialOrAllAllowed := slices.Contains(strings.Split(rollapp.InitialSequencer, ","), msg.Creator) || rollapp.InitialSequencer == "*"
		if !isInitialOrAllAllowed {
			return nil, types.ErrNotInitialSequencer
		}
		if err := k.rollappKeeper.SealRollapp(ctx, msg.RollappId); err != nil {
			return nil, err
		}
	}

	bond := sdk.Coins{}
	if minBond := k.GetParams(ctx).MinBond; !(minBond.IsNil() || minBond.IsZero()) {
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

		seqAcc := sdk.MustAccAddressFromBech32(msg.Creator)
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
		Tokens:       bond,
	}

	// we currently only support setting next proposer (or empty one) before the rotation started. This is in order to
	// avoid handling the case a potential next proposer bonds in the middle of a rotation.
	// This will be handled in next iteration.
	nextProposer, ok := k.GetNextProposer(ctx, msg.RollappId)
	if ok && nextProposer.IsEmpty() {
		k.Logger(ctx).Info("rotation in progress. sequencer registration disabled", "rollappId", sequencer.RollappId)
		return nil, types.ErrRotationInProgress
	}

	// if no proposer set for he rollapp, set this sequencer as the proposer
	_, proposerExists := k.GetProposer(ctx, msg.RollappId)
	if !proposerExists {
		k.SetProposer(ctx, sequencer.RollappId, sequencer.Address)
	}

	k.SetSequencer(ctx, sequencer)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateSequencer,
			sdk.NewAttribute(types.AttributeKeyRollappId, msg.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBond, msg.Bond.String()),
			sdk.NewAttribute(types.AttributeKeyProposer, strconv.FormatBool(!proposerExists)),
		),
	)

	return &types.MsgCreateSequencerResponse{}, nil
}

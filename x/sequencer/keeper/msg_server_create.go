package keeper

import (
	"context"
	"slices"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check to see if the rollapp has been registered before
	rollapp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, rollapptypes.ErrRollappNotFound
	}

	// check to see if the seq has been registered before
	if _, err := k.RealSequencer(ctx, msg.Creator); err == nil {
		return nil, types.ErrSequencerAlreadyExists
	}

	pkAddr, err := types.PubKeyAddr(msg.DymintPubKey)
	if err != nil {
		return nil, errorsmod.Wrap(err, "pub key addr")
	}
	if _, err := k.SequencerByDymintAddr(ctx, pkAddr); err == nil {
		return nil, gerrc.ErrAlreadyExists.Wrap("pub key in use")
	}

	if err := k.sufficientBond(ctx, msg.RollappId, msg.Bond); err != nil {
		return nil, err
	}

	if err := msg.VMSpecificValidate(rollapp.VmType); err != nil {
		return nil, errorsmod.Wrap(err, "vm specific validate")
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
		isInitialSeq := slices.Contains(strings.Split(rollapp.InitialSequencer, ","), msg.Creator)
		anyAllowed := rollapp.InitialSequencer == "*"
		if !anyAllowed && !isInitialSeq {
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

	seq := k.NewSequencer(ctx, msg.RollappId)

	// set a reward address. if empty, use a creator address.
	rewardAddr := msg.RewardAddr
	if msg.RewardAddr == "" {
		rewardAddr = msg.Creator
	}

	seq.RewardAddr = rewardAddr
	seq.DymintPubKey = msg.DymintPubKey
	seq.Address = msg.Creator
	seq.Status = types.Bonded
	seq.Metadata = msg.Metadata
	seq.OptedIn = true
	seq.SetWhitelistedRelayers(msg.WhitelistedRelayers)

	if err := k.sendToModule(ctx, seq, msg.Bond); err != nil {
		return nil, err
	}

	k.SetSequencer(ctx, *seq)
	if err := k.SetSequencerByDymintAddr(ctx, pkAddr, seq.Address); err != nil {
		return nil, errorsmod.Wrapf(err, "set sequencer by dymint addr: %s: proposer hash: %x", seq.Address, pkAddr)
	}

	proposer := k.GetProposer(ctx, msg.RollappId)
	if proposer.Sentinel() {
		if err := k.RecoverFromSentinel(ctx, msg.RollappId); err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateSequencer,
			sdk.NewAttribute(types.AttributeKeyRollappId, msg.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBond, msg.Bond.String()),
			sdk.NewAttribute(types.AttributeKeyProposer, strconv.FormatBool(k.IsProposer(ctx, *seq))),
		),
	)

	return &types.MsgCreateSequencerResponse{}, nil
}

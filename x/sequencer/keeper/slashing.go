package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Slashing slashes the sequencer for misbehaviour
// Slashing can occur on both Bonded and Unbonding sequencers
func (k Keeper) Slashing(ctx sdk.Context, seqAddr string) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Status == types.Unbonded {
		return errorsmod.Wrap(
			types.ErrInvalidSequencerStatus,
			"can't slash unbonded sequencer",
		)
	}

	seqTokens := seq.Tokens
	if !seqTokens.Empty() {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, seqTokens)
		if err != nil {
			return err
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to slash", "sequencer", seq.Address)
	}
	seq.Tokens = sdk.Coins{}

	oldStatus := seq.Status
	wasProposer := seq.Proposer
	// in case we are slashing an unbonding sequencer, we need to remove it from the unbonding queue
	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
	} else {
		// in case the sequencer is currently reducing its bond, then we need to remove it from the decreasing bond queue
		// all the tokens are burned, so we don't need to reduce the bond anymore
		bondReduce, isDecreasingBond := k.getSequencerDecreasingBond(ctx, seq.SequencerAddress)
		if isDecreasingBond {
			k.removeDecreasingBondQueue(ctx, bondReduce)
		}
	}

	// set the status to unbonded
	seq.Status = types.Unbonded
	seq.Jailed = true
	seq.Proposer = false
	seq.UnbondingHeight = ctx.BlockHeight()
	seq.UnbondTime = ctx.BlockHeader().Time
	k.UpdateSequencer(ctx, seq, oldStatus)

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlashed,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyBond, seqTokens.String()),
		),
	)

	// rotate proposer if the slashed sequencer was the proposer
	if wasProposer {
		k.RotateProposer(ctx, seq.RollappId)
	}

	return nil
}

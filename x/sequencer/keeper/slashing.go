package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Unbond defines a method for removing coins from sequencer's bond
// Slashing can occur on both Bonded and Unbonding sequencers
func (k Keeper) Slashing(ctx sdk.Context, seqAddr string) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Status == types.Unbonded {
		return sdkerrors.Wrap(
			types.ErrInvalidSequencerStatus,
			"cant slash unbonded sequencer",
		)
	}

	if !seq.Tokens.Empty() {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, seq.Tokens)
		if err != nil {
			return err
		}
		seq.Tokens = sdk.Coins{}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to slash", "sequencer", seq.SequencerAddress)
	}

	oldStatus := seq.Status
	wasPropser := seq.Proposer
	//in case we are slashing an unbonding sequencer, we need to remove it from the unbonding queue
	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
	}

	// set the status to unbonded
	seq.Status = types.Unbonded
	seq.Jailed = true
	seq.Proposer = false
	seq.UnbondingHeight = ctx.BlockHeight()
	seq.UnbondTime = ctx.BlockHeader().Time
	k.UpdateSequencer(ctx, seq, oldStatus)

	// rotate proposer if the slashed sequencer was the proposer
	if wasPropser {
		k.RotateProposer(ctx, seq.RollappId)
	}

	return nil
}

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

	tokens := seq.Tokens

	if err := k.Slash(ctx, seq, tokens, nil); err != nil {
		return err // TODO:
	}

	if err := k.Jail(ctx, seq); err != nil {
		return err // TODO:
	}

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlashed,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyBond, tokens.String()),
		),
	)

	return nil
}

func MulCoinsDec(coins sdk.Coins, dec sdk.Dec) sdk.Coins {
	// TODO: use it
	return sdk.Coins{}
}

func (k Keeper) Slash(ctx sdk.Context, seq types.Sequencer, amt sdk.Coins, recipientAddr *sdk.AccAddress) error {
	seq.Tokens.Sub(amt...)
	if recipientAddr != nil {
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, *recipientAddr, amt)
		if err != nil {
			return errorsmod.Wrap(err, "send coins from module to account")
		}
	} else {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt)
		if err != nil {
			return errorsmod.Wrap(err, "burn coins")
		}
	}
	// TODO: write back sequencer?
	return nil
}

func (k Keeper) Jail(ctx sdk.Context, seq types.Sequencer) error {
	// TODO: check contents of this, since it was copied

	oldStatus := seq.Status
	wasProposer := seq.Proposer
	// in case we are slashing an unbonding sequencer, we need to remove it from the unbonding queue
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
	if wasProposer {
		k.RotateProposer(ctx, seq.RollappId)
	}

	return nil
}

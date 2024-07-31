package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/ucoin"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// SlashAndJailFraud slashes the sequencer for misbehaviour other than liveness issues
// Can occur on both Bonded and Unbonding sequencers
func (k Keeper) SlashAndJailFraud(ctx sdk.Context, seqAddr string) error {
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

	// slash 100% of tokens
	if err := k.Slash(ctx, &seq, tokens); err != nil {
		return errorsmod.Wrap(err, "slash")
	}

	if err := k.Jail(ctx, seq); err != nil {
		return errorsmod.Wrap(err, "jail")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlashed,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyBond, tokens.String()),
		),
	)

	return nil
}

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	seq, err := k.LivenessLiableSequencer(ctx, rollappID)
	if err != nil {
		return err
	}
	mul := k.GetParams(ctx).LivenessSlashMultiplier
	tokens := seq.Tokens
	amt := ucoin.MulDec(mul, tokens...)
	// TODO: make sure to be correct wrt. min bond, see https://github.com/dymensionxyz/dymension/issues/1019
	return k.Slash(ctx, &seq, amt)
}

func (k Keeper) JailLiveness(ctx sdk.Context, rollappID string) error {
	seq, err := k.LivenessLiableSequencer(ctx, rollappID)
	if err != nil {
		return errorsmod.Wrap(err, "liveness liable sequencer")
	}
	return k.Jail(ctx, seq)
}

// LivenessLiableSequencer returns the sequencer who is responsible for ensuring liveness
func (k Keeper) LivenessLiableSequencer(ctx sdk.Context, rollappID string) (types.Sequencer, error) {
	// TODO: find the sequencer who is currently responsible for ensuring liveness
	//  https://github.com/dymensionxyz/dymension/issues/1018
	return types.Sequencer{}, errorsmod.Wrap(gerrc.ErrNotFound, "not implemented")
}

func (k Keeper) Slash(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coins) error {
	seq.Tokens = seq.Tokens.Sub(amt...)
	err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt)
	if err != nil {
		return errorsmod.Wrap(err, "burn coins")
	}
	return nil
}

func (k Keeper) Jail(ctx sdk.Context, seq types.Sequencer) error {
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

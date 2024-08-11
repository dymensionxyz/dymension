package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

// SlashAndJailFraud slashes the sequencer for misbehaviour other than liveness issues
// Can occur on both Bonded and Unbonding sequencers
func (k Keeper) SlashAndJailFraud(ctx sdk.Context, seqAddr string) error {
	seq, err := k.unbondSequencerAndBurn(ctx, seqAddr)
	if err != nil {
		return fmt.Errorf("slash sequencer: %w", err)
	}

	seq.Jailed = true
	seq.UnbondRequestHeight = ctx.BlockHeight()
	seq.UnbondTime = ctx.BlockTime()
	k.UpdateSequencer(ctx, *seq)

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlashed,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyBond, seqTokens.String()),
		),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeJailed,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
		),
	)

	return nil
}

// InstantUnbondAllSequencers unbonds all sequencers for a rollapp
// This is called by the `FraudSubmitted` hook
func (k Keeper) InstantUnbondAllSequencers(ctx sdk.Context, rollappID string) error {
	// unbond all bonded/unbonding sequencers
	bonded := k.GetSequencersByRollappByStatus(ctx, rollappID, types.Bonded)
	unbonding := k.GetSequencersByRollappByStatus(ctx, rollappID, types.Unbonding)
	for _, sequencer := range append(bonded, unbonding...) {
		err := k.unbondSequencer(ctx, sequencer.SequencerAddress)
		if err != nil {
			return err
		}
	}

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
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlashed,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
			sdk.NewAttribute(types.AttributeKeyBond, amt.String()),
		),
	)
	return nil
}

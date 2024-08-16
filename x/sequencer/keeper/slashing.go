package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

func (k Keeper) JailSequencerOnFraud(ctx sdk.Context, seqAddr string) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	if err := k.Jail(ctx, seq); err != nil {
		return errorsmod.Wrap(err, "jail")
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
	proposer, found := k.GetProposer(ctx, rollappID)
	if !found {
		return types.Sequencer{}, types.ErrNoProposer
	}
	return proposer, nil
}

func (k Keeper) Slash(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coins) error {
	if seq.Status == types.Unbonded {
		return errorsmod.Wrap(
			types.ErrInvalidSequencerStatus,
			"can't slash unbonded sequencer",
		)
	}

	err := k.reduceSequencerBond(ctx, seq, amt, true)
	if err != nil {
		return errorsmod.Wrap(err, "remove sequencer bond")
	}
	k.UpdateSequencer(ctx, *seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlashed,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
			sdk.NewAttribute(types.AttributeKeyBond, amt.String()),
		),
	)
	return nil
}

func (k Keeper) Jail(ctx sdk.Context, seq types.Sequencer) error {
	err := k.unbondSequencerAndJail(ctx, seq.Address)
	if err != nil {
		return errorsmod.Wrap(err, "unbond and jail")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeJailed,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
			sdk.NewAttribute(types.AttributeKeyBond, seq.Tokens.String()),
		),
	)

	return nil
}

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
			"cant slash unbonded sequencer",
		)
	}

	seqTokens := seq.Tokens
	if !seqTokens.Empty() {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, seqTokens)
		if err != nil {
			return err
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to slash", "sequencer", seq.SequencerAddress)
	}
	seq.Tokens = sdk.Coins{}

	oldStatus := seq.Status
	// in case we are slashing an unbonding sequencer, we need to remove it from the unbonding queue
	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
	}

	// FIXME: check for notice period queue and remove if needed

	//FIXME: check for proposer/next and remove if needed

	// set the status to unbonded
	seq.Status = types.Unbonded
	seq.Jailed = true

	seq.UnbondRequestHeight = ctx.BlockHeight()
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

	return nil
}

// in case of fraud, we need to unbond all other bonded sequencers as the rollapp is frozen
func (k Keeper) forceUnbondSequencer(ctx sdk.Context, seqAddr string) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Status == types.Unbonded {
		return errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is already unbonded",
		)
	}

	oldStatus := seq.Status

	seqTokens := seq.Tokens
	if !seqTokens.Empty() {
		seqAcc, err := sdk.AccAddressFromBech32(seq.SequencerAddress)
		if err != nil {
			return err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, seqTokens)
		if err != nil {
			return err
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to unbond", "sequencer", seq.SequencerAddress)
	}

	// set the status to unbonded and remove from the unbonding queue if needed
	seq.Status = types.Unbonded
	seq.Tokens = sdk.Coins{}

	k.UpdateSequencer(ctx, seq, oldStatus)

	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
	}

	//TODO: clear notice period queue if needed

	//TODO: clear proposer/next if needed

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyBond, seqTokens.String()),
		),
	)

	return nil
}

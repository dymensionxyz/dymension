package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"
)

// UnbondCondition defines an unbond condition implementer.
// It is implemented by modules.
// Returning false means the sequencer will not be allowed to unbond, it should also
// contain the unbond reason.
type UnbondCondition interface {
	CanUnbond(ctx sdk.Context, sequencer types.Sequencer) error
}

func (k Keeper) tryUnbond(ctx sdk.Context, seq types.Sequencer, amt *sdk.Coin) error {
	if k.IsProposerOrSuccessor(ctx, seq) {
		return types.ErrUnbondProposerOrSuccessor
	}
	for _, c := range k.unbondConditions {
		if err := c.CanUnbond(ctx, seq); err != nil {
			return errorsmod.Wrap(err, "other module can unbond")
		}
	}
	if amt != nil {
		// partial refund
		bond := seq.TokensCoin()
		minBond := k.GetParams(ctx).MinBond
		maxReduction, _ := bond.SafeSub(minBond)
		if maxReduction.IsLT(*amt) {
			return errorsmod.Wrapf(types.ErrUnbondNotAllowed,
				"attempted reduction: %s, max reduction: %s",
				*amt, ucoin.NonNegative(maxReduction),
			)
		}
		return errorsmod.Wrap(k.refundTokens(ctx, seq, *amt), "refund")
	}
	// total refund + unbond
	if err := k.refundTokens(ctx, seq, seq.TokensCoin()); err != nil {
		return errorsmod.Wrap(err, "refund")
	}
	return errorsmod.Wrap(k.unbond(ctx, seq), "unbond")
}

func (k Keeper) unbond(ctx sdk.Context, seq types.Sequencer) error {
	if k.isNextProposer(ctx, seq) {
		return gerrc.ErrInternal.Wrap("unbond next proposer")
	}
	seq.Status = types.Unbonded
	if k.isProposer(ctx, seq) {
		k.SetProposer(ctx, seq.RollappId, SentinelSeqAddr)
	}
	return nil
}

func (k Keeper) refundTokens(ctx sdk.Context, seq types.Sequencer, amt sdk.Coin) error {
	return errorsmod.Wrap(k.moveTokens(ctx, seq, amt, uptr.To(seq.AccAddr())), "move tokens")
}

func (k Keeper) moveTokens(ctx sdk.Context, seq types.Sequencer, amt sdk.Coin, recipient *sdk.AccAddress) error {
	amts := sdk.NewCoins(amt)
	if recipient != nil {
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, *recipient, amts)
		if err != nil {
			return errorsmod.Wrap(err, "bank send")
		}
	} else {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amts)
		if err != nil {
			return errorsmod.Wrap(err, "burn")
		}
	}
	// TODO: write object
	return nil
}

// unbond unbonds a sequencer
// if jail is true, the sequencer is jailed as well (cannot be bonded again)
// bonded tokens are refunded by default, unless jail is true
func (k Keeper) unbondLegacy(ctx sdk.Context, seqAddr string, jail bool) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrSequencerNotFound
	}

	if seq.Status == types.Unbonded {
		return errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is already unbonded",
		)
	}
	// keep the old status for updating the sequencer
	oldStatus := seq.Status

	// handle bond: tokens refunded by default, unless jail is true
	err := k.reduceSequencerBond(ctx, &seq, seq.Tokens, jail)
	if err != nil {
		return errorsmod.Wrap(err, "remove sequencer bond")
	}

	/* ------------------------------ store cleanup ----------------------------- */
	// remove from queue if unbonding
	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
	} else {
		// remove from notice period queue if needed
		if seq.IsNoticePeriodInProgress() {
			k.removeNoticePeriodSequencer(ctx, seq)
		}

		// if we unbond the proposer, remove it
		// the caller should rotate the proposer
		if k.isProposer(ctx, seq.RollappId, seqAddr) {
			k.removeProposer(ctx, seq.RollappId)
		}

		// if we unbond the next proposer, we're in the middle of rotation
		// the caller should clean the rotation state
		if k.isNextProposer(ctx, seq.RollappId, seqAddr) {
			k.removeNextProposer(ctx, seq.RollappId)
		}
	}
	// in case the sequencer is currently reducing its bond, then we need to remove it from the decreasing bond queue
	// all the tokens are returned, so we don't need to reduce the bond anymore
	if bondReductionIDs := k.getBondReductionIDsBySequencer(ctx, seq.Address); len(bondReductionIDs) > 0 {
		for _, bondReductionID := range bondReductionIDs {
			bondReduction, found := k.GetBondReduction(ctx, bondReductionID)
			if found {
				k.removeBondReduction(ctx, bondReductionID, bondReduction)
			}
		}
	}

	if jail {
		seq.Jailed = true
	}
	// set the unbonding height and time, if not already set.
	// to avoid leaving unbonded sequencer in the store with no unbond height or time
	if seq.UnbondRequestHeight == 0 {
		seq.UnbondRequestHeight = ctx.BlockHeight()
	}
	if seq.UnbondTime.IsZero() {
		seq.UnbondTime = ctx.BlockTime()
	}

	// update the sequencer in store
	seq.Status = types.Unbonded
	k.UpdateSequencer(ctx, &seq, oldStatus)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
		),
	)

	return nil
}

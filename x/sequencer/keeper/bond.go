package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

// UnbondBlocker allows vetoing unbond attempts
type UnbondBlocker interface {
	// CanUnbond should return a types.UnbondNotAllowed error with a reason, or nil (or another error)
	CanUnbond(ctx sdk.Context, sequencer types.Sequencer) error
}

// TryUnbond will try to either partially or totally unbond a sequencer.
// The sequencer may not be allowed to unbond, based on certain conditions.
// A partial unbonding refunds tokens, but doesn't allow the remaining bond to fall below a threshold.
// A total unbond refunds all tokens and changes status to unbonded.
func (k Keeper) TryUnbond(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin) error {
	if k.IsProposer(ctx, *seq) || k.IsSuccessor(ctx, *seq) {
		return types.ErrUnbondProposerOrSuccessor
	}
	for _, c := range k.unbondBlockers {
		if err := c.CanUnbond(ctx, *seq); err != nil {
			return errorsmod.Wrap(err, "other module")
		}
	}
	bond := seq.TokensCoin()
	minBond := k.GetParams(ctx).MinBond
	maxReduction, _ := bond.SafeSub(minBond)
	isPartial := !amt.IsEqual(bond)
	if isPartial && maxReduction.IsLT(amt) {
		return errorsmod.Wrapf(types.ErrUnbondNotAllowed,
			"attempted reduction: %s, max reduction: %s",
			amt, ucoin.NonNegative(maxReduction),
		)
	}
	if err := k.refund(ctx, seq, amt); err != nil {
		return errorsmod.Wrap(err, "refund")
	}
	if seq.Tokens.IsZero() {
		return errorsmod.Wrap(k.unbond(ctx, seq), "unbond")
	}
	return nil
}

// set unbonded status and clear proposer/successor if necessary
func (k Keeper) unbond(ctx sdk.Context, seq *types.Sequencer) error {
	if k.IsSuccessor(ctx, *seq) {
		return gerrc.ErrInternal.Wrap(`unbond next proposer: it shouldnt be possible because
they cannot do frauds and they cannot unbond gracefully`)
	}
	seq.Status = types.Unbonded

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
		),
	)
	if k.IsProposer(ctx, *seq) {
		k.SetProposer(ctx, seq.RollappId, types.SentinelSeqAddr)
		// we assume the current successor will not be happy if the proposer suddenly unbonds
		k.SetSuccessor(ctx, seq.RollappId, types.SentinelSeqAddr)
	}
	return nil
}

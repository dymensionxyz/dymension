package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

type UnbondChecker interface {
	// CanUnbond should return a types.UnbondNotAllowed error with a reason, or nil (or another error)
	CanUnbond(ctx sdk.Context, sequencer types.Sequencer) error
}

func (k Keeper) tryUnbond(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin) error {
	if k.IsProposerOrSuccessor(ctx, *seq) {
		return types.ErrUnbondProposerOrSuccessor
	}
	for _, c := range k.unbondConditions {
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
	if err := k.refundTokens(ctx, seq, amt); err != nil {
		return errorsmod.Wrap(err, "refund")
	}
	if seq.Tokens.IsZero() {
		return errorsmod.Wrap(k.unbond(ctx, seq), "unbond")
	}
	return nil
}

func (k Keeper) unbond(ctx sdk.Context, seq *types.Sequencer) error {
	if k.isNextProposer(ctx, seq) {
		return gerrc.ErrInternal.Wrap("unbond next proposer")
	}
	seq.Status = types.Unbonded
	if k.isProposerLeg(ctx, seq) {
		k.SetProposer(ctx, seq.RollappId, types.SentinelSeqAddr)
	}
	return nil
}

func (k Keeper) burnTokens(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin) error {
	seq.SetTokensCoin(seq.TokensCoin().Sub(amt))
	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amt))
}

func (k Keeper) refundTokens(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin) error {
	return errorsmod.Wrap(k.sendTokens(ctx, seq, amt, seq.AccAddr()), "send tokens")
}

func (k Keeper) sendTokens(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin, recipient sdk.AccAddress) error {
	seq.SetTokensCoin(seq.TokensCoin().Sub(amt))
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(amt))
}

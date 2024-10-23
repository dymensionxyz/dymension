package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/dymensionxyz/dymension-rdk/x/sequencers/types"
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
		k.SetProposer(ctx, seq.RollappId, types2.SentinelSeqAddr)
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

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// TryKickProposer tries to remove the incumbent proposer. It requires the incumbent
// proposer to be below a threshold of bond. The caller must also be bonded and opted in.
func (k Keeper) TryKickProposer(ctx sdk.Context, kicker types.Sequencer) error {
	if !kicker.IsPotentialProposer() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "kicker is not a potential proposer")
	}

	ra := kicker.RollappId

	proposer := k.GetProposer(ctx, ra)

	if !k.Kickable(ctx, proposer) {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "not kickable")
	}

	// clear the proposer
	k.abruptRemoveProposer(ctx, ra)

	// This will call hard fork on the rollapp, which will also optOut all sequencers
	err := k.hooks.AfterKickProposer(ctx, proposer)
	if err != nil {
		return errorsmod.Wrap(err, "kick proposer callbacks")
	}

	// optIn the kicker
	if err := kicker.SetOptedIn(ctx, true); err != nil {
		return errorsmod.Wrap(err, "set opted in")
	}
	k.SetSequencer(ctx, kicker)

	// this will choose kicker as next proposer, since he is the only opted in and bonded
	// sequencer remaining.
	if err := k.ChooseProposerAfterSentinel(ctx, ra); err != nil {
		return errorsmod.Wrap(err, "choose proposer")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventKickedProposer{
		Rollapp:  ra,
		Kicker:   kicker.Address,
		Proposer: proposer.Address,
	}); err != nil {
		return err
	}

	return nil
}

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	seq := k.GetProposer(ctx, rollappID)
	if seq.Sentinel() {
		return nil
	}

	// correct formula is e.g. min(sequencer tokens, max(1, sequencer tokens * 0.01 ))

	err := k.livenessSlash(ctx, &seq)
	if err != nil {
		return errorsmod.Wrap(err, "slash")
	}
	k.increasePenaltyDowntime(ctx, &seq)
	k.SetSequencer(ctx, seq)
	return nil
}

func (k Keeper) livenessSlash(ctx sdk.Context, seq *types.Sequencer) error {
	mul := k.GetParams(ctx).LivenessSlashMinMultiplier
	abs := k.GetParams(ctx).LivenessSlashMinAbsolute
	tokens := seq.TokensCoin()
	tokensMul := ucoin.MulDec(mul, tokens)
	amt := ucoin.SimpleMin(tokens, ucoin.SimpleMax(abs, tokensMul[0]))
	return errorsmod.Wrap(k.slash(ctx, seq, amt, math.LegacyZeroDec(), nil), "slash")
}

func (k Keeper) reducePenaltyUptime(ctx sdk.Context, seq *types.Sequencer) {
	diff := k.GetParams(ctx).PenaltyReductionStateUpdate()
	diff = min(diff, seq.GetPenalty())
	seq.SetPenalty(seq.GetPenalty() - diff)
}

func (k Keeper) increasePenaltyDowntime(ctx sdk.Context, seq *types.Sequencer) {
	penalty := k.GetParams(ctx).PenaltyLiveness()
	seq.SetPenalty(seq.GetPenalty() + penalty)
}

// Takes an optional rewardee addr who will receive some bounty
// Currently there is no dishonor penalty (anyway we slash 100%)
func (k Keeper) PunishSequencer(ctx sdk.Context, seqAddr string, rewardee *sdk.AccAddress) error {
	var (
		rewardMul = math.LegacyZeroDec()
		addr      = []byte(nil)
	)

	seq, err := k.RealSequencer(ctx, seqAddr)
	if err != nil {
		return err
	}

	if rewardee != nil {
		rewardMul = math.LegacyMustNewDecFromStr("0.5") // TODO: parameterise
		addr = *rewardee
	}

	err = k.slash(ctx, &seq, seq.TokensCoin(), rewardMul, addr)
	if err != nil {
		return errorsmod.Wrap(err, "slash")
	}
	k.SetSequencer(ctx, seq)
	return nil
}

func (k Keeper) slash(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin, rewardMul math.LegacyDec, rewardee sdk.AccAddress) error {
	rewardCoin := ucoin.MulDec(rewardMul, amt)[0]
	if !rewardCoin.IsZero() {
		err := k.sendFromModule(ctx, seq, rewardCoin, rewardee)
		if err != nil {
			return errorsmod.Wrap(err, "send")
		}
	}
	remainder := amt.Sub(rewardCoin)
	err := errorsmod.Wrap(k.burn(ctx, seq, remainder), "burn")
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlashed,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
			sdk.NewAttribute(types.AttributeKeyRemainingAmt, seq.TokensCoin().String()),
			sdk.NewAttribute(types.AttributeKeyAmt, amt.String()),
		),
	)
	return err
}

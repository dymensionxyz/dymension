package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

func (k Keeper) KickProposer(ctx sdk.Context, kicker types.Sequencer) error {
	if !kicker.IsPotentialProposer() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "not ready to propose")
	}

	ra := kicker.RollappId

	proposer := k.GetProposer(ctx, ra)

	if k.Kickable(ctx, proposer) {
		if err := k.unbond(ctx, &proposer); err != nil {
			return errorsmod.Wrap(err, "unbond")
		}
		k.SetSequencer(ctx, proposer)
		if err := k.optOutAllSequencers(ctx, ra, kicker.Address); err != nil {
			return errorsmod.Wrap(err, "opt out all seqs")
		}
		k.hooks.AfterKickProposer(ctx, proposer)

		if err := uevent.EmitTypedEvent(ctx, &types.EventKickedProposer{
			Rollapp:  ra,
			Kicker:   kicker.Address,
			Proposer: proposer.Address,
		}); err != nil {
			return err
		}
	}

	if err := k.ChooseProposer(ctx, ra); err != nil {
		return errorsmod.Wrap(err, "choose proposer")
	}

	return nil
}

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	seq := k.GetProposer(ctx, rollappID)
	if seq.Sentinel() {
		return nil
	}
	mul := k.GetParams(ctx).LivenessSlashMinMultiplier
	abs := k.GetParams(ctx).LivenessSlashMinAbsolute
	tokens := seq.TokensCoin()
	tokensMul := ucoin.MulDec(mul, tokens)
	amt := ucoin.SimpleMin(tokens, ucoin.SimpleMax(abs, tokensMul[0]))
	err := errorsmod.Wrap(k.slash(ctx, &seq, amt, sdk.ZeroDec(), nil), "slash")
	k.SetSequencer(ctx, seq)
	return err
}

func (k Keeper) HandleFraud(ctx sdk.Context, seq types.Sequencer, rewardee *sdk.AccAddress) error {
	var err error
	if rewardee != nil {
		rewardMul := sdk.MustNewDecFromStr("0.5") // TODO: parameterise
		err = k.slash(ctx, &seq, seq.TokensCoin(), rewardMul, *rewardee)
	} else {
		err = k.slash(ctx, &seq, seq.TokensCoin(), sdk.ZeroDec(), nil)
	}
	if err != nil {
		return errorsmod.Wrap(err, "slash")
	}
	err = errorsmod.Wrap(k.unbond(ctx, &seq), "unbond")
	k.SetSequencer(ctx, seq)
	if err := k.optOutAllSequencers(ctx, seq.RollappId); err != nil {
		return errorsmod.Wrap(err, "opt out all seqs")
	}
	return err
}

func (k Keeper) slash(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin, rewardMul sdk.Dec, rewardee sdk.AccAddress) error {
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

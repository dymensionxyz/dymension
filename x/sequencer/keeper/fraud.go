package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

func (k Keeper) abruptRemoveSequencer(ctx sdk.Context, seq types.Sequencer) error {

	if k.IsSuccessor(ctx, seq) {
		return gerrc.ErrUnimplemented.Wrap("successor")
	}

	if !k.IsProposer(ctx, seq) {
		return nil
	}

	ra := seq.RollappId
	k.SetProposer(ctx, ra, types.SentinelSeqAddr)

	k.unbond(ctx, &seq)
	k.SetSequencer(ctx, seq)

	// This will call hard fork on the rollapp, which will also optOut all sequencers
	return k.hooks.AfterKickProposer(ctx, ra)
}

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	seq := k.GetProposer(ctx, rollappID)
	if seq.Sentinel() {
		return nil
	}

	// correct formula is e.g. min(sequencer tokens, max(1, sequencer tokens * 0.01 ))

	mul := k.GetParams(ctx).LivenessSlashMinMultiplier
	abs := k.GetParams(ctx).LivenessSlashMinAbsolute
	tokens := seq.TokensCoin()
	tokensMul := ucoin.MulDec(mul, tokens)
	amt := ucoin.SimpleMin(tokens, ucoin.SimpleMax(abs, tokensMul[0]))
	err := errorsmod.Wrap(k.slash(ctx, &seq, amt, sdk.ZeroDec(), nil), "slash")
	k.SetSequencer(ctx, seq)
	return err
}

// Takes an optional rewardee addr who will receive some bounty
func (k Keeper) SlashAllTokens(ctx sdk.Context, seqAddr string, rewardee *sdk.AccAddress) error {
	var (
		rewardMul = sdk.ZeroDec()
		addr      = []byte(nil)
	)

	seq, err := k.RealSequencer(ctx, seqAddr)
	if err != nil {
		return err
	}

	if rewardee != nil {
		rewardMul = sdk.MustNewDecFromStr("0.5") // TODO: parameterise
		addr = *rewardee
	}

	err = k.slash(ctx, &seq, seq.TokensCoin(), rewardMul, addr)
	if err != nil {
		return errorsmod.Wrap(err, "slash")
	}

	k.SetSequencer(ctx, seq)
	return nil
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

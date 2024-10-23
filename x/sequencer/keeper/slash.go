package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	seq, err := k.LivenessLiableSequencer(ctx, rollappID)
	if err != nil {
		return err
	}
	if
	mul := k.GetParams(ctx).LivenessSlashMinMultiplier
	abs := k.GetParams(ctx).LivenessSlashMinAbsolute
	tokens := seq.TokensCoin()
	tokensMul := ucoin.MulDec(mul, tokens)
	amt := ucoin.SimpleMin(tokens, ucoin.SimpleMax(abs, tokensMul[0]))
	// TODO: make sure to be correct wrt. min bond, see https://github.com/dymensionxyz/dymension/issues/1019
	return errorsmod.Wrap(k.slash(ctx, seq, amt, sdk.ZeroDec(), nil), "slash")
}

func (k Keeper) HandleFraud(ctx sdk.Context, seq types.Sequencer, rewardee *sdk.AccAddress) error {
	var err error
	if rewardee != nil {
		rewardMul := sdk.MustNewDecFromStr("0.5")
		err = k.slash(ctx, seq, seq.TokensCoin(), rewardMul, *rewardee)
	} else {
		err = k.slash(ctx, seq, seq.TokensCoin(), sdk.ZeroDec(), nil)
	}
	if err != nil {
		return errorsmod.Wrap(err, "slash")
	}
	return errorsmod.Wrap(k.unbond(ctx, seq), "unbond")
}

func (k Keeper) slash(ctx sdk.Context, seq types.Sequencer, amt sdk.Coin, rewardMul sdk.Dec, rewardee sdk.AccAddress) error {
	seq.Tokens = seq.Tokens.Sub(amt)
	rewardCoin := ucoin.MulDec(rewardMul, amt)[0]
	if !rewardCoin.IsZero() {
		_ = rewardCoin // TODO: send to rewardee
	}
	return nil
}

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	seq := k.GetProposer(ctx, rollappID)
	mul := k.GetParams(ctx).LivenessSlashMinMultiplier
	abs := k.GetParams(ctx).LivenessSlashMinAbsolute
	tokens := seq.TokensCoin()
	tokensMul := ucoin.MulDec(mul, tokens)
	amt := ucoin.SimpleMin(tokens, ucoin.SimpleMax(abs, tokensMul[0]))
	err := errorsmod.Wrap(k.slash(ctx, seq, amt, sdk.ZeroDec(), nil), "slash")
	k.SetSequencer(ctx, seq)
	return err
}

func (k Keeper) HandleFraud(ctx sdk.Context, seq types.Sequencer, rewardee *sdk.AccAddress) error {
	var err error
	if rewardee != nil {
		rewardMul := sdk.MustNewDecFromStr("0.5") // TODO: parameterise
		err = k.slash(ctx, seq, seq.TokensCoin(), rewardMul, *rewardee)
	} else {
		err = k.slash(ctx, seq, seq.TokensCoin(), sdk.ZeroDec(), nil)
	}
	if err != nil {
		return errorsmod.Wrap(err, "slash")
	}
	err = errorsmod.Wrap(k.unbond(ctx, &seq), "unbond")
	k.SetSequencer(ctx, seq)
	k.optOutAllSequencers(ctx, seq.RollappId)
	return err
}

func (k Keeper) slash(ctx sdk.Context, seq types.Sequencer, amt sdk.Coin, rewardMul sdk.Dec, rewardee sdk.AccAddress) error {
	rewardCoin := ucoin.MulDec(rewardMul, amt)[0]
	if !rewardCoin.IsZero() {
		err := k.sendFromModule(ctx, &seq, rewardCoin, rewardee)
		if err != nil {
			return errorsmod.Wrap(err, "send")
		}
	}
	remainder := amt.Sub(rewardCoin)
	return errorsmod.Wrap(k.burn(ctx, &seq, remainder), "burn")
}

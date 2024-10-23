package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	seq, err := k.LivenessLiableSequencer(ctx, rollappID)
	if err != nil {
		return err
	}
	mul := k.GetParams(ctx).LivenessSlashMinMultiplier
	abs := k.GetParams(ctx).LivenessSlashMinAbsolute
	tokens := seq.Tokens
	tokensMul := ucoin.MulDec(mul, tokens...)
	amt :=
	// TODO: make sure to be correct wrt. min bond, see https://github.com/dymensionxyz/dymension/issues/1019
	return k.slash(ctx, &seq, amt)
}

func (k Keeper) HandleFraud(ctx sdk.Context, seq types.Sequencer, rewardee *sdk.AccAddress) error {
}

func (k Keeper) slash(ctx sdk.Context, seq types.Sequencer, amt sdk.Coin, rewardMul sdk.Dec, rewardee *sdk.AccAddress) error {
}

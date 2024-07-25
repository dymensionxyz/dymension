package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)


func MulCoinsDec(coins sdk.Coins, dec sdk.Dec ) sdk.Coins {
	return sdk.Coins{}
}

func (k Keeper) LivenessSlashAndJail(ctx sdk.Context, args types.LivenessSlashAndJailArgs, burnMultiplier sdk.Dec, recipients ...types.LivenessSlashAndJailFundsRecipient) (types.LivenessSlashAndJailResult, error) {

	args := types.LivenessSlashAndJailArgs{
		HHub:                      0,
		HNoticeExpired:            nil,
		HUpdate:                   0,
		HubBlockTime:              0,
		SlashTimeNoUpdate:         0,
		SlashTimeNoTerminalUpdate: 0,
		SlashInterval:             0,
		SlashMultiplier:           sdk.Dec{},
		JailTime:                  0,
		Balance:                   nil,
		MinBond:                   nil,
	}

	slashAmt, jail := args.Calculate()

	burnAmt := MulCoinsDec(slashAmt, burnMultiplier)
	err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnAmt)
	if err != nil {
		return types.LivenessSlashAndJailResult{}, err // TODO:
	}

	for _, r := range recipients {
		sendAmt := MulCoinsDec(slashAmt, r.Multiplier)
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, r.Addr, sendAmt)
		if err!=nil{
			return types.LivenessSlashAndJailResult{}, err // TODO:
		}



		newCoins := slashAmt.New
		sdk.Coins{}
		amt :=r.Multiplier.MulInt(slashAmt.)

	}
}

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string, rewardAddr sdk.AccAddress) (SlashLivenessResult, error)  {
	p := k.GetParams(ctx).Liveness()
	slashAmt, jail := LivenessSlashAndJail(
		LivenessSlashAndJailArgs{
			ctx.BlockHeight(),
			0,
			0,
			p.HubExpectedBlockTime,
			p.SlashTime,
			time.Nanosecond, // TODO:
			p.SlashInterval,
			p.SlashMultiplier,
			p.JailTime,
			sdk.Coins{},
			sdk.Coins{}, // TODO:
		},
	)

	k.bankKeeper.SendCoins(ctx, )
	rewardee :=
}



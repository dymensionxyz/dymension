package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type SlashLivenessResult struct {
	slashed sdk.Coins
	jailed bool
	timeUntilNextSlashPossible time.Time
	rewarded sdk.Coins
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

type LivenessSlashAndJailArgs struct {
	HHub            int64
	HNoticeExpired *int64
	HUpdate         int64
	HubBlockTime    time.Duration
	SlashTimeNoUpdate       time.Duration
	SlashTimeNoTerminalUpdate time.Duration
	SlashInterval   time.Duration
	SlashMultiplier sdk.Dec
	JailTime        time.Duration
	Balance         sdk.Coins
	MinBond         sdk.Coins
}

func LivenessSlashAndJail(args LivenessSlashAndJailArgs) (slashAmt sdk.Coins, jail bool) {
}

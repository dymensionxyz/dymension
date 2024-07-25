package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) bool {
	p := k.GetParams(ctx).Liveness()
	slashAmt, jail := LivenessSlashAndJail(
		ctx.BlockHeight(),
		0,
		p.HubExpectedBlockTime,
		p.SlashTime,
		p.SlashInterval,
		p.SlashMultiplier,
		p.JailTime,
		sdk.Coins{},
		sdk.Coins{}, // TODO:
	)
}

func LivenessSlashAndJail(
	hHub int64,
	hUpdate int64,
	hubBlockTime time.Duration,
	slashTime time.Duration,
	slashInterval time.Duration,
	slashMultiplier sdk.Dec,
	jailTime time.Duration,
	balance sdk.Coins,
	minBond sdk.Coins,
) (slashAmt sdk.Coins, jail bool) {
}

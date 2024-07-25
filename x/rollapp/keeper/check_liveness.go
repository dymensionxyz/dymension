package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SlashLiveness(ctx sdk.Context, rollappID string) bool {
	k.GetParams(ctx)
	slashAmt, jail := LivenessSlashAndJail(
		ctx.BlockHeight(),
		0,
		time.Second,
		time.Second,
		time.Second,
		sdk.Dec{},
		time.Second,
		sdk.Coins{},
		sdk.Coins{},
	)
}

func LivenessSlashAndJail(
	hHub int64,
	hUpdate int64,
	hubBlockTime time.Duration,
	slashTime time.Duration,
	slashInterval time.Duration,
	slashFactor sdk.Dec,
	jailTime time.Duration,
	balance sdk.Coins,
	minBond sdk.Coins,
) (slashAmt sdk.Coins, jail bool) {
}

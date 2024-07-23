package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
def liveness_slash(hubBlockTime, slashTime, slashInterval, slashFactor, jailTime, hubHeight, height_last_update,
                   balance):
    down_time = (hubHeight - height_last_update) * hubBlockTime
    if down_time < slashTime:
        return 0, False
    if down_time < jailTime:
        last_down_time = (hubHeight - 1 - height_last_update) * hubBlockTime
        if last_down_time // slashInterval < down_time // slashInterval:
            return slashFactor * balance, False
        return 0, False
    return 0, True
*/

type SlashAndJail func(
	hubBlockTime time.Duration,
	slashTime time.Duration,
	slashInterval time.Duration,
	slashFactor sdk.Dec,
	jailTime time.Duration,
	hHub uint64,
	hUpdate uint64,
	balance sdk.Coins,
) (slashAmt sdk.Coins, jailed bool)

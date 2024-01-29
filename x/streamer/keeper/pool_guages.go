package keeper

import (
	"time"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) CreatePoolGauge(ctx sdk.Context, poolId uint64) error {
	for _, duration := range k.ik.GetLockableDurations(ctx) {
		_, err := k.ik.CreateGauge(
			ctx,
			true,
			k.ak.GetModuleAddress(types.ModuleName),
			sdk.Coins{},
			lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         gammtypes.GetPoolShareDenom(poolId),
				Duration:      duration,
				Timestamp:     time.Time{},
			},
			ctx.BlockTime(),
			1,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// TODO: move to incentives module

func (k Keeper) CreatePoolGauge(ctx sdk.Context, poolId uint64) error {
	_, err := k.ik.CreateAssetGauge(
		ctx,
		true,
		k.ak.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		lockuptypes.QueryCondition{
			Denom:    gammtypes.GetPoolShareDenom(poolId),
			LockAge:  k.ik.GetParams(ctx).MinLockAge,
			Duration: k.ik.GetParams(ctx).MinLockDuration,
		},
		ctx.BlockTime(),
		1,
	)
	return err
}

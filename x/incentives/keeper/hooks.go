package keeper

import (
	"fmt"
	"time"

	ctypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	streamermoduletypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

/* -------------------------------------------------------------------------- */
/*                                epoch hooks                                 */
/* -------------------------------------------------------------------------- */

var _ epochstypes.EpochHooks = EpochHooks{}

type EpochHooks struct {
	Keeper
}

func (k Keeper) EpochHooks() EpochHooks {
	return EpochHooks{k}
}

// BeforeEpochStart is the epoch start hook.
func (k EpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (k EpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	params := k.GetParams(ctx)
	if epochIdentifier == params.DistrEpochIdentifier {
		// begin distribution if it's start time
		gauges := k.GetUpcomingGauges(ctx)
		for _, gauge := range gauges {
			if !ctx.BlockTime().Before(gauge.StartTime) {
				if err := k.moveUpcomingGaugeToActiveGauge(ctx, gauge); err != nil {
					return err
				}
			}
		}

		// distribute due to epoch event
		gauges = k.GetActiveGauges(ctx)
		_, err := k.DistributeOnEpochEnd(ctx, gauges)
		if err != nil {
			return err
		}
	}
	return nil
}

/* -------------------------------------------------------------------------- */
/*                                 pool hooks                                 */
/* -------------------------------------------------------------------------- */

var _ gammtypes.GammHooks = PoolHooks{}

type PoolHooks struct {
	ctypes.StubGammHooks
	Keeper
}

func (k Keeper) PoolHooks() PoolHooks {
	return PoolHooks{ctypes.StubGammHooks{}, k}
}

// AfterPoolCreated creates a gauge for each pool’s lockable duration.
func (h PoolHooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for _, duration := range h.GetLockableDurations(ctx) {
		_, err := h.CreateAssetGauge(
			ctx,
			true,
			// historically, x/streamer is the owner of rollapp gauges
			h.ak.GetModuleAddress(streamermoduletypes.ModuleName),
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
			ctx.Logger().Error("Failed to create pool gauge", "duration", duration, "error", err)
		}
	}
}

/* -------------------------------------------------------------------------- */
/*                                rollapp hooks                               */
/* -------------------------------------------------------------------------- */

var _ rollapptypes.RollappHooks = RollappHooks{}

type RollappHooks struct {
	rollapptypes.StubRollappCreatedHooks
	Keeper
}

func (k Keeper) RollappHooks() RollappHooks {
	return RollappHooks{rollapptypes.StubRollappCreatedHooks{}, k}
}

// RollappCreated implements types.RollappHooks.
func (h RollappHooks) RollappCreated(ctx sdk.Context, rollappID, _ string, _ sdk.AccAddress) error {
	_, err := h.CreateRollappGauge(ctx, rollappID)
	if err != nil {
		ctx.Logger().Error("Failed to create rollapp gauge", "error", err)
		return fmt.Errorf("create rollapp gauge: %w", err)
	}
	return nil
}

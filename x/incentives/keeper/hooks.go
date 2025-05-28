package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
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

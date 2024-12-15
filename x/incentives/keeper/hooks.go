package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeforeEpochStart is the epoch start hook.
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	params := k.GetParams(ctx)
	if epochIdentifier == params.DistrEpochIdentifier {
		// activate gauges that have started
		gauges := k.GetUpcomingGauges(ctx)
		for _, gauge := range gauges {
			if !ctx.BlockTime().Before(gauge.StartTime) {
				if err := k.moveUpcomingGaugeToActiveGauge(ctx, gauge); err != nil {
					return err
				}
			}
		}

		// get active, non-perpetual gauges
		// perpetual gauges are paid directly by the protocol (i.e streamer)
		newGauges := make([]types.Gauge, 0, len(gauges))
		for _, gauge := range k.GetActiveGauges(ctx) {
			if !gauge.IsPerpetual {
				newGauges = append(newGauges, gauge)
			}
		}
		gauges = newGauges

		// distribute due to epoch event
		_, err := k.DistributeOnEpochEnd(ctx, gauges)
		if err != nil {
			return err
		}
	}
	return nil
}

// ___________________________________________________________________________________________________

// Hooks is the wrapper struct for the incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Hooks returns the hook wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// BeforeEpochStart is the epoch start hook.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

// AfterEpochEnd is the epoch end hook.
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

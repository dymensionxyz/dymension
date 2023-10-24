package keeper

import (
	"github.com/dymensionxyz/dymension/x/streamer/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

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
		// begin distribution if it's start time
		streams := k.GetUpcomingStreams(ctx)
		for _, stream := range streams {
			if !ctx.BlockTime().Before(stream.StartTime) {
				if err := k.moveUpcomingStreamToActiveStream(ctx, stream); err != nil {
					return err
				}
			}
		}

		// if len(streams) > 10 {
		// 	ctx.EventManager().IncreaseCapacity(2e6)
		// }

		// distribute due to epoch event
		streams = k.GetActiveStreams(ctx)
		// only distribute to active streams that are for native denoms
		// or non-perpetual and for synthetic denoms.
		// We distribute to perpetual synthetic denoms elsewhere in superfluid.
		distrStreams := []types.Stream{}
		for _, stream := range streams {
			isSynthetic := lockuptypes.IsSyntheticDenom(stream.DistributeTo.Denom)
			if !(isSynthetic && stream.IsPerpetual) {
				distrStreams = append(distrStreams, stream)
			}
		}
		_, err := k.Distribute(ctx, distrStreams)
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

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	ctypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// Hooks is the wrapper struct for the streamer keeper.
type Hooks struct {
	ctypes.StubGammHooks
	rollapptypes.StubRollappCreatedHooks
	k Keeper
}

var (
	_ epochstypes.EpochHooks    = Hooks{}
	_ gammtypes.GammHooks       = Hooks{}
	_ rollapptypes.RollappHooks = Hooks{}
)

// Hooks returns the hook wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k: k}
}

/* -------------------------------------------------------------------------- */
/*                                 epoch hooks                                */
/* -------------------------------------------------------------------------- */

// BeforeEpochStart updates the streams based on a new epoch and emits an event.
// It moves upcoming streams to active if the start time has been reached.
// It updates active streams with respect to the new epoch and saves them.
// Finally, it emits an event with the number of active streams.
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string) error {
	// Move upcoming streams to active if start time reached
	upcomingStreams := k.GetUpcomingStreams(ctx)
	for _, s := range upcomingStreams {
		if !ctx.BlockTime().Before(s.StartTime) {
			err := k.moveUpcomingStreamToActiveStream(ctx, s)
			if err != nil {
				return fmt.Errorf("move upcoming stream to active stream: %w", err)
			}
		}
	}

	toStart := k.GetActiveStreamsForEpoch(ctx, epochIdentifier)

	// Update streams with respect to a new epoch and save them
	for _, s := range toStart {
		updated, err := k.UpdateStreamAtEpochStart(ctx, s)
		if err != nil {
			return fmt.Errorf("update stream '%d' at epoch start: %w", s.Id, err)
		}
		// Save the stream
		err = k.SetStream(ctx, &updated)
		if err != nil {
			return fmt.Errorf("set stream: %w", err)
		}
	}

	err := uevent.EmitTypedEvent(ctx, &types.EventEpochStart{
		ActiveStreamsNum: uint64(len(toStart)),
	})
	if err != nil {
		return fmt.Errorf("emit typed event: %w", err)
	}

	return nil
}

// AfterEpochEnd distributes rewards, updates streams, and saves the changes to the state after the epoch end.
// It distributes rewards to streams that have the specified epoch identifier or aborts if there are no streams
// in this epoch. After the distribution, it resets the epoch pointer to the very fist gauge.
// The method uses three caches:
//   - Stream cache for updating stream distributed coins
//   - Gauge cache for updating gauge coins
//   - Number of locks per denom to reduce the number of requests for x/lockup
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string) (sdk.Coins, error) {
	toDistribute := k.GetActiveStreamsForEpoch(ctx, epochIdentifier)

	if len(toDistribute) == 0 {
		// Nothing to distribute
		return sdk.Coins{}, nil
	}

	epochPointer, err := k.GetEpochPointer(ctx, epochIdentifier)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("get epoch pointer for epoch '%s': %w", epochIdentifier, err)
	}

	// Init helper caches
	streamCache := newStreamInfo(toDistribute)
	gaugeCache := newGaugeInfo()

	// Cache specific for asset gauges. Helps reduce the number of x/lockup requests.
	denomLockCache := incentivestypes.NewDenomLocksCache()

	// Calculate rewards and fill caches
	distrResult := k.CalculateRewards(ctx, epochPointer, types.IterationsNoLimit, streamCache, gaugeCache, denomLockCache)

	// Send coins to distribute to the x/incentives module
	err = k.bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, incentivestypes.ModuleName, streamCache.totalDistr)
	if err != nil {
		return nil, fmt.Errorf("send coins: %w", err)
	}

	// Distribute the rewards
	const EpochEnd = true
	_, err = k.ik.Distribute(ctx, gaugeCache.getGauges(), denomLockCache, EpochEnd)
	if err != nil {
		return nil, fmt.Errorf("distribute: %w", err)
	}

	// Update streams with respect to a new epoch and save them
	for _, s := range streamCache.getStreams() {
		updated, err := k.UpdateStreamAtEpochEnd(ctx, s)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("update stream '%d' at epoch start: %w", s.Id, err)
		}
		// Save the stream
		err = k.SetStream(ctx, &updated)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("set stream: %w", err)
		}
	}

	// Reset the epoch pointer
	distrResult.NewPointer.SetToFirstGauge()
	err = k.SaveEpochPointer(ctx, distrResult.NewPointer)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("save epoch pointer: %w", err)
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventEpochEnd{
		Iterations:  distrResult.Iterations,
		Distributed: streamCache.totalDistr,
	})
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("emit typed event: %w", err)
	}

	ctx.Logger().Info("Streamer distributed coins", "amount", streamCache.totalDistr.String())

	return streamCache.totalDistr, nil
}

// BeforeEpochStart is the epoch start hook.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, _ int64) error {
	err := h.k.BeforeEpochStart(ctx, epochIdentifier)
	if err != nil {
		return fmt.Errorf("x/streamer: before epoch '%s' start: %w", epochIdentifier, err)
	}
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) error {
	_, err := h.k.AfterEpochEnd(ctx, epochIdentifier)
	if err != nil {
		return fmt.Errorf("x/streamer: after epoch '%s' end: %w", epochIdentifier, err)
	}
	return nil
}

/* -------------------------------------------------------------------------- */
/*                                 pool hooks                                 */
/* -------------------------------------------------------------------------- */

// AfterPoolCreated creates a gauge for each pool’s lockable duration.
func (h Hooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	err := h.k.CreatePoolGauge(ctx, poolId)
	if err != nil {
		ctx.Logger().Error("Failed to create pool gauge", "error", err)
	}
}

/* -------------------------------------------------------------------------- */
/*                                rollapp hooks                               */
/* -------------------------------------------------------------------------- */
// AfterStateFinalized implements types.RollappHooks.

// RollappCreated implements types.RollappHooks.
func (h Hooks) RollappCreated(ctx sdk.Context, rollappID, _ string, _ sdk.AccAddress) error {
	err := h.k.CreateRollappGauge(ctx, rollappID)
	if err != nil {
		ctx.Logger().Error("Failed to create rollapp gauge", "error", err)
		return err
	}
	return nil
}

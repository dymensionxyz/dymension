package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	ctypes "github.com/dymensionxyz/dymension/v3/x/common/types"
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
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string) (sdk.Coins, error) {
	// Get active streams
	activeStreams := k.GetActiveStreamsForEpoch(ctx, epochIdentifier)
	if len(activeStreams) == 0 {
		// Nothing to distribute
		return sdk.Coins{}, nil
	}

	// Get epoch pointer for the current epoch
	epochPointer, err := k.GetEpochPointer(ctx, epochIdentifier)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("get epoch pointer for epoch '%s': %w", epochIdentifier, err)
	}

	// Distribute rewards
	const epochEnd = true
	coins, iterations, err := k.Distribute(ctx, []types.EpochPointer{epochPointer}, activeStreams, types.IterationsNoLimit, epochEnd)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("distribute: %w", err)
	}

	// Reset the epoch pointer
	epochPointer.SetToFirstGauge()
	err = k.SaveEpochPointer(ctx, epochPointer)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("save epoch pointer: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventEpochEnd{
		Iterations:  iterations,
		Distributed: coins,
	})
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("emit typed event: %w", err)
	}

	ctx.Logger().Info("Streamer distributed coins", "amount", coins.String())

	return coins, nil
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

package keeper

import (
	"fmt"

	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	ctypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

// BeforeEpochStart is the epoch start hook.
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string) (sdk.Coins, error) {
	activeStreams := k.GetActiveStreams(ctx)

	// Move upcoming streams to active if start time reached
	upcomingStreams := k.GetUpcomingStreams(ctx)
	var newlyActivated []types.Stream
	for _, s := range upcomingStreams {
		if !ctx.BlockTime().Before(s.StartTime) {
			err := k.moveUpcomingStreamToActiveStream(ctx, s)
			if err != nil {
				return sdk.Coins{}, fmt.Errorf("move upcoming stream to active stream: %w", err)
			}

			if epochIdentifier == s.DistrEpochIdentifier {
				newlyActivated = append(newlyActivated, s)
			}
		}
	}

	if len(activeStreams) == 0 && len(newlyActivated) == 0 {
		// Nothing to distribute
		return sdk.Coins{}, nil
	}

	epochPointer, err := k.GetEpochPointer(ctx, epochIdentifier)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("get epoch pointer for epoch '%s': %w", epochIdentifier, err)
	}

	distrResult := k.DistributeRewards(ctx, epochPointer, types.IterationsNoLimit, activeStreams)

	// Update streams with respect to a new epoch and save them
	for _, s := range append(distrResult.FilledStreams, newlyActivated...) {
		updated, err := k.UpdateStreamAtEpochStart(ctx, s)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("update stream '%d' at epoch start: %w", s.Id, err)
		}
		// Save the stream
		err = k.SetStream(ctx, &updated)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("set stream: %w", err)
		}
	}

	distrResult.NewPointer.SetToFirstGauge()
	err = k.SaveEpochPointer(ctx, distrResult.NewPointer)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("save epoch pointer: %w", err)
	}

	ctx.Logger().Info("Streamer distributed coins", "amount", distrResult.DistributedCoins.String())

	return distrResult.DistributedCoins, nil
}

// BeforeEpochStart is the epoch start hook.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

// AfterEpochEnd is the epoch end hook.
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) error {
	_, err := h.k.AfterEpochEnd(ctx, epochIdentifier)
	if err != nil {
		return err
	}
	// TODO: add event
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

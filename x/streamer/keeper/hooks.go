package keeper

import (
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
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	streams := k.GetUpcomingStreams(ctx)
	// move to active if start time reached
	for _, stream := range streams {
		if !ctx.BlockTime().Before(stream.StartTime) {
			if err := k.moveUpcomingStreamToActiveStream(ctx, stream); err != nil {
				return err
			}
		}
	}

	// distribute due to epoch event
	streams = k.GetActiveStreams(ctx)
	distrStreams := []types.Stream{}
	for _, stream := range streams {
		// begin distribution if it's correct epoch
		if epochIdentifier != stream.DistrEpochIdentifier {
			continue
		}
		distrStreams = append(distrStreams, stream)
	}

	if len(distrStreams) == 0 {
		return nil
	}

	distributedAmt, err := k.Distribute(ctx, distrStreams)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Streamer distributed coins", "amount", distributedAmt.String())
	return nil
}

// BeforeEpochStart is the epoch start hook.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

// AfterEpochEnd is the epoch end hook.
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
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

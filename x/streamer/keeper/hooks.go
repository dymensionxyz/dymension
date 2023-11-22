package keeper

import (
	"github.com/dymensionxyz/dymension/x/streamer/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// ___________________________________________________________________________________________________

// Hooks is the wrapper struct for the streamer keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}
var _ gammtypes.GammHooks = Hooks{}

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

// AfterPoolCreated creates a gauge for each pool’s lockable duration.
func (h Hooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	err := h.k.CreatePoolGauge(ctx, poolId)
	if err != nil {
		ctx.Logger().Error("Failed to create pool gauge", "error", err)
	}
}

// AfterJoinPool hook is a noop.
func (h Hooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
}

// AfterExitPool hook is a noop.
func (h Hooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
}

// AfterSwap hook is a noop.
func (h Hooks) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
}

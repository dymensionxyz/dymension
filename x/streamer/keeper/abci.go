package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// EndBlock iterates over the epoch pointers, calculates rewards, distributes them, and updates the streams.
func (k Keeper) EndBlock(ctx sdk.Context) error {
	epochPointers, err := k.GetAllEpochPointers(ctx)
	if err != nil {
		return fmt.Errorf("get all epoch pointers: %w", err)
	}

	streams := k.GetActiveStreams(ctx)
	maxIterations := k.GetParams(ctx).MaxIterationsPerBlock

	const nonEpochEnd = false
	coins, iterations, err := k.Distribute(ctx, epochPointers, streams, maxIterations, nonEpochEnd)
	if err != nil {
		return fmt.Errorf("distribute: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventEndBlock{
		Iterations:    iterations,
		MaxIterations: maxIterations,
		Distributed:   coins,
	})
	if err != nil {
		return fmt.Errorf("emit typed event: %w", err)
	}

	return nil
}

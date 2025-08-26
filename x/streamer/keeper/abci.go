package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// BeginBlock processes pump streams for potential pump operations
func (k Keeper) BeginBlock(ctx sdk.Context) error {
	// Get all active streams
	streams := k.GetActiveStreams(ctx)
	
	// Filter pump streams
	var pumpStreams []types.Stream
	for _, stream := range streams {
		if stream.PumpParams != nil {
			pumpStreams = append(pumpStreams, stream)
		}
	}
	
	// Process pump streams
	if len(pumpStreams) > 0 {
		err := k.DistributePumpStreams(ctx, pumpStreams)
		if err != nil {
			return fmt.Errorf("failed to distribute pump streams: %w", err)
		}
	}
	
	return nil
}

// EndBlock iterates over the epoch pointers, calculates rewards, distributes them, and updates the streams.
func (k Keeper) EndBlock(ctx sdk.Context) error {
	epochPointers, err := k.GetAllEpochPointers(ctx)
	if err != nil {
		return fmt.Errorf("get all epoch pointers: %w", err)
	}

	streams := k.GetActiveStreams(ctx)
	maxIterations := k.GetParams(ctx).MaxIterationsPerBlock

	const epochEnd = false
	coins, iterations, err := k.Distribute(ctx, epochPointers, streams, maxIterations, epochEnd)
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

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/internal/pagination"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	// Get all active streams
	streams := k.GetActiveStreams(ctx)

	// Temporary map for convenient calculations
	streamMap := make(map[uint64]types.Stream, len(streams))
	for _, stream := range streams {
		streamMap[stream.Id] = stream
	}

	epochPointers, err := k.GetAllEpochPointers(ctx)
	if err != nil {
		return fmt.Errorf("get all epoch pointers: %w", err)
	}

	maxIterations := k.GetParams(ctx).MaxIterationsPerBlock
	totalIterations := uint64(0)

	for _, p := range epochPointers {
		remainIterations := maxIterations - totalIterations

		if remainIterations <= 0 {
			// no more iterations available for this block
			break
		}

		newPointer, iters := IterateEpochPointer(p, streams, remainIterations, func(v StreamGauge) pagination.Stop {
			distributed, errX := k.DistributeToGauge(ctx, v.Stream.EpochCoins, v.Gauge, v.Stream.DistributeTo.TotalWeight)
			if errX != nil {
				// Ignore this gauge
				k.Logger(ctx).
					With("streamID", v.Stream.Id, "gaugeID", v.Gauge.GaugeId, "error", errX.Error()).
					Error("Failed to distribute to gauge")
			}

			// Update distributed coins for the stream
			stream := streamMap[v.Stream.Id]
			stream.DistributedCoins = stream.DistributedCoins.Add(distributed...)
			streamMap[v.Stream.Id] = stream

			return pagination.Continue
		})

		err = k.SaveEpochPointer(ctx, newPointer)
		if err != nil {
			return fmt.Errorf("save epoch pointer: %w", err)
		}
		totalIterations += iters
	}

	// Save stream updates
	for _, stream := range streamMap {
		err := k.setStream(ctx, &stream)
		if err != nil {
			return fmt.Errorf("set stream: %w", err)
		}
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventEndBlock{
		Iterations:    totalIterations,
		MaxIterations: maxIterations,
	})
	if err != nil {
		return fmt.Errorf("emit typed event: %w", err)
	}

	return nil
}

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	streams := k.GetActiveStreams(ctx)

	epochPointers, err := k.GetAllEpochPointers(ctx)
	if err != nil {
		return fmt.Errorf("get all epoch pointers: %w", err)
	}

	maxIterations := k.GetParams(ctx).MaxIterationsPerBlock
	totalIterations := uint64(0)
	totalDistributed := sdk.NewCoins()

	for _, p := range epochPointers {
		remainIterations := maxIterations - totalIterations

		if remainIterations <= 0 {
			break // no more iterations available for this block
		}

		result := k.DistributeRewards(ctx, p, remainIterations, streams)

		totalIterations += result.Iterations
		totalDistributed = totalDistributed.Add(result.DistributedCoins...)
		streams = result.FilledStreams

		err = k.SaveEpochPointer(ctx, result.NewPointer)
		if err != nil {
			return fmt.Errorf("save epoch pointer: %w", err)
		}
	}

	// Save stream updates
	for _, stream := range streams {
		err = k.SetStream(ctx, &stream)
		if err != nil {
			return fmt.Errorf("set stream: %w", err)
		}
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventEndBlock{
		Iterations:    totalIterations,
		MaxIterations: maxIterations,
		Distributed:   totalDistributed,
	})
	if err != nil {
		return fmt.Errorf("emit typed event: %w", err)
	}

	return nil
}

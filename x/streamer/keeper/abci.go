package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	streams := k.GetActiveStreams(ctx)

	epochPointers, err := k.GetAllEpochPointers(ctx)
	if err != nil {
		return fmt.Errorf("get all epoch pointers: %w", err)
	}

	// Sort epoch pointers to distribute to shorter epochs first
	types.SortEpochPointers(epochPointers)

	maxIterations := k.GetParams(ctx).MaxIterationsPerBlock
	totalIterations := uint64(0)

	streamCache := newStreamInfo(streams)
	gaugeCache := newGaugeInfo()

	for _, p := range epochPointers {
		remainIterations := maxIterations - totalIterations

		if remainIterations <= 0 {
			break // no more iterations available for this block
		}

		result := k.CalculateRewards(ctx, p, remainIterations, streamCache, gaugeCache)

		totalIterations += result.Iterations

		err = k.SaveEpochPointer(ctx, result.NewPointer)
		if err != nil {
			return fmt.Errorf("save epoch pointer: %w", err)
		}
	}

	// Filter gauges to distribute
	toDistribute := k.filterGauges(ctx, gaugeCache)

	// Send coins to distribute to the x/incentives module
	err = k.bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, incentivestypes.ModuleName, streamCache.totalDistr)
	if err != nil {
		return fmt.Errorf("send coins: %w", err)
	}

	// Distribute the rewards
	_, err = k.ik.Distribute(ctx, toDistribute)
	if err != nil {
		return fmt.Errorf("distribute: %w", err)
	}

	// Save stream updates
	for _, stream := range streamCache.getStreams() {
		err = k.SetStream(ctx, &stream)
		if err != nil {
			return fmt.Errorf("set stream: %w", err)
		}
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventEndBlock{
		Iterations:    totalIterations,
		MaxIterations: maxIterations,
		Distributed:   streamCache.totalDistr,
	})
	if err != nil {
		return fmt.Errorf("emit typed event: %w", err)
	}

	return nil
}

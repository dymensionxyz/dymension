package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) {
	epochPointers := k.GetEpochPointers
}

type EpochPointer struct {
	epochRewards sdk.Coins
	epochID      string
	streamID     uint64
	gaugeID      uint64
}

// Notes:
// GetStreamsFromID - ??
func (k Keeper) processEpochPointer(ctx sdk.Context, p EpochPointer) error {
	nextP, found := k.nextValidEpochPointer(ctx, p)
	if !found {
		return nil // no more gauges to iterate
	}
	if nextP.streamID == p.streamID {
		// the current stream is valid, so we adjust the gauge reference
		nextP.gaugeID = p.gaugeID
	}

	// now nextP points to the valid gauge
}

func (k Keeper) fillIterations(ctx sdk.Context, iterationsLimit int, p EpochPointer) (sdk.Coins, int, error) {
	if iterationsLimit <= 0 {
		// TODO
	}

	totalDistributed := sdk.NewCoins()
	totalIterations := 0

	// less than 10 streams every epoch
	//streams := k.GetStreamsStartingFromID(ctx, p.streamID)
	var streams []types.Stream
	for _, stream := range streams {
		totalWeight := sdk.NewDecFromInt(stream.DistributeTo.TotalWeight)
		for _, gauge := range stream.DistributeTo.Records {
			if totalIterations > iterationsLimit {
				break
			}

			if gauge.GaugeId < p.gaugeID {
				continue // TODO: do we need to count this "empty" iterations?
			}

			distributed, err := k.DistributeToGauge(ctx, p.epochRewards, gauge, totalWeight)
			if err != nil {
				return sdk.Coins{}, 0, err
			}

			totalDistributed = totalDistributed.Add(distributed...)
			totalIterations++
		}
	}

	// TODO: update every stream

	return totalDistributed, totalIterations, nil
}

// if the initial epoch pointer is valid, return it. note that it might have non-first gauge ID
// otherwise, return the first valid stream with the first gauge ID
// return false if no more streams are presented
// now x/streamer support only one empty distribution: from x/sponsorship, fix it?
func (k Keeper) nextValidEpochPointer(ctx sdk.Context, p EpochPointer) (EpochPointer, bool) {
	const Break = true
	const Continue = false
	var found = false

	// iterate streams starting including the start position
	k.IterateStreamsFromID(ctx, p.streamID, func(s types.Stream) bool {
		if s.DistrEpochIdentifier == p.epochID && len(s.DistributeTo.Records) != 0 {
			found = true
			p = EpochPointer{
				epochID:  p.epochID,
				streamID: s.Id,
				gaugeID:  s.DistributeTo.Records[0].GaugeId,
			}
			return Break
		}
		return Continue
	})

	return p, found
}

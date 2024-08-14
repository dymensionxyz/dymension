package keeper

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) {
	//epochPointers := k.GetEpochPointers
	var epochPointers []types.EpochPointer

	// less than 10 streams every epoch
	streams := k.GetActiveStreams(ctx)

	//iterationsPerBlock := k.GetParams(ctx)
	maxPerBlock := 500
	iterationsPerBlock := 0

	totalDistributed := sdk.NewCoins()

	for _, pointer := range epochPointers {
		remainIterations := maxPerBlock - iterationsPerBlock

		if remainIterations <= 0 {
			// no more iterations available for this block
			break
		}

		result, err := k.ProcessEpochPointer(ctx, remainIterations, pointer, streams)
		if err != nil {
			panic(err) // ?? or Logger.Error
		}

		totalDistributed = totalDistributed.Add(result.Distributed...)
		iterationsPerBlock += result.Iterations
	}

	// TODO event
}

type ProcessEpochPointerResult struct {
	Distributed sdk.Coins
	Iterations  int
	Pointer     types.EpochPointer
}

func (k Keeper) ProcessEpochPointer(
	ctx sdk.Context,
	iterationsLimit int,
	p types.EpochPointer,
	streams []types.Stream,
) (ProcessEpochPointerResult, error) {
	totalDistributed := sdk.NewCoins()

	iter := NewStreamIterator(streams, p.StreamId, p.GaugeId, p.EpochIdentifier)
	for ; iter.Iterations() <= iterationsLimit && iter.Valid(); iter.Next() {
		stream := iter.Stream()
		gauge := iter.Gauge()

		distributed, err := k.DistributeToGauge(ctx, p.EpochCoins, gauge, stream.DistributeTo.TotalWeight)
		if err != nil {
			// TODO: don't return error??
			return ProcessEpochPointerResult{}, err
		}

		totalDistributed = totalDistributed.Add(distributed...)

		p.StreamId = stream.Id
		p.GaugeId = gauge.GaugeId
	}

	return ProcessEpochPointerResult{
		Distributed: totalDistributed,
		Iterations:  iter.Iterations(),
		Pointer:     p,
	}, nil
}

type StreamIterator struct {
	data            []types.Stream
	streamIdx       int
	gaugeIdx        int
	epochIdentifier string
	iterations      int
}

func NewStreamIterator(data []types.Stream, startStreamID uint64, startGaugeID uint64, epochIdentifier string) StreamIterator {
	// streamIdx is the position where the stream is found, or the position where it would appear in the sort order
	streamIdx, _ := slices.BinarySearchFunc(data, startStreamID, func(stream types.Stream, targetID uint64) int {
		return int(stream.Id - targetID)
	})

	// streamIdx is the position where the gauge is found, or the position where it would appear in the sort order
	gaugeIdx, _ := slices.BinarySearchFunc(data[streamIdx].DistributeTo.Records, startGaugeID, func(record types.DistrRecord, targetID uint64) int {
		return int(record.GaugeId - targetID)
	})

	return StreamIterator{
		data:            data,
		streamIdx:       streamIdx,
		gaugeIdx:        gaugeIdx,
		epochIdentifier: epochIdentifier,
		iterations:      0,
	}
}

func (i *StreamIterator) Next() {
	i.gaugeIdx++
	if i.gaugeIdx >= len(i.data[i.streamIdx].DistributeTo.Records) {
		// Find the next appropriate stream
		i.streamIdx++
		for ; i.streamIdx < len(i.data); i.streamIdx++ {
			if i.data[i.streamIdx].DistrEpochIdentifier == i.epochIdentifier {
				i.gaugeIdx = 0
				break
			}
		}
	}
	i.iterations++
}

func (i StreamIterator) Stream() types.Stream {
	return i.data[i.streamIdx]
}

func (i StreamIterator) Gauge() types.DistrRecord {
	return i.data[i.streamIdx].DistributeTo.Records[i.gaugeIdx]
}

func (i StreamIterator) Valid() bool {
	return i.streamIdx < len(i.data)
}

func (i StreamIterator) Iterations() int {
	return i.iterations
}

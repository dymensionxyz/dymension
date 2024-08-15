package keeper

import (
	"fmt"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/internal/pagination"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	streams := k.GetActiveStreams(ctx)
	epochPointers, err := k.GetAllEpochPointers(ctx)
	if err != nil {
		return fmt.Errorf("get all epoch pointers: %w", err)
	}

	totalDistributed := sdk.NewCoins()
	maxIterations := k.GetParams(ctx).MaxIterationsPerBlock
	totalIterations := uint64(0)

	for _, p := range epochPointers {
		remainIterations := maxIterations - totalIterations

		if remainIterations <= 0 {
			// no more iterations available for this block
			break
		}

		newPointer, iters := IterateEpochPointer(p, streams, remainIterations, func(v StreamGauge) pagination.Stop {
			distributed, errX := k.DistributeToGauge(ctx, p.EpochCoins, v.Gauge, v.Stream.DistributeTo.TotalWeight)
			if errX != nil {
				// Ignore this gauge
				k.Logger(ctx).
					With("streamID", v.Stream.Id, "gaugeID", v.Gauge.GaugeId, "error", errX.Error()).
					Error("Failed to distribute to gauge")
			}

			totalDistributed = totalDistributed.Add(distributed...)
			return pagination.Continue
		})

		err = k.SaveEpochPointer(ctx, newPointer)
		if err != nil {
			return fmt.Errorf("save epoch pointer: %w", err)
		}
		totalIterations += iters
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

func IterateEpochPointer(
	p types.EpochPointer,
	streams []types.Stream,
	maxIterations uint64,
	cb func(v StreamGauge) pagination.Stop,
) (types.EpochPointer, uint64) {
	iter := NewStreamIterator(streams, p.StreamId, p.GaugeId, p.EpochIdentifier)
	iterations := pagination.Paginate(iter, maxIterations, cb)

	// Set pointer to the next unprocessed gauge. If the iterator is invalid, then
	// the last gauge is reached. Use special values in that case.
	if iter.Valid() {
		v := iter.Value()
		p.StreamId = v.Stream.Id
		p.GaugeId = v.Gauge.GaugeId
	} else {
		p.StreamId = types.MaxStreamID
		p.GaugeId = types.MaxGaugeID
	}

	return p, iterations
}

type StreamGauge struct {
	Stream types.Stream
	Gauge  types.DistrRecord
}

var _ pagination.Iterator[StreamGauge] = new(StreamIterator)

type StreamIterator struct {
	data            []types.Stream
	streamIdx       int
	gaugeIdx        int
	epochIdentifier string
	iterations      int
}

func NewStreamIterator(data []types.Stream, startStreamID uint64, startGaugeID uint64, epochIdentifier string) *StreamIterator {
	// streamIdx is the position where the stream is found, or the position where it would appear in the sort order
	streamIdx, _ := slices.BinarySearchFunc(data, startStreamID, func(stream types.Stream, targetID uint64) int {
		return cmpUint64(stream.Id, targetID)
	})

	// startStreamID is greater than all the existing streams, the pointer is initially invalid
	if streamIdx >= len(data) {
		return &StreamIterator{
			data:            data,
			streamIdx:       streamIdx,
			gaugeIdx:        0,
			epochIdentifier: epochIdentifier,
		}
	}

	// gaugeIdx is the position where the gauge is found, or the position where it would appear in the sort order
	gaugeIdx, _ := slices.BinarySearchFunc(data[streamIdx].DistributeTo.Records, startGaugeID, func(record types.DistrRecord, targetID uint64) int {
		return cmpUint64(record.GaugeId, targetID)
	})

	iter := &StreamIterator{
		data:            data,
		streamIdx:       streamIdx,
		gaugeIdx:        gaugeIdx,
		epochIdentifier: epochIdentifier,
	}

	if !iter.validInvariants() {
		iter.findNextStream()
	}

	return iter
}

// Next ... . Calling Next on the invalid iterator would cause panic. TODO: readme
func (i *StreamIterator) Next() {
	i.gaugeIdx++

	if !i.validInvariants() {
		i.findNextStream()
	}
}

// findNextStream find the next appropriate stream, e.i., a non-empty stream having the matching epoch identifier
func (i *StreamIterator) findNextStream() {
	// Put the pointer to the next stream
	i.gaugeIdx = 0
	i.streamIdx++
	for ; i.streamIdx < len(i.data); i.streamIdx++ {
		if i.validInvariants() {
			return
		}
	}
}

// If it is the last gauge of the stream, or it has mismatching epoch identifier, then try finding
// the next non-empty stream with the matching epoch identifier, and put a pointer to it's first gauge
func (i StreamIterator) validInvariants() bool {
	i1 := i.streamIdx < len(i.data)
	if !i1 {
		return false
	}

	i2 := len(i.data[i.streamIdx].DistributeTo.Records) != 0
	if !i2 {
		return false
	}

	i3 := i.data[i.streamIdx].DistrEpochIdentifier == i.epochIdentifier
	if !i3 {
		return false
	}

	i4 := i.gaugeIdx < len(i.data[i.streamIdx].DistributeTo.Records)
	if !i4 {
		return false
	}

	return true
}

func (i StreamIterator) Value() StreamGauge {
	return StreamGauge{
		Stream: i.data[i.streamIdx],
		Gauge:  i.data[i.streamIdx].DistributeTo.Records[i.gaugeIdx],
	}
}

func (i StreamIterator) Stream() types.Stream {
	return i.data[i.streamIdx]
}

func (i StreamIterator) Gauge() types.DistrRecord {
	return i.data[i.streamIdx].DistributeTo.Records[i.gaugeIdx]
}

func (i StreamIterator) Valid() bool {
	return i.validInvariants()
}

func cmpUint64(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

package keeper

import (
	"slices"

	"github.com/dymensionxyz/dymension/v3/utils/pagination"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func IterateEpochPointer(
	p types.EpochPointer,
	streams []types.Stream,
	maxIterations uint64,
	cb func(v StreamGauge) (stop bool, weight uint64),
) (types.EpochPointer, uint64) {
	iter := NewStreamIterator(streams, p.StreamId, p.GaugeId, p.EpochIdentifier)
	iterations := pagination.Paginate(iter, maxIterations, cb)

	// Set pointer to the next unprocessed gauge. If the iterator is invalid, then
	// the last gauge is reached. Use special values in that case.
	if iter.Valid() {
		v := iter.Value()
		p.Set(v.Stream.Id, v.Gauge.GaugeId)
	} else {
		p.SetToLastGauge()
	}

	return p, iterations
}

// StreamGauge is a special type to help StreamIterator implement pagination.Iterator.
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
}

// NewStreamIterator a new StreamIterator starting from the provided stream and gauge IDs. First, it finds a starting
// position in the stream slice. Then it checks if it is valid and tries to find the next appropriate stream if not.
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

// Next iterates to the next appropriate gauge. It can make the iterator invalid.
func (i *StreamIterator) Next() {
	i.gaugeIdx++

	if !i.validInvariants() {
		i.findNextStream()
	}
}

// findNextStream find the next appropriate stream.
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

// validInvariants validates the iterator invariants:
// 1. streamIdx is less than the number of streams: the iterator points to the existing stream
// 2. Stream is non-empty: there are some gauges assigned to this stream
// 3. Stream epoch identifier matches the provided
// 4. gaugeIdx is less than the number of gauges: the iterator points to the existing gauge
func (i StreamIterator) validInvariants() bool {
	/////  1. streamIdx is less than the number of streams
	return i.streamIdx < len(i.data) &&

		// 2. stream is non-empty
		len(i.data[i.streamIdx].DistributeTo.Records) != 0 &&

		// 3. stream epoch identifier matches the provided
		i.data[i.streamIdx].DistrEpochIdentifier == i.epochIdentifier &&

		// 4. gaugeIdx is less than the number of gauges
		i.gaugeIdx < len(i.data[i.streamIdx].DistributeTo.Records)
}

func (i StreamIterator) Value() StreamGauge {
	return StreamGauge{
		Stream: i.data[i.streamIdx],
		Gauge:  i.data[i.streamIdx].DistributeTo.Records[i.gaugeIdx],
	}
}

func (i StreamIterator) Valid() bool {
	return i.validInvariants()
}

func CmpStreams(a, b types.Stream) int {
	return cmpUint64(a.Id, b.Id)
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

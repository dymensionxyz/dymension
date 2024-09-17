package types

import (
	"math"
	"slices"
	"time"
)

const (
	IterationsNoLimit uint64 = math.MaxUint64

	MaxStreamID uint64 = math.MaxUint64
	MaxGaugeID  uint64 = math.MaxUint64

	MinStreamID uint64 = 0
	MinGaugeID  uint64 = 0
)

func NewEpochPointer(epochIdentifier string, epochDuration time.Duration) EpochPointer {
	return EpochPointer{
		StreamId:        MinStreamID,
		GaugeId:         MinGaugeID,
		EpochIdentifier: epochIdentifier,
		EpochDuration:   epochDuration,
	}
}

func (p *EpochPointer) Set(streamId uint64, gaugeId uint64) {
	p.StreamId = streamId
	p.GaugeId = gaugeId
}

func (p *EpochPointer) SetToFirstGauge() {
	p.Set(MinStreamID, MinGaugeID)
}

func (p *EpochPointer) SetToLastGauge() {
	p.Set(MaxStreamID, MaxGaugeID)
}

func SortEpochPointers(ep []EpochPointer) {
	slices.SortFunc(ep, func(a, b EpochPointer) int {
		return cmpDurations(a.EpochDuration, b.EpochDuration)
	})
}

func cmpDurations(a, b time.Duration) int {
	return cmpInt64(int64(a), int64(b))
}

func cmpInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

package types

import (
	"math"
)

const (
	IterationsNoLimit uint64 = math.MaxUint64

	MaxStreamID uint64 = math.MaxUint64
	MaxGaugeID  uint64 = math.MaxUint64

	MinStreamID uint64 = 0
	MinGaugeID  uint64 = 0
)

func NewEpochPointer(epochIdentifier string) EpochPointer {
	return EpochPointer{
		StreamId:        MinStreamID,
		GaugeId:         MinGaugeID,
		EpochIdentifier: epochIdentifier,
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

//
//func SortEpochPointers(ep []EpochPointer) {
//	slices.SortFunc(ep, func(a, b EpochPointer) int {
//
//	})
//}
//
//func

package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewStream creates a new stream struct given the required stream parameters.
func NewStream(id uint64, distrTo sdk.AccAddress, coins sdk.Coins, startTime time.Time, numEpochsPaidOver uint64, filledEpochs uint64, distrCoins sdk.Coins) Stream {
	return Stream{
		Id:                id,
		DistributeTo:      distrTo.String(),
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
		FilledEpochs:      filledEpochs,
		DistributedCoins:  distrCoins,
	}
}

// IsUpcomingStream returns true if the stream's distribution start time is after the provided time.
func (stream Stream) IsUpcomingStream(curTime time.Time) bool {
	return curTime.Before(stream.StartTime)
}

// IsActiveStream returns true if the stream is in an active state during the provided time.
func (stream Stream) IsActiveStream(curTime time.Time) bool {
	if curTime.After(stream.StartTime) || curTime.Equal(stream.StartTime) && (stream.FilledEpochs < stream.NumEpochsPaidOver) {
		return true
	}
	return false
}

// IsFinishedStream returns true if the stream is in a finished state during the provided time.
func (stream Stream) IsFinishedStream(curTime time.Time) bool {
	return !stream.IsUpcomingStream(curTime) && !stream.IsActiveStream(curTime)
}

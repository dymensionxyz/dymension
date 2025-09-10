package types

import (
	"fmt"
	time "time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewStream creates a new stream struct given the required stream parameters.
func NewStream(
	id uint64,
	distrTo DistrInfo,
	coins sdk.Coins,
	startTime time.Time,
	epochIdentifier string,
	numEpochsPaidOver uint64,
	sponsored bool,
	pumpParams *MsgCreateStream_PumpParams,
) Stream {
	epochCoins := coins.QuoInt(math.NewIntFromUint64(numEpochsPaidOver))
	var pump *PumpParams
	if pumpParams != nil {
		pump = &PumpParams{
			NumTopRollapps:  pumpParams.NumTopRollapps,
			EpochBudget:     epochCoins[0].Amount,
			EpochBudgetLeft: epochCoins[0].Amount,
			NumPumps:        pumpParams.NumPumps,
			PumpDistr:       pumpParams.PumpDistr,
		}
	}
	return Stream{
		Id:                   id,
		DistributeTo:         distrTo,
		Coins:                coins,
		StartTime:            startTime,
		DistrEpochIdentifier: epochIdentifier,
		NumEpochsPaidOver:    numEpochsPaidOver,
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins{},
		Sponsored:            sponsored,
		EpochCoins:           coins.QuoInt(math.NewIntFromUint64(numEpochsPaidOver)),
		PumpParams:           pump,
	}
}

// IsUpcomingStream returns true if the stream's distribution start time is after the provided time.
func (stream Stream) IsUpcomingStream(curTime time.Time) bool {
	return curTime.Before(stream.StartTime)
}

// IsActiveStream returns true if the stream is in an active state during the provided time.
func (stream Stream) IsActiveStream(curTime time.Time) bool {
	if (curTime.After(stream.StartTime) || curTime.Equal(stream.StartTime)) && (stream.FilledEpochs < stream.NumEpochsPaidOver) {
		return true
	}
	return false
}

// IsFinishedStream returns true if the stream is in a finished state during the provided time.
func (stream Stream) IsFinishedStream(curTime time.Time) bool {
	return !stream.IsUpcomingStream(curTime) && !stream.IsActiveStream(curTime)
}

func (stream *Stream) AddDistributedCoins(coins sdk.Coins) {
	stream.DistributedCoins = stream.DistributedCoins.Add(coins...)
}

func (stream Stream) Key() uint64 {
	return stream.Id
}

// IsPumpStream returns true if the stream has pump parameters configured
func (stream Stream) IsPumpStream() bool {
	return stream.PumpParams != nil
}

func DefaultPumpParams() *MsgCreateStream_PumpParams {
	return &MsgCreateStream_PumpParams{
		NumTopRollapps: 1,
		NumPumps:       1,
		PumpDistr:      PumpDistr_PUMP_DISTR_UNIFORM,
	}
}

func (p PumpParams) ValidateBasic() error {
	if p.PumpDistr == PumpDistr_PUMP_DISTR_UNSPECIFIED {
		return fmt.Errorf("pump distribution must be set")
	}
	if p.NumPumps == 0 {
		return fmt.Errorf("num pumps must be greater than 0")
	}
	switch t := p.Target.(type) {
	case *PumpParams_Pool:
		if err := sdk.ValidateDenom(t.Pool.TokenOut); err != nil {
			return fmt.Errorf("invalid token out: %w", err)
		}
	case *PumpParams_Rollapps:
		if t.Rollapps.NumTopRollapps == 0 {
			return fmt.Errorf("num top rollapps must be greater than 0")
		}
	default:
		return fmt.Errorf("invalid target type: %T", t)
	}
	return nil
}

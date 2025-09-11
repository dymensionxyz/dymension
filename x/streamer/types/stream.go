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
	pumpParams *PumpParams,
) Stream {
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
		PumpParams:           pumpParams,
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

func DefaultRollappsPumpParams() MsgCreateStream_PumpParams {
	return MsgCreateStream_PumpParams{
		Target:    &MsgCreateStream_PumpParams_Rollapps{Rollapps: &TargetTopRollapps{NumTopRollapps: 1}},
		NumPumps:  1,
		PumpDistr: PumpDistr_PUMP_DISTR_UNIFORM,
	}
}

func DefaultPoolPumpParams() MsgCreateStream_PumpParams {
	return MsgCreateStream_PumpParams{
		Target:    &MsgCreateStream_PumpParams_Pool{Pool: &TargetPool{TokenOut: sdk.DefaultBondDenom}},
		NumPumps:  1,
		PumpDistr: PumpDistr_PUMP_DISTR_UNIFORM,
	}
}

func (p MsgCreateStream_PumpParams) ValidateBasic() error {
	if p.PumpDistr == PumpDistr_PUMP_DISTR_UNSPECIFIED {
		return fmt.Errorf("pump distribution must be set")
	}
	if p.NumPumps == 0 {
		return fmt.Errorf("num pumps must be greater than 0")
	}
	switch t := p.Target.(type) {
	case *MsgCreateStream_PumpParams_Pool:
		if err := sdk.ValidateDenom(t.Pool.TokenOut); err != nil {
			return fmt.Errorf("invalid token out: %w", err)
		}
		if t == nil || t.Pool == nil {
			return fmt.Errorf("pool target must not be null")
		}
	case *MsgCreateStream_PumpParams_Rollapps:
		if t.Rollapps.NumTopRollapps == 0 {
			return fmt.Errorf("num top rollapps must be greater than 0")
		}
		if t == nil || t.Rollapps == nil {
			return fmt.Errorf("rollapps target must not be null")
		}
	default:
		return fmt.Errorf("invalid target type: %T", t)
	}
	return nil
}

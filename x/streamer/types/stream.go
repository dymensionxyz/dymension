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

func (stream Stream) LeftCoins() sdk.Coins {
	return stream.Coins.Sub(stream.DistributedCoins...)
}

type PumpTarget isMsgCreatePumpStream_Target

func PumpTargetRollapps(numTopRollapps uint32) PumpTarget {
	return &MsgCreatePumpStream_Rollapps{
		Rollapps: &TargetTopRollapps{NumTopRollapps: numTopRollapps},
	}
}

func PumpTargetPool(poolId uint64, tokenOut string) PumpTarget {
	return &MsgCreatePumpStream_Pool{
		Pool: &TargetPool{
			PoolId:   poolId,
			TokenOut: tokenOut,
		},
	}
}

func ValidatePumpStreamParams(coins sdk.Coins, numPumps uint64, pumpDistr PumpDistr, target PumpTarget) error {
	if coins.Len() != 1 {
		return fmt.Errorf("pump stream must have one coin")
	}
	if pumpDistr == PumpDistr_PUMP_DISTR_UNSPECIFIED {
		return fmt.Errorf("pump distribution must be set")
	}
	if numPumps == 0 {
		return fmt.Errorf("num pumps must be greater than 0")
	}
	switch t := target.(type) {
	case *MsgCreatePumpStream_Pool:
		if t == nil || t.Pool == nil {
			return fmt.Errorf("pool target must not be null")
		}
		if err := sdk.ValidateDenom(t.Pool.TokenOut); err != nil {
			return fmt.Errorf("invalid token out: %w", err)
		}
	case *MsgCreatePumpStream_Rollapps:
		if t == nil || t.Rollapps == nil {
			return fmt.Errorf("rollapps target must not be null")
		}
		if t.Rollapps.NumTopRollapps == 0 {
			return fmt.Errorf("num top rollapps must be greater than 0")
		}
	default:
		return fmt.Errorf("invalid target type: %T", t)
	}
	return nil
}

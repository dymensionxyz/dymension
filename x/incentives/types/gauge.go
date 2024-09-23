package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// NewAssetGauge creates a new asset gauge to stream rewards to some asset lockup conditions.
func NewAssetGauge(id uint64, isPerpetual bool, distrTo lockuptypes.QueryCondition, coins sdk.Coins, startTime time.Time, numEpochsPaidOver uint64) Gauge {
	return Gauge{
		Id:                id,
		IsPerpetual:       isPerpetual,
		DistributeTo:      &Gauge_Asset{&distrTo},
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
		FilledEpochs:      0,
		DistributedCoins:  sdk.NewCoins(),
	}
}

// NewRollappGauge creates a new rollapp gauge to stream rewards to a rollapp.
func NewRollappGauge(id uint64, rollappId string) Gauge {
	return Gauge{
		Id:                id,
		IsPerpetual:       true,
		DistributeTo:      &Gauge_Rollapp{Rollapp: &RollappGauge{RollappId: rollappId}},
		Coins:             sdk.NewCoins(),
		StartTime:         time.Time{},
		NumEpochsPaidOver: 0,
		FilledEpochs:      0,
		DistributedCoins:  sdk.NewCoins(),
	}
}

// IsActiveGauge returns true if the gauge is in an active state during the provided time.
func (gauge Gauge) IsActiveGauge(curTime time.Time) bool {
	if (curTime.After(gauge.StartTime) || curTime.Equal(gauge.StartTime)) && (gauge.IsPerpetual || gauge.FilledEpochs < gauge.NumEpochsPaidOver) {
		return true
	}
	return false
}

// IsUpcomingGauge returns true if the gauge's distribution start time is after the provided time.
func (gauge Gauge) IsUpcomingGauge(curTime time.Time) bool {
	return curTime.Before(gauge.StartTime)
}

// IsFinishedGauge returns true if the gauge is in a finished state during the provided time.
func (gauge Gauge) IsFinishedGauge(curTime time.Time) bool {
	return !gauge.IsUpcomingGauge(curTime) && !gauge.IsActiveGauge(curTime)
}

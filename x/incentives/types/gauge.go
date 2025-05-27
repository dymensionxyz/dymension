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

// NewEndorsementGauge creates a new endorsement gauge to stream rewards to rollapp endorsers.
func NewEndorsementGauge(id uint64, isPerpetual bool, rollappId string, coins sdk.Coins, startTime time.Time, numEpochsPaidOver uint64) Gauge {
	return Gauge{
		Id:                id,
		IsPerpetual:       isPerpetual,
		DistributeTo:      &Gauge_Endorsement{Endorsement: &EndorsementGauge{RollappId: rollappId}},
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
		FilledEpochs:      0,
		DistributedCoins:  sdk.NewCoins(),
	}
}

// IsActiveGauge returns true if the gauge is in an active state during the provided time.
// For EndorsementGauges (both perpetual and non-perpetual), "active" also means it must have non-zero coins.
// For other non-perpetual gauges, it's active if its filled_epochs < num_epochs_paid_over.
// For other perpetual gauges, it's active if started.
func (gauge Gauge) IsActiveGauge(curTime time.Time) bool {
	if curTime.Before(gauge.StartTime) {
		return false // Not started yet
	}

	// Check if it's an EndorsementGauge
	// Note: gauge.GetEndorsement() would be cleaner if available, but direct type assertion is used here
	// based on how NewEndorsementGauge and other parts of the codebase might construct/access it.
	// This assumes DistributeTo is one of the concrete *Gauge_... types.
	var isEndorsementGauge bool
	if _, ok := gauge.DistributeTo.(*Gauge_Endorsement); ok {
		isEndorsementGauge = true
	}

	if gauge.IsPerpetual {
		if isEndorsementGauge {
			// Perpetual Endorsement Gauge is active if it has funds
			return !gauge.Coins.IsZero()
		}
		// Other perpetual gauges are active once started
		return true
	} else {
		// Non-perpetual gauges
		if isEndorsementGauge {
			// Non-perpetual Endorsement Gauge is active if it has funds
			// NumEpochsPaidOver and FilledEpochs are not used for its lifecycle determination.
			return !gauge.Coins.IsZero()
		}
		// Other non-perpetual gauges depend on epochs
		return gauge.FilledEpochs < gauge.NumEpochsPaidOver
	}
}

// IsUpcomingGauge returns true if the gauge's distribution start time is after the provided time.
func (gauge Gauge) IsUpcomingGauge(curTime time.Time) bool {
	return curTime.Before(gauge.StartTime)
}

// IsFinishedGauge returns true if the gauge is in a finished state.
// - If a gauge is upcoming, it's not finished.
// - Perpetual gauges (of any type) are never considered finished by this method.
// - For non-perpetual EndorsementGauges, "finished" means its coins are depleted.
// - For other non-perpetual gauges, "finished" means its filled_epochs >= num_epochs_paid_over.
func (gauge Gauge) IsFinishedGauge(curTime time.Time) bool {
	if gauge.IsUpcomingGauge(curTime) {
		return false // Not finished if it hasn't started
	}

	// Perpetual gauges of any type are never considered finished by this method's criteria.
	// Their lifecycle is ongoing, potentially awaiting refills or other actions.
	if gauge.IsPerpetual {
		return false
	}

	// At this point, gauge is not upcoming and not perpetual.
	// Now, determine if it's an EndorsementGauge.
	var isEndorsementGauge bool
	if _, ok := gauge.DistributeTo.(*Gauge_Endorsement); ok {
		isEndorsementGauge = true
	}

	if isEndorsementGauge {
		// Non-perpetual Endorsement Gauge is finished if its coins are depleted.
		// Using IsZero() for robustness, ensuring all coin amounts are zero.
		return gauge.Coins.IsZero()
	} else {
		// Non-perpetual, non-Endorsement Gauges are finished based on epoch counting.
		return gauge.FilledEpochs >= gauge.NumEpochsPaidOver
	}
}

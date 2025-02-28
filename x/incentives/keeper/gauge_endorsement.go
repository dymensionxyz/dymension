package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

func (k Keeper) CreateEndorsementGauge(ctx sdk.Context, rollappId string) (uint64, error) {
	// Ensure the rollapp exists
	_, found := k.rk.GetRollapp(ctx, rollappId)
	if !found {
		return 0, fmt.Errorf("rollapp %s not found", rollappId)
	}

	gauge := types.NewEndorsementGauge(k.GetLastGaugeID(ctx)+1, rollappId)

	err := k.setGauge(ctx, &gauge)
	if err != nil {
		return 0, err
	}
	k.SetLastGaugeID(ctx, gauge.Id)

	combinedKeys := combineKeys(types.KeyPrefixUpcomingGauges, getTimeKey(gauge.StartTime))
	err = k.CreateGaugeRefKeys(ctx, &gauge, combinedKeys, true)
	if err != nil {
		return 0, err
	}
	k.hooks.AfterCreateGauge(ctx, gauge.Id)
	return gauge.Id, nil
}

// CONTRACT: the gauge must be an endorsement gauge
// CONTRACT: the gauge must exist
// CONTRACT: this must be called on epoch end
func (k Keeper) updateEndorsementGaugeOnEpochEnd(ctx sdk.Context, gauge types.Gauge) error {
	gaugeBalance := gauge.Coins.Sub(gauge.DistributedCoins...)
	epochRewards := gaugeBalance.QuoInt(math.NewIntFromUint64(gauge.NumEpochsPaidOver - gauge.FilledEpochs))

	endorsement := gauge.DistributeTo.(*types.Gauge_Endorsement).Endorsement
	endorsement.EpochRewards = epochRewards // we operate a pointer
	gauge.FilledEpochs += 1

	if err := k.setGauge(ctx, &gauge); err != nil {
		return err
	}

	return nil
}

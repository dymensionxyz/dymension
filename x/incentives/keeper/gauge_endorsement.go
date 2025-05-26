package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// CreateEndorsementGauge creates a gauge and sends coins to the gauge.
func (k Keeper) CreateEndorsementGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo types.EndorsementGauge, startTime time.Time, numEpochsPaidOver uint64) (uint64, error) {
	// Ensure the rollapp exists
	_, found := k.rk.GetRollapp(ctx, distrTo.RollappId)
	if !found {
		return 0, fmt.Errorf("rollapp %s not found", distrTo.RollappId)
	}

	gauge := types.NewEndorsementGauge(k.GetLastGaugeID(ctx)+1, isPerpetual, distrTo.RollappId, coins, startTime, numEpochsPaidOver)

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, gauge.Coins); err != nil {
		return 0, err
	}

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
	epochRewards := gaugeBalance
	if !gauge.IsPerpetual {
		remainingEpochs := math.NewIntFromUint64(gauge.NumEpochsPaidOver - gauge.FilledEpochs)
		epochRewards = gaugeBalance.QuoInt(remainingEpochs)
	}

	gauge.FilledEpochs += 1
	gauge.DistributedCoins = gauge.DistributedCoins.Add(epochRewards...)

	// Update endorsement total coins with the epoch rewards
	endorsementGauge := gauge.DistributeTo.(*types.Gauge_Endorsement)

	err := k.spk.UpdateEndorsementTotalCoins(ctx, endorsementGauge.Endorsement.RollappId, epochRewards)
	if err != nil {
		return fmt.Errorf("update endorsement total coins: %w", err)
	}

	if err := k.setGauge(ctx, &gauge); err != nil {
		return err
	}

	return nil
}

func (k Keeper) DistributeEndorsementRewards(ctx sdk.Context, user sdk.AccAddress, gaugeId uint64, rewards sdk.Coins) error {
	gauge, err := k.GetGaugeByID(ctx, gaugeId)
	if err != nil {
		return fmt.Errorf("get gauge by ID: %w", err)
	}

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, user, rewards)
	if err != nil {
		return fmt.Errorf("send coins from x/incentives to user: %w", err)
	}

	gauge.DistributedCoins = gauge.DistributedCoins.Add(rewards...)
	err = k.setGauge(ctx, gauge)
	if err != nil {
		return fmt.Errorf("set gauge: %w", err)
	}

	return nil
}

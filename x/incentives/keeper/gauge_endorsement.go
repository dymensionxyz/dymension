package keeper

import (
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
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

	// Update endorsement total coins with the epoch rewards
	endorsementGauge := gauge.DistributeTo.(*types.Gauge_Endorsement)

	err := k.spk.UpdateEndorsementTotalCoins(ctx, endorsementGauge.Endorsement.RollappId, epochRewards)
	if errors.Is(err, sponsorshiptypes.ErrNoEndorsers) {
		// Don't fill this epoch, save rewards for the future
		return nil
	}
	if err != nil {
		return fmt.Errorf("update endorsement total coins: %w", err)
	}

	gauge.FilledEpochs += 1
	gauge.DistributedCoins = gauge.DistributedCoins.Add(epochRewards...)

	if err := k.setGauge(ctx, &gauge); err != nil {
		return err
	}

	return nil
}

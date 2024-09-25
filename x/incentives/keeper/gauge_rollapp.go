package keeper

import (
	"fmt"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateRollappGauge creates a gauge and sends coins to the gauge.
func (k Keeper) CreateRollappGauge(ctx sdk.Context, rollappId string) (uint64, error) {
	// Ensure the rollapp exists
	_, found := k.rk.GetRollapp(ctx, rollappId)
	if !found {
		return 0, fmt.Errorf("rollapp %s not found", rollappId)
	}

	gauge := types.NewRollappGauge(k.GetLastGaugeID(ctx)+1, rollappId)

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

func (k Keeper) distributeToRollappGauge(ctx sdk.Context, gauge types.Gauge) (totalDistrCoins sdk.Coins, err error) {
	// Get the rollapp owner
	rollapp, found := k.rk.GetRollapp(ctx, gauge.GetRollapp().RollappId)
	if !found {
		return sdk.Coins{}, fmt.Errorf("gauge %d: rollapp %s not found", gauge.Id, gauge.GetRollapp().RollappId)
	}
	// Ignore the error since the owner must always be valid in x/rollapp
	addr := sdk.MustAccAddressFromBech32(rollapp.Owner)

	totalDistrCoins = gauge.Coins.Sub(gauge.DistributedCoins...)
	if totalDistrCoins.Empty() {
		ctx.Logger().Debug(fmt.Sprintf("gauge %d is empty, skipping", gauge.Id))
		return sdk.Coins{}, nil
	}

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, totalDistrCoins)
	if err != nil {
		return sdk.Coins{}, err
	}

	return totalDistrCoins, nil
}

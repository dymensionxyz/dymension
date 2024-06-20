package keeper

import (
	"fmt"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateRollappGauge creates a gauge and sends coins to the gauge.
func (k Keeper) CreateRollappGauge(ctx sdk.Context, owner sdk.AccAddress, rollappId string) (uint64, error) {
	// Ensure the rollapp exists
	_, found := k.rk.GetRollapp(ctx, rollappId)
	if !found {
		return 0, fmt.Errorf("rollapp %s not found", rollappId)
	}

	gauge := types.Gauge{
		Id:          k.GetLastGaugeID(ctx) + 1,
		IsPerpetual: true,
		DistributeTo: &types.Gauge_Rollapp{
			Rollapp: &types.RollappGauge{RollappId: rollappId},
		},
		NumEpochsPaidOver: 1,
	}

	err := k.setGauge(ctx, &gauge)
	if err != nil {
		return 0, err
	}
	k.SetLastGaugeID(ctx, gauge.Id)

	combinedKeys := combineKeys(types.KeyPrefixUpcomingGauges, getTimeKey(gauge.StartTime))
	activeOrUpcomingGauge := true

	err = k.CreateGaugeRefKeys(ctx, &gauge, combinedKeys, activeOrUpcomingGauge)
	if err != nil {
		return 0, err
	}
	k.hooks.AfterCreateGauge(ctx, gauge.Id)
	return gauge.Id, nil
}

func (k Keeper) distributeToRollappGauge(ctx sdk.Context, gauge types.Gauge) (totalDistrCoins sdk.Coins, err error) {
	defer func() {
		err = k.updateGaugePostDistribute(ctx, gauge, totalDistrCoins)
	}()

	seqs := k.sq.GetSequencersByRollapp(ctx, gauge.GetRollapp().RollappId)
	if len(seqs) == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("no sequencers found for rollapp %s", gauge.GetRollapp().RollappId))
		return sdk.Coins{}, nil
	}

	var addr sdk.AccAddress
	for _, seq := range seqs {
		if seq.Proposer {
			addr, _ = sdk.AccAddressFromBech32(seq.SequencerAddress)
			break
		}
	}

	if addr.Empty() {
		k.Logger(ctx).Error(fmt.Sprintf("no active sequencer found for rollapp %s", gauge.GetRollapp().RollappId))
		return sdk.Coins{}, nil
	}

	totalDistrCoins = gauge.Coins.Sub(gauge.DistributedCoins...)
	if totalDistrCoins.Empty() {
		ctx.Logger().Debug(fmt.Sprintf("gauge %d is empty, skipping", gauge.Id))
		return totalDistrCoins, nil
	}

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, totalDistrCoins)
	if err != nil {
		return totalDistrCoins, err
	}

	return totalDistrCoins, err

}

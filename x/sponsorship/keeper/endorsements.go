package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// UpdateEndorsements updates the endorsements for all the rollapps from `update`.
//
// CONTRACT: All gauges must exist.
func (k Keeper) UpdateEndorsements(ctx sdk.Context, update types.Distribution) error {
	for _, weight := range update.Gauges {
		// The gauge must exist. It must be validated in Keeper.validateWeights beforehand.
		gauge, _ := k.incentivesKeeper.GetGaugeByID(ctx, weight.GaugeId)

		raGauge, ok := gauge.DistributeTo.(*incentivestypes.Gauge_Rollapp)
		if !ok {
			// the gauge is not a RA gauge
			continue
		}

		err := k.UpdateEndorsement(ctx, raGauge.Rollapp.RollappId, types.UpdateTotalShares(weight.Power))
		if err != nil {
			return fmt.Errorf("update endorsement shares: rollapp %s: %w", raGauge.Rollapp.RollappId, err)
		}
	}
	return nil
}

func (k Keeper) Claim(ctx sdk.Context, sender sdk.AccAddress, gaugeId uint64) error {
	return nil
}

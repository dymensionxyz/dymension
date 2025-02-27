package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var _ incentivestypes.IncentiveHooks = IncentivesHooks{}

type IncentivesHooks struct {
	k Keeper
}

func (k Keeper) IncentiveHooks() IncentivesHooks {
	return IncentivesHooks{k}
}

// AfterCreateGauge adds the endorsement gauge to the rollapp endorsement when an endorsement gauge is created.
// We need to keep track of all the endorsement gauges for the rollapp so ...
// TODO: do we need it actually?
func (i IncentivesHooks) AfterCreateGauge(ctx sdk.Context, gaugeId uint64) {
	// the gauge must exist at this point
	gauge, _ := i.k.incentivesKeeper.GetGaugeByID(ctx, gaugeId)

	raGauge, ok := gauge.DistributeTo.(*incentivestypes.Gauge_Endorsement)
	if !ok {
		// here we focus on endorsement gauges only
		return
	}

	err := i.k.UpdateEndorsement(ctx, raGauge.Endorsement.RollappId, types.AddEndorsementGauge(gaugeId))
	if err != nil {
		ctx.Logger().With("rollappId", raGauge.Endorsement.RollappId, "gaugeId", gaugeId, "error", err).
			Error("Failed to add endorsement gauge to rollapp endorsement")
	}
}

func (i IncentivesHooks) GaugeFinished(ctx sdk.Context, gaugeId uint64) {
	// the gauge must exist at this point
	gauge, _ := i.k.incentivesKeeper.GetGaugeByID(ctx, gaugeId)

	raGauge, ok := gauge.DistributeTo.(*incentivestypes.Gauge_Endorsement)
	if !ok {
		// here we focus on endorsement gauges only
		return
	}

	err := i.k.UpdateEndorsement(ctx, raGauge.Endorsement.RollappId, types.RemoveEndorsementGauge(gaugeId))
	if err != nil {
		ctx.Logger().With("rollappId", raGauge.Endorsement.RollappId, "gaugeId", gaugeId, "error", err).
			Error("Failed to remove endorsement gauge from rollapp endorsement")
	}
}

func (i IncentivesHooks) AfterAddToGauge(sdk.Context, uint64) {}

func (i IncentivesHooks) AfterEpochDistribution(sdk.Context) {}

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

var _ incentivestypes.IncentiveHooks = IncentivesHooks{}

type IncentivesHooks struct {
	k Keeper
}

func (k Keeper) IncentiveHooks() IncentivesHooks {
	return IncentivesHooks{k}
}

func (i IncentivesHooks) AfterCreateGauge(ctx sdk.Context, gaugeId uint64) {
	//TODO implement me
	panic("implement me")
}

func (i IncentivesHooks) GaugeFinished(ctx sdk.Context, gaugeId uint64) {
	//TODO implement me
	panic("implement me")
}

func (i IncentivesHooks) AfterAddToGauge(sdk.Context, uint64) {}

func (i IncentivesHooks) AfterEpochDistribution(sdk.Context) {}

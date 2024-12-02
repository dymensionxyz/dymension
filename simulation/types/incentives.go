package types

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// RandomGauge takes a context, then returns a random existing gauge.
func RandomGauge(ctx sdk.Context, r *rand.Rand, k IncentivesKeeper) *incentivestypes.Gauge {
	gauges := k.GetGauges(ctx)
	if len(gauges) == 0 {
		return nil
	}
	return &gauges[r.Intn(len(gauges))]
}

// RandomGaugeSubset takes a context, a random number generator, and an IncentivesKeeper,
// then returns a random subset of gauges. Gauges are non-duplicated.
func RandomGaugeSubset(ctx sdk.Context, r *rand.Rand, k IncentivesKeeper) []incentivestypes.Gauge {
	allGauges := k.GetGauges(ctx)
	if len(allGauges) == 0 {
		return nil
	}

	numGauges := r.Intn(len(allGauges)) + 1

	// Shuffle the list of all gauges
	r.Shuffle(len(allGauges), func(i, j int) {
		allGauges[i], allGauges[j] = allGauges[j], allGauges[i]
	})

	// Select the first numGauges elements from the shuffled list
	return allGauges[:numGauges]
}

package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// RandomizedGenState generates a random GenesisState for iro module
func RandomizedGenState(simState *module.SimulationState) {
	// Generate number of plans. each operation will test one plan in random
	numPlans := 30
	plans := make([]types.Plan, numPlans)

	for i := 0; i < numPlans; i++ {
		plan := generateRandomPlan(simState.Rand, uint64(i+1))
		plans[i] = plan
	}

	iroGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		Plans:  plans,
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&iroGenesis)
}

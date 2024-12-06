package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// RandomizedGenState generates a random GenesisState
func RandomizedGenState(simState *module.SimulationState) {

	g := types.GenesisState{
		Params: types.DefaultParams(),
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&g)
}

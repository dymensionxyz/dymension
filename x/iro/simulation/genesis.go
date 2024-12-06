package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// RandomizedGenState generates a random GenesisState for iro module
func RandomizedGenState(simState *module.SimulationState) {
	iroGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		Plans:  nil,
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&iroGenesis)
}

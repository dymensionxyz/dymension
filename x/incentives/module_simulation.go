package incentives

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/dymensionxyz/dymension/v3/x/incentives/simulation"
)

// ----------------------------------------------------------------------------
// AppModuleSimulation
// ----------------------------------------------------------------------------

// GenerateGenesisState creates a randomized GenState of x/incentives.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for supply module's types.
func (AppModule) RegisterStoreDecoder(sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the module's operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(simState.AppParams, simState.Cdc, am.accountKeeper, am.bankKeeper, am.epochKeeper, am.keeper)
}

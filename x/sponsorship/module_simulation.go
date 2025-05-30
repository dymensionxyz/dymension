package sponsorship

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/simulation"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// ----------------------------------------------------------------------------
// AppModuleSimulation
// ----------------------------------------------------------------------------

// GenerateGenesisState creates a randomized GenState of x/incentives.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for supply module's types.
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types.ModuleName] = collcompat.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.Schema())
}

// WeightedOperations returns the all the module's operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(simState.AppParams, simState.TxConfig, am.accountKeeper, am.bankKeeper, am.incentivesKeeper, am.stakingKeeper, am.keeper)
}

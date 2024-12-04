package streamer

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"

	"github.com/dymensionxyz/dymension/v3/x/streamer/simulation"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// ----------------------------------------------------------------------------
// AppModuleSimulation
// ----------------------------------------------------------------------------

// GenerateGenesisState creates a randomized GenState of x/incentives.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for supply module's types.
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[types.ModuleName] = collcompat.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.Schema())
}

// WeightedOperations returns the all the module's operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(simState.AppParams, simState.Cdc, am.keeper)
}

// ProposalContents returns all the gov content functions used to
// simulate governance proposals.
func (AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent { //nolint:staticcheck
	return simulation.ProposalContents()
}

// ProposalMsgs returns all the gov msgs used to simulate governance proposals.
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RandomizedParams creates randomized fee market param changes for the simulator.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	return nil
}

package rollapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/simulation"
)

// ----------------------------------------------------------------------------
// AppModuleSimulation
// ----------------------------------------------------------------------------

func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
}

func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.NewOpFactory(*am.keeper, am.sk, simState).Messages()
}

// Legacy proposals
func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent { //nolint:staticcheck
	return nil
}

// Modern proposals
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return nil
}

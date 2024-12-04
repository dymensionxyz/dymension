package streamer

import (
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

func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[types.ModuleName] = collcompat.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.Schema())
}

func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.NewOpFactory(&am.keeper, am.sk, simState).Messages()
}

func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent { //nolint:staticcheck
	return simulation.NewOpFactory(&am.keeper, am.sk, simState).Proposals()
}

func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return nil
}

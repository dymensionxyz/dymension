package dymns

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = params.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	eibcGenesis := dymnstypes.GenesisState{
		Params: dymnstypes.DefaultParams(),
	}
	simState.GenState[dymnstypes.ModuleName] = simState.Cdc.MustMarshalJSON(&eibcGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	return operations
}

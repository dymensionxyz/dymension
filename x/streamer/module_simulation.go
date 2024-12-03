package streamer

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/simulation"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// avoid unused import cycle
var (
	_ = simulation.FindAccount
	_ = keeper.Keeper{}
	_ = rand.Int63()
	_ = sdk.AccAddress{}
	_ = baseapp.BaseApp{}
)

// GenerateGenesisState creates a randomized GenState of the streamer module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for streamer module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the streamer module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams,
		simState.Cdc,
		am.accountKeeper,
		am.bankKeeper,
		am.incentivesKeeper,
		am.keeper,
	)
}

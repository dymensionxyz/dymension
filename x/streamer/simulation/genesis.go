package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// Simulation parameter constants
const (
	MaxIterationsPerBlock = "max_iterations_per_block"
)

// GenMaxIterationsPerBlock randomized MaxIterationsPerBlock
func GenMaxIterationsPerBlock(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 100, 10000))
}

// RandomizedGenState generates a random GenesisState for streamer module
func RandomizedGenState(simState *module.SimulationState) {
	var maxIterationsPerBlock uint64

	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxIterationsPerBlock, &maxIterationsPerBlock, simState.Rand,
		func(r *rand.Rand) { maxIterationsPerBlock = GenMaxIterationsPerBlock(r) },
	)

	streamerGenesis := types.GenesisState{
		Params: types.Params{
			MaxIterationsPerBlock: maxIterationsPerBlock,
		},
		Streams:       []types.Stream{},
		LastStreamId:  0,
		EpochPointers: []types.EpochPointer{},
	}

	bz, err := json.MarshalIndent(&streamerGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated streamer parameters:\n%s\n", bz)

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&streamerGenesis)
}

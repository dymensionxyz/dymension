package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func getMinAllocationWeight(r *rand.Rand) math.Int {
	w, _ := simulation.RandPositiveInt(r, types.DefaultMinAllocationWeight)
	return w
}

func getMinVotingPower(r *rand.Rand) math.Int {
	// use comparatively small numbers as the initial account balance is always bounded with max Int64 in simulation
	w, _ := simulation.RandPositiveInt(r, commontypes.ADYM.MulRaw(100_000))
	return w
}

// RandomizedGenState generates a random GenesisState for staking
func RandomizedGenState(simState *module.SimulationState) {
	genesis := &types.GenesisState{
		Params: types.Params{
			MinAllocationWeight: getMinAllocationWeight(simState.Rand),
			MinVotingPower:      getMinVotingPower(simState.Rand),
		},
		VoterInfos: nil,
	}

	bz, err := json.MarshalIndent(genesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated sponsorship parameters:\n%s\n", bz)

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}

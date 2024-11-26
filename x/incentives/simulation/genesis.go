package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

func getFee(r *rand.Rand) math.Int {
	// use comparatively small numbers as the initial account balance is always bounded with max Int64 in simulation
	w, _ := simulation.RandPositiveInt(r, commontypes.ADYM.MulRaw(100_000))
	return w
}

// RandomizedGenState generates a random GenesisState for x/incentives.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.GenesisState{
		Params: types.Params{
			DistrEpochIdentifier: "day",
			CreateGaugeBaseFee:   getFee(simState.Rand),
			AddToGaugeBaseFee:    getFee(simState.Rand),
			AddDenomFee:          getFee(simState.Rand),
		},
		LockableDurations: []time.Duration{
			time.Second,
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
	}

	bz, err := json.MarshalIndent(genesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated incentives parameters:\n%s\n", bz)

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&genesis)
}

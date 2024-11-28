package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// RandomizedGenState generates a random GenesisState for iro module
func RandomizedGenState(simState *module.SimulationState) {

	// Generate random number of initial plans (1-5)
	numPlans := 1 + simState.Rand.Intn(4)
	plans := make([]types.Plan, numPlans)

	for i := 0; i < numPlans; i++ {
		plan := generateRandomPlan(simState.Rand, uint64(i+1))
		plans[i] = plan
	}

	// TODO: randomize params

	iroGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		Plans:  plans,
	}

	bz, err := json.MarshalIndent(&iroGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated iro parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&iroGenesis)
}

func generateRandomPlan(r *rand.Rand, id uint64) types.Plan {
	// Generate random account address for owner
	rollappId := "rollapp_1234-1" // FIXME: randomize

	// Generate random bonding curve
	curve := generateRandomBondingCurve(r)

	// Generate random timestamps for pre-launch and start time
	now := time.Now()
	preLaunchTime := now.Add(time.Duration(r.Int63n(7*24)) * time.Hour)       // Within a week
	startTime := preLaunchTime.Add(time.Duration(r.Int63n(7*24)) * time.Hour) // Within a week after pre-launch

	// Generate random allocated amount (between 1M and 10M)
	// FIXME: do all 10^18
	baseDenom := types.IRODenom(rollappId)
	allocatedAmount := simtypes.RandomAmount(r, math.NewInt(1000000000000000000)).Add(math.NewInt(1000))
	allocation := sdk.Coin{
		Denom:  baseDenom,
		Amount: allocatedAmount,
	}

	plan := types.NewPlan(id, rollappId, allocation, curve, startTime, preLaunchTime, types.IncentivePlanParams{})

	return plan
}

func generateRandomBondingCurve(r *rand.Rand) types.BondingCurve {
	m := simtypes.RandomDecAmount(r, math.LegacyMustNewDecFromStr("1"))
	n := simtypes.RandomDecAmount(r, math.LegacyMustNewDecFromStr("2"))
	// fixme: enforce 	MaxNPrecision = 3 // Maximum allowed decimal precision for the N parameter

	// linear bonding curve as default
	return types.BondingCurve{
		M: m,
		N: n,
		C: math.LegacyZeroDec(),
	}
}

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
	"github.com/dymensionxyz/sdk-utils/utils/urand"
)

var (
	// DYM represents 1 DYM. Equals to 10^18 ADYM.
	DYM = math.NewIntWithDecimal(1, 18)
)

// RandomizedGenState generates a random GenesisState for iro module
func RandomizedGenState(simState *module.SimulationState) {
	// Generate random number of initial plans (1-10)
	numPlans := 1 + simState.Rand.Intn(10)
	plans := make([]types.Plan, numPlans)

	for i := 0; i < numPlans; i++ {
		plan := generateRandomPlan(simState.Rand, uint64(i+1))
		plans[i] = plan
	}

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
	rollappId := urand.RollappID()

	// Generate random bonding curve
	curve := generateRandomBondingCurve(r)

	// Generate random timestamps for pre-launch and start time
	startTime := time.Now()
	preLaunchTime := startTime.Add(24 * time.Hour)

	// Generate random allocated amount (between 1000 and 10B)
	baseDenom := types.IRODenom(rollappId)
	allocatedAmount := simtypes.RandomAmount(r, math.NewInt(1e10)).Add(math.NewInt(1000)).MulRaw(1e18)
	allocation := sdk.Coin{
		Denom:  baseDenom,
		Amount: allocatedAmount,
	}

	plan := types.NewPlan(id, rollappId, allocation, curve, startTime, preLaunchTime, types.DefaultIncentivePlanParams())

	// randomize starting sold amount (ensure > 1)
	minSoldAmt := math.NewInt(1).MulRaw(1e18) // 1 token minimum
	soldAmt := simtypes.RandomAmount(r, allocatedAmount.Sub(minSoldAmt)).Add(minSoldAmt)
	plan.SoldAmt = soldAmt

	return plan
}

func generateRandomBondingCurve(r *rand.Rand) types.BondingCurve {
	// TODO: fix me: genreate values over target raised dym

	// we set M to close values we see in the real world
	m := simtypes.RandomDecAmount(r, math.LegacyNewDecFromIntWithPrec(math.NewInt(1), 6)).Add(math.LegacyNewDecFromIntWithPrec(math.NewInt(1), 9))

	// Generate N with maximum precision of 3
	nInt := r.Int63n(1500)                  // Generate a random integer between 0 and 1499
	n := math.LegacyNewDecWithPrec(nInt, 3) // Convert to decimal with 3 decimal places
	// add 0.5 as minimum value
	n = n.Add(math.LegacyNewDecWithPrec(5, 1)) // Add 0.5

	return types.BondingCurve{
		M: m,
		N: n,
		C: math.LegacyZeroDec(),
	}
}

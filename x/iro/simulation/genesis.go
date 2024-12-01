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
	rollappId := urand.RollappID()

	// Generate random bonding curve
	curve := generateRandomBondingCurve(r)

	// Generate random timestamps for pre-launch and start time
	startTime := time.Now()
	preLaunchTime := startTime.Add(24 * time.Hour)

	// Generate random allocated amount (between 1000 and 1B)
	baseDenom := types.IRODenom(rollappId)
	allocatedAmount := simtypes.RandomAmount(r, math.NewInt(10^9)).Add(math.NewInt(1000)).Mul(DYM)
	allocation := sdk.Coin{
		Denom:  baseDenom,
		Amount: allocatedAmount,
	}

	plan := types.NewPlan(id, rollappId, allocation, curve, startTime, preLaunchTime, types.DefaultIncentivePlanParams())

	// randomize starting sold amount
	soldAmt := simtypes.RandomAmount(r, allocatedAmount)
	plan.SoldAmt = soldAmt

	return plan
}

func generateRandomBondingCurve(r *rand.Rand) types.BondingCurve {
	m := simtypes.RandomDecAmount(r, math.LegacyMustNewDecFromStr("1"))

	// Generate N with maximum precision of 3
	nInt := r.Int63n(1000)                     // Generate a random integer between 0 and 999
	nDec := math.LegacyNewDecWithPrec(nInt, 3) // Convert to decimal with 3 decimal places
	n := nDec.Add(math.LegacyOneDec())         // Add 1 to ensure n > 1

	return types.BondingCurve{
		M: m,
		N: n,
		C: math.LegacyZeroDec(),
	}
}

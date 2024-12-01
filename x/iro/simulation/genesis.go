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

	// Generate random timestamps for pre-launch and start time
	startTime := time.Now()
	preLaunchTime := startTime.Add(24 * time.Hour)

	// Generate random allocated amount (between 100_000 and 1B)
	baseDenom := types.IRODenom(rollappId)
	allocatedAmount := simtypes.RandomAmount(r, math.NewInt(1e9).SubRaw(100_000)).AddRaw(100_000).MulRaw(1e18)
	allocation := sdk.Coin{
		Denom:  baseDenom,
		Amount: allocatedAmount,
	}

	// Generate random bonding curve
	// var curve types.BondingCurve
	curve := generateRandomBondingCurve(r, allocatedAmount)
	// for range 10 {
	// 	// make sure we generate a valid curve
	// 	if !c.M.IsPositive() {
	// 		continue
	// 	}
	// 	c :=
	// 	curve = c
	// 	break
	// }
	plan := types.NewPlan(id, rollappId, allocation, curve, startTime, preLaunchTime, types.DefaultIncentivePlanParams())

	// randomize starting sold amount (ensure > 1)
	minSoldAmt := math.NewInt(1).MulRaw(1e18) // 1 token minimum
	minUnsoldAmt := math.NewInt(100).MulRaw(1e18)
	soldAmt := simtypes.RandomAmount(r, allocatedAmount.Sub(minUnsoldAmt).Sub(minSoldAmt)).Add(minSoldAmt)
	plan.SoldAmt = soldAmt

	return plan
}

func generateRandomBondingCurve(r *rand.Rand, allocatedAmount math.Int) types.BondingCurve {
	// Generate N with maximum precision of 3
	nInt := r.Int63n(1500)                  // Generate a random integer between 0 and 1499
	n := math.LegacyNewDecWithPrec(nInt, 3) // Convert to decimal with 3 decimal places
	// add 0.5 as minimum value
	n = n.Add(math.LegacyNewDecWithPrec(5, 1)) // Add 0.5

	// targetRaiseDYM between 10 and 100M DYM
	targetRaiseDYM := simtypes.RandomAmount(r, math.NewInt(1e8)).AddRaw(10)

	// Scale allocatedAmount from base denomination to decimal representation
	allocatedTokens := types.ScaleFromBase(allocatedAmount, types.DYMDecimals)

	m := types.CalculateM(
		math.LegacyNewDecFromInt(targetRaiseDYM),
		allocatedTokens,
		n,
		math.LegacyZeroDec(),
	)

	return types.BondingCurve{
		M: m,
		N: n,
		C: math.LegacyZeroDec(),
	}
}

package simulation

import (
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
	// Generate number of plans. each operation will test one plan in random
	numPlans := 30
	plans := make([]types.Plan, numPlans)

	for i := 0; i < numPlans; i++ {
		plan := generateRandomPlan(simState.Rand, uint64(i+1))
		plans[i] = plan
	}

	iroGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		Plans:  plans,
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&iroGenesis)
}

func generateRandomPlan(r *rand.Rand, id uint64) types.Plan {
	rollappId := urand.RollappID()

	startTime := time.Now()
	preLaunchTime := startTime.Add(24 * time.Hour)

	// Generate random allocated amount (between 100_000 and 1_000_000_000 RA tokens)
	baseDenom := types.IRODenom(rollappId)
	allocatedAmount := simtypes.RandomAmount(r, math.NewInt(1e9).SubRaw(100_000)).AddRaw(100_000).MulRaw(1e18)
	allocation := sdk.Coin{
		Denom:  baseDenom,
		Amount: allocatedAmount,
	}

	// Generate random bonding curve
	curve := generateRandomBondingCurve(r, allocatedAmount)
	plan := types.NewPlan(id, rollappId, allocation, curve, startTime, preLaunchTime, types.DefaultIncentivePlanParams())

	// randomize starting sold amount
	// minSoldAmt < soldAmt < allocatedAmount - minUnsoldAmt
	minSoldAmt := math.NewInt(1).MulRaw(1e18) // 1 token minimum
	minUnsoldAmt := math.NewInt(100).MulRaw(1e18)
	soldAmt := simtypes.RandomAmount(r, allocatedAmount.Sub(minUnsoldAmt).Sub(minSoldAmt)).Add(minSoldAmt)
	plan.SoldAmt = soldAmt

	return plan
}

func generateRandomBondingCurve(r *rand.Rand, allocatedAmount math.Int) types.BondingCurve {
	// Generate 0.5 < N < 1.5 with maximum precision of 3
	nInt := r.Int63n(1000)                     // Generate a random integer between 0 and 999
	n := math.LegacyNewDecWithPrec(nInt, 3)    // Convert to decimal with 3 decimal places
	n = n.Add(math.LegacyNewDecWithPrec(5, 1)) // Add 0.5

	// targetRaiseDYM between 10K and 100M DYM
	targetRaiseDYM := simtypes.RandomAmount(r, math.NewInt(1e8)).AddRaw(10_000)

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

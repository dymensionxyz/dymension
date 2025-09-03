package keeper_test

import (
	"slices"

	"cosmossdk.io/math"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (s *KeeperTestSuite) TestShouldPump() {
	b, err := s.App.StreamerKeeper.EpochBlocks(s.Ctx, "day")
	s.Require().NoError(err)

	pumpNum := uint64(9000)

	s.Run("GenerateUnifiedRandom", func() {
		// Pump hash
		ctx := hashPump(s.Ctx)
		r1 := math.NewIntFromBigIntMut(
			keeper.GenerateUnifiedRandomModInt(ctx, b.BigIntMut(), nil),
		) //  7639

		// No pump hash
		ctx = hashNoPump(s.Ctx)
		r2 := math.NewIntFromBigIntMut(
			keeper.GenerateUnifiedRandomModInt(ctx, b.BigIntMut(), nil),
		) //  11118

		middle := math.NewIntFromUint64(pumpNum)

		s.Require().True(r1.LT(middle), "expected r1 < middle, got: %s < %s", r1, middle)
		s.Require().True(middle.LT(r2), "expected middle < r2, got: %s < %s ", middle, r2)
	})

	s.Run("ShouldPump", func() {
		// Pump hash should pump
		ctx := hashPump(s.Ctx)
		pumpAmt, err := keeper.ShouldPump(
			ctx,
			types.PumpParams{
				NumTopRollapps:  0,
				EpochBudget:     commontypes.DYM.MulRaw(10),
				EpochBudgetLeft: commontypes.DYM.MulRaw(10),
				NumPumps:        pumpNum,
				PumpDistr:       types.PumpDistr_PUMP_DISTR_UNIFORM,
			},
			b,
		)
		s.Require().NoError(err)
		s.Require().False(pumpAmt.IsZero())
		expectedAmt := math.NewInt(2040061966151279)
		s.Require().True(expectedAmt.Equal(pumpAmt), "expected %s, got: %s", expectedAmt, pumpAmt)

		// No pump hash should not pump
		ctx = hashNoPump(s.Ctx)
		pumpAmt, err = keeper.ShouldPump(
			ctx,
			types.PumpParams{
				NumTopRollapps:  0,
				EpochBudget:     commontypes.DYM.MulRaw(10),
				EpochBudgetLeft: commontypes.DYM.MulRaw(10),
				NumPumps:        pumpNum,
				PumpDistr:       types.PumpDistr_PUMP_DISTR_UNIFORM,
			},
			b,
		)
		s.Require().NoError(err)
		s.Require().True(pumpAmt.IsZero())
	})
}

func (s *KeeperTestSuite) TestUniformRandom() {
	s.T().Skip("This test is for debugging and visualizing the distribution.")
	ctx := hashPump(s.Ctx)

	const iterations = 250

	modulo := math.NewInt(10_000)
	values := make([]uint64, 0, iterations)
	total := math.ZeroInt()

	for iteration := 0; iteration < iterations; iteration++ {
		hash := ctx.HeaderInfo().Hash
		newHash := nextPermutation([32]byte(hash), iteration)
		headerInfo := ctx.HeaderInfo()
		headerInfo.Hash = newHash[:]
		ctx = ctx.WithHeaderInfo(headerInfo)

		random := keeper.GenerateUnifiedRandomModInt(ctx, modulo.BigInt(), nil)
		total = total.Add(math.NewIntFromBigIntMut(random))
		values = append(values, random.Uint64())
	}

	slices.Sort(values)
	for _, v := range values {
		println(v)
	}
	s.T().Log("Target mean", modulo.QuoRaw(2))
	s.T().Log("Actual mean", total.QuoRaw(iterations))
}

func (s *KeeperTestSuite) TestExpRandom() {
	s.T().Skip("This test is for debugging and visualizing the distribution.")
	ctx := hashPump(s.Ctx)

	const iterations = 250

	budget := math.NewInt(100_000)
	pumpNum := math.NewInt(100)
	values := make([]uint64, 0, iterations)
	total := math.ZeroInt()

	for iteration := 0; iteration < iterations; iteration++ {
		hash := ctx.HeaderInfo().Hash
		newHash := nextPermutation([32]byte(hash), iteration)
		headerInfo := ctx.HeaderInfo()
		headerInfo.Hash = newHash[:]
		ctx = ctx.WithHeaderInfo(headerInfo)

		random := keeper.GenerateExpRandomLambdaInt(ctx, pumpNum.BigInt(), budget.BigInt(), nil)
		total = total.Add(math.NewIntFromBigIntMut(random))
		values = append(values, random.Uint64())
	}

	slices.Sort(values)
	for _, v := range values {
		println(v)
	}
	s.T().Log("Target mean", budget.Quo(pumpNum))
	s.T().Log("Actual mean", total.QuoRaw(iterations))
}

func (s *KeeperTestSuite) TestPumpAmtSamplesUniform() {
	s.T().Skip("This test is for debugging and visualizing the distribution.")

	var (
		epochBudget     = math.NewInt(200_000)
		epochBudgetLeft = epochBudget
		pumpNum         = int64(200)
		ctx             = hashPump(s.Ctx)
		pumpFunc        = types.PumpDistr_PUMP_DISTR_UNIFORM
	)

	values := make([]math.Int, 0, pumpNum)
	total := math.ZeroInt()

	for iteration := int64(0); iteration < pumpNum; iteration++ {
		hash := ctx.HeaderInfo().Hash
		newHash := nextPermutation([32]byte(hash), int(iteration))
		headerInfo := ctx.HeaderInfo()
		headerInfo.Hash = newHash[:]
		ctx = ctx.WithHeaderInfo(headerInfo)

		pumpAmt, err := keeper.PumpAmt(ctx, types.PumpParams{
			NumTopRollapps:  0,
			EpochBudget:     epochBudget,
			EpochBudgetLeft: epochBudgetLeft,
			NumPumps:        uint64(pumpNum),
			PumpDistr:       pumpFunc,
		})
		s.Require().NoError(err)

		epochBudgetLeft = epochBudgetLeft.Sub(pumpAmt)
		total = total.Add(pumpAmt)
		values = append(values, pumpAmt)
	}

	valuesCpy := make([]math.Int, len(values))
	copy(valuesCpy, values)
	slices.SortFunc(values, func(a, b math.Int) int {
		if a.LT(b) {
			return -1
		}
		if a.GT(b) {
			return 1
		}
		return 0
	})

	s.T().Log("Sorted samples – CDF function")
	for _, v := range values {
		println(v.String())
	}

	s.T().Log("Not sorted samples")
	for _, v := range valuesCpy {
		println(v.String())
	}

	s.T().Log("Target mean", epochBudget.QuoRaw(pumpNum))
	s.T().Log("Actual mean", total.QuoRaw(pumpNum))
	s.T().Log("Total distr", total)
}

func nextPermutation(currentHash [32]byte, seed int) [32]byte {
	for i := 0; i < 32; i++ {
		currentHash[i] ^= byte((seed + i*7) % 256)
	}
	return currentHash
}

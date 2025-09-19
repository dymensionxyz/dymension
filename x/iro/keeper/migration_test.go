package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

func (s *KeeperTestSuite) setV1IRO(rollappId string) types.Plan {
	var (
		allocation        = sdk.NewCoin(types.IRODenom(rollappId), math.NewInt(100_000).MulRaw(1e18))
		curve             = types.DefaultBondingCurve()
		defaultIncentives = types.DefaultIncentivePlanParams()
	)

	// simulate old IRO plan
	curve.LiquidityDenomDecimals = 0
	curve.RollappDenomDecimals = 0

	id := s.App.IROKeeper.GetNextPlanIdAndIncrement(s.Ctx)
	plan := types.NewPlan(id, rollappId, "", allocation, curve, 0, defaultIncentives, math.LegacyOneDec(), 0, 0)
	plan.MaxAmountToSell = math.ZeroInt()
	plan.LiquidityPart = math.LegacyDec{}
	plan.VestingPlan = types.IROVestingPlan{}

	s.App.IROKeeper.SetPlan(s.Ctx, plan)
	return plan
}

func (s *KeeperTestSuite) TestMigrate1to2() {
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	// set simple IRO plan
	_ = s.setV1IRO("rollapp1")

	// set settled IRO plan
	rollappDenom := "rollappSettledDenom"
	plan := s.setV1IRO("rollapp2")
	plan.SettledDenom = rollappDenom
	s.App.IROKeeper.SetPlan(s.Ctx, plan)

	// create balancer pool to simulate settled IRO plan
	poolId := s.PreparePoolWithCoins(sdk.NewCoins(
		sdk.NewCoin(rollappDenom, math.NewInt(1_000_000).MulRaw(1e18)),
		sdk.NewCoin("adym", math.NewInt(1_000_000).MulRaw(1e18)),
	))
	// assert fee token created
	_, err := s.App.TxFeesKeeper.GetFeeToken(s.Ctx, rollappDenom)
	s.Require().NoError(err)

	// set almost 100% sold IRO plan
	almostSoldPlan := s.setV1IRO("rollapp3")
	// Set SoldAmt equal to MaxAmountToSell to simulate fully sold IRO
	// In the migration, MaxAmountToSell will be calculated from equilibrium
	almostSoldPlan.SoldAmt = almostSoldPlan.TotalAllocation.Amount.Mul(math.NewInt(9)).Quo(math.NewInt(10)) // 90% of total allocation
	s.App.IROKeeper.SetPlan(s.Ctx, almostSoldPlan)

	// fund the plan with tokens to allow graduation
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(almostSoldPlan.GetIRODenom(), almostSoldPlan.TotalAllocation.Amount))) // rollapp tokens
	s.FundAcc(almostSoldPlan.GetAddress(), sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(1_000).MulRaw(1e18))))                        // raised liquidity tokens

	s.Ctx = s.Ctx.WithBlockTime(time.Now())
	m := keeper.NewMigrator(*s.App.IROKeeper)
	err = m.Migrate1to2(s.Ctx)
	s.Require().NoError(err)

	// check plan1
	plan, ok := s.App.IROKeeper.GetPlanByRollapp(s.Ctx, "rollapp1")
	s.Require().True(ok, "plan should be found")

	s.Require().False(plan.MaxAmountToSell.IsZero(), "max amount to sell not set correctly")
	s.Require().True(plan.LiquidityPart.Equal(math.LegacyOneDec()), "liquidity part not set correctly")
	s.Require().NotEmpty(plan.LiquidityDenom, "liquidity denom not set correctly")
	s.Require().NotZero(plan.BondingCurve.LiquidityDenomDecimals, "bonding curve liquidity denom decimals not set correctly")
	s.Require().NotZero(plan.BondingCurve.RollappDenomDecimals, "bonding curve rollapp denom decimals not set correctly")
	s.Require().True(plan.TradingEnabled, "trading enabled not set correctly")

	// check poolID set for plan2
	plan, ok = s.App.IROKeeper.GetPlanByRollapp(s.Ctx, "rollapp2")
	s.Require().True(ok, "plan should be found")
	s.Require().True(plan.IsSettled(), "plan should be settled")
	s.Require().Equal(poolId, plan.GraduatedPoolId, "poolID not set correctly")
	poolDenoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, poolId)
	s.Require().NoError(err)
	s.Require().ElementsMatch(poolDenoms, []string{plan.SettledDenom, "adym"})

	// check rollapp3 is graduated
	plan, ok = s.App.IROKeeper.GetPlanByRollapp(s.Ctx, "rollapp3")
	s.Require().True(ok, "plan should be found")
	s.Require().True(plan.IsGraduated(), "plan should be graduated")
	s.Require().NotZero(plan.GraduatedPoolId, "poolID not set correctly")
	poolDenoms, err = s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, plan.GraduatedPoolId)
	s.Require().NoError(err)
	s.Require().ElementsMatch(poolDenoms, []string{plan.GetIRODenom(), "adym"})

	// check incentives are set
	gauges, err := s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(plan.GraduatedPoolId))
	s.Require().NoError(err)
	s.Require().Len(gauges, 2)
}

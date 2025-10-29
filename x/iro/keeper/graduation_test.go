package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	testutil "github.com/dymensionxyz/dymension/v3/testutil/math"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

// This test:
// creates custom IRO plan
// buy all the tokens and asserts that the plan is graduated
// settles the graudauted plan and asserts that the pool is updated with the settled denom
func (s *KeeperTestSuite) TestGraduatePlan() {
	curve := types.BondingCurve{
		M:                      math.LegacyMustNewDecFromStr("0"),
		N:                      math.LegacyMustNewDecFromStr("1"),
		C:                      math.LegacyMustNewDecFromStr("0.1"), // each token costs 0.1 DYM
		RollappDenomDecimals:   18,
		LiquidityDenomDecimals: 18,
	}

	startTime := time.Now()
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	rollappId := s.CreateDefaultRollapp()
	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	k := s.App.IROKeeper

	// Create IRO plan
	allocation := math.NewInt(1_000_000).MulRaw(1e18)
	liquidityPart := types.DefaultParams().MinLiquidityPart
	apptesting.FundAccount(s.App, s.Ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), sdk.NewCoins(sdk.NewCoin("adym", k.GetParams(s.Ctx).CreationFee)))

	eq := types.FindEquilibrium(curve, allocation, liquidityPart)
	planId, err := k.CreatePlan(s.Ctx, "adym", allocation, eq, time.Hour, startTime, true, false, rollapp, curve, types.DefaultIncentivePlanParams(), liquidityPart, time.Hour, 0)
	s.Require().NoError(err)
	plan := k.MustGetPlan(s.Ctx, planId)

	// Buy all tokens
	buyAmt := plan.MaxAmountToSell.Sub(plan.SoldAmt)
	buyer := sample.Acc()
	s.BuySomeTokens(planId, buyer, buyAmt)

	// assert we created a pool
	plan = k.MustGetPlan(s.Ctx, planId)
	s.Require().True(plan.IsGraduated())
	s.Require().False(plan.IsSettled())

	// assert ~all rollapptokens were used for the pool
	tokensLeftovers := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), plan.GetIRODenom())
	s.Require().NoError(testutil.ApproxEqual(tokensLeftovers.Amount, math.ZeroInt(), math.NewInt(1e18)), "not all tokens were used for the pool: %s", tokensLeftovers.String())

	// assert rollapp is launchable
	rollapp = s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	s.Require().True(rollapp.PreLaunchTime.Equal(s.Ctx.BlockTime()))

	// Assert liquidity pool
	poolId := plan.GraduatedPoolId
	pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
	s.Require().NoError(err)

	// The expected DYM in the pool is the amount to sell * 0.1 (fixed price) + 0.1 DYM creation fee
	expectedRaisedLiquidity := plan.MaxAmountToSell.ToLegacyDec().Mul(curve.C).TruncateInt()
	expectedDYMInPool := expectedRaisedLiquidity.ToLegacyDec().Mul(liquidityPart).TruncateInt()
	expectedTokensInPool := expectedDYMInPool.ToLegacyDec().Quo(curve.C).TruncateInt()

	poolCoins := pool.GetTotalPoolLiquidity(s.Ctx)
	s.Require().Equal(expectedDYMInPool.String(), poolCoins.AmountOf("adym").String())
	s.Require().Equal(expectedTokensInPool.String(), poolCoins.AmountOf(plan.GetIRODenom()).String(), "poolCoins: %s", poolCoins.String())

	// Assert pool price
	lastIROPrice := plan.SpotPrice()
	price, err := pool.SpotPrice(s.Ctx, "adym", plan.GetIRODenom())
	s.Require().NoError(err)
	s.Require().Equal(lastIROPrice, price)

	// Assert incentives (only perpetual gauge due to pool creation, no custom gauge with leftovers)
	gauges, err := s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(poolId))
	s.Assert().NoError(err)
	s.Require().Len(gauges, 1)
	gauge := gauges[0]
	s.Require().True(gauge.IsPerpetual)
	s.Require().True(gauge.Coins.Empty(), "incentives are not empty: %s", gauge.Coins.String())

	// Settle
	rollappDenom := "dasdasdasdasdsa"
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, allocation)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)
	plan = k.MustGetPlan(s.Ctx, planId)
	s.Require().True(plan.IsSettled())

	// assert pool is updated with the settled denom
	pool, err = s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
	s.Require().NoError(err)
	poolCoins = pool.GetTotalPoolLiquidity(s.Ctx)
	s.Require().Equal(expectedTokensInPool.String(), poolCoins.AmountOf(rollappDenom).String(), "poolCoins: %s", poolCoins.String())

	// assert no change in incentives
	gauges, err = s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(poolId))
	s.Assert().NoError(err)
	s.Require().Len(gauges, 1)

	// assert all unclaimable iro denom tokens are burned
	iroCoins := s.App.BankKeeper.GetSupply(s.Ctx, plan.GetIRODenom())
	expectedIROCoins := plan.SoldAmt.Sub(plan.ClaimedAmt)
	s.Require().Equal(expectedIROCoins.String(), iroCoins.Amount.String())

	// FIXME: assert fee token is updated
}

// This test:
// creates fair launch IRO plan
// buy all the tokens and asserts that the plan is graduated, and the liquidity pool is according to the target raise
// settles the graudauted plan and asserts that the pool is updated with the settled denom
func (s *KeeperTestSuite) TestGraduateStandardLaunchPlan() {
	k := s.App.IROKeeper
	iroParams := k.GetParams(s.Ctx)

	s.App.BankKeeper.SetDenomMetaData(s.Ctx, dymDenomMetadata)
	gammParams := s.App.GAMMKeeper.GetParams(s.Ctx)
	gammParams.AllowedPoolCreationDenoms = append(gammParams.AllowedPoolCreationDenoms, "adym")
	s.App.GAMMKeeper.SetParams(s.Ctx, gammParams)

	rollappId := s.CreateStandardLaunchRollapp()
	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	owner := rollapp.Owner

	// Fund owner with liquidity denom for creation fee
	s.FundAcc(sdk.MustAccAddressFromBech32(owner), sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18))))

	res, err := s.msgServer.CreateStandardLaunchPlan(s.Ctx, &types.MsgCreateStandardLaunchPlan{
		RollappId:      rollappId,
		Owner:          owner,
		TradingEnabled: true,
		LiquidityDenom: "adym",
	})
	s.Require().NoError(err)
	planId := res.PlanId

	// Buy all tokens
	plan := k.MustGetPlan(s.Ctx, planId)
	buyAmt := plan.MaxAmountToSell.Sub(plan.SoldAmt)
	buyer := sample.Acc()
	s.BuySomeTokens(planId, buyer, buyAmt)

	// assert we created a pool
	plan = k.MustGetPlan(s.Ctx, planId)
	s.Require().True(plan.IsGraduated())
	s.Require().False(plan.IsSettled())

	// assert ~all rollapptokens were used for the pool
	tokensLeftovers := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), plan.GetIRODenom())
	s.Require().NoError(testutil.ApproxEqual(tokensLeftovers.Amount, math.ZeroInt(), math.NewInt(1e18)), "not all tokens were used for the pool: %s", tokensLeftovers.String())

	// assert rollapp is launchable
	rollapp = s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	s.Require().True(rollapp.PreLaunchTime.Equal(s.Ctx.BlockTime()))

	// Assert liquidity pool
	poolId := plan.GraduatedPoolId
	pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
	s.Require().NoError(err)

	// The expected DYM in the pool is the target raise (1% tolerance)
	expectedDYMInPool := iroParams.StandardLaunch.TargetRaise.Amount
	poolCoins := pool.GetTotalPoolLiquidity(s.Ctx)
	s.Require().NoError(testutil.ApproxEqualRatio(expectedDYMInPool, poolCoins.AmountOf("adym"), 0.01))

	rollappTokensInPool := poolCoins.AmountOf(plan.GetIRODenom())
	s.Require().NoError(testutil.ApproxEqualRatio(plan.TotalAllocation.Amount.Sub(plan.MaxAmountToSell), rollappTokensInPool, 0.01))

	// Assert pool price
	lastIROPrice := plan.SpotPrice()
	price, err := pool.SpotPrice(s.Ctx, "adym", plan.GetIRODenom())
	s.Require().NoError(err)
	s.Require().Equal(lastIROPrice, price)

	// Assert incentives (only perpetual gauge due to pool creation, no custom gauge with leftovers)
	gauges, err := s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(poolId))
	s.Assert().NoError(err)
	s.Require().Len(gauges, 1)
	gauge := gauges[0]
	s.Require().True(gauge.IsPerpetual)
	s.Require().True(gauge.Coins.Empty(), "incentives are not empty: %s", gauge.Coins.String())

	// Settle
	rollappDenom := "dasdasdasdasdsa"
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, iroParams.StandardLaunch.AllocationAmount)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)
	plan = k.MustGetPlan(s.Ctx, planId)
	s.Require().True(plan.IsSettled())

	// assert pool is updated with the settled denom
	pool, err = s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
	s.Require().NoError(err)
	poolCoins = pool.GetTotalPoolLiquidity(s.Ctx)
	s.Require().Equal(poolCoins.AmountOf(rollappDenom), rollappTokensInPool)

	// assert no change in incentives
	gauges, err = s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(poolId))
	s.Assert().NoError(err)
	s.Require().Len(gauges, 1)

	// assert all unclaimable iro denom tokens are burned
	iroCoins := s.App.BankKeeper.GetSupply(s.Ctx, plan.GetIRODenom())
	expectedIROCoins := plan.SoldAmt.Sub(plan.ClaimedAmt)
	s.Require().Equal(expectedIROCoins.String(), iroCoins.Amount.String())

	// FIXME: assert fee token is updated
}

func (s *KeeperTestSuite) TestGraduationGasFree() {
	startTime := time.Now()
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	k := s.App.IROKeeper

	rollappId := s.CreateDefaultRollapp()
	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	s.FundAcc(sdk.MustAccAddressFromBech32(rollapp.Owner), sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(1_000_000_000).MulRaw(1e18))))

	buyer := sample.Acc()
	s.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(1_000_000_000).MulRaw(1e18))))

	// Create IRO plan
	allocation := math.NewInt(1_000_000).MulRaw(1e18)
	liquidityPart := types.DefaultParams().MinLiquidityPart
	curve := types.BondingCurve{
		M:                      math.LegacyMustNewDecFromStr("0"),
		N:                      math.LegacyMustNewDecFromStr("1"),
		C:                      math.LegacyMustNewDecFromStr("0.1"),
		RollappDenomDecimals:   18,
		LiquidityDenomDecimals: 18,
	}
	eq := types.FindEquilibrium(curve, allocation, liquidityPart)
	planId, err := k.CreatePlan(s.Ctx, "adym", allocation, eq, time.Hour, startTime, true, false, rollapp, curve, types.DefaultIncentivePlanParams(), liquidityPart, time.Hour, 0)
	s.Require().NoError(err)

	// check how much gas is consumed by standard buy
	alreadySpent := s.Ctx.GasMeter().GasConsumed()
	_, err = k.Buy(s.Ctx, planId, buyer, math.NewInt(500_000).MulRaw(1e18), math.NewInt(1_000_000_000).MulRaw(1e18))
	s.Require().NoError(err)
	gasCheckpoint := s.Ctx.GasMeter().GasConsumed()
	buyGasCost := gasCheckpoint - alreadySpent
	s.Require().Greater(buyGasCost, uint64(0))

	// Buy all left tokens
	plan := k.MustGetPlan(s.Ctx, planId)
	buyAmt := plan.MaxAmountToSell.Sub(plan.SoldAmt)
	_, err = k.Buy(s.Ctx, planId, buyer, buyAmt, math.NewInt(1_000_000_000).MulRaw(1e18))
	s.Require().NoError(err)

	// assert that the gas cost is the same within 5%
	buyWithGraduationGasCost := s.Ctx.GasMeter().GasConsumed() - gasCheckpoint
	s.Require().NoError(testutil.ApproxEqualRatio(math.NewIntFromUint64(buyGasCost), math.NewIntFromUint64(buyWithGraduationGasCost), 0.05))
}

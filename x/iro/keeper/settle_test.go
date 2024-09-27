package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	keeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func (s *KeeperTestSuite) TestSettle() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	endTime := startTime.Add(time.Hour)
	amt := sdk.NewInt(1_000_000).MulRaw(1e18)
	rollappDenom := "dasdasdasdasdsa"

	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, endTime, rollapp, curve, incentives)
	s.Require().NoError(err)
	planDenom := k.MustGetPlan(s.Ctx, planId).TotalAllocation.Denom

	// assert initial FUT balance
	balance := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().Equal(amt, balance.Amount)

	// buy some tokens
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	soldAmt := sdk.NewInt(1_000).MulRaw(1e18)
	s.BuySomeTokens(planId, sample.Acc(), soldAmt)

	// settle should fail as no rollappDenom balance available
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().Error(err)

	// should succeed after fund
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	// settle again should fail as already settled
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().Error(err)

	// assert no FUT balance in the account
	balance = s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().True(balance.IsZero())

	// assert sold amount is kept in the account and not used for liquidity pool
	balance = s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), rollappDenom)
	s.Require().Equal(soldAmt, balance.Amount)
}

// Test liquidity pool bootstrap
func (s *KeeperTestSuite) TestBootstrapLiquidityPool() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	allocation := sdk.NewInt(1_000_000).MulRaw(1e18)
	maxAmt := sdk.NewInt(1_000_000_000).MulRaw(1e18)
	rollappDenom := "dasdasdasdasdsa"

	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)

	// create IRO plan
	apptesting.FundAccount(s.App, s.Ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), sdk.NewCoins(sdk.NewCoin(appparams.BaseDenom, k.GetParams(s.Ctx).CreationFee)))
	planId, err := k.CreatePlan(s.Ctx, allocation, startTime, startTime.Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)

	// buy some tokens
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", maxAmt))
	s.FundAcc(buyer, buyersFunds)

	err = k.Buy(s.Ctx, planId, buyer, sdk.NewInt(1_000).MulRaw(1e18), maxAmt)
	s.Require().NoError(err)

	plan := k.MustGetPlan(s.Ctx, planId)
	raisedDYM := k.BK.GetBalance(s.Ctx, plan.GetAddress(), appparams.BaseDenom)
	preSettleCoins := sdk.NewCoins(raisedDYM, sdk.NewCoin(rollappDenom, allocation.Sub(plan.SoldAmt)))

	// settle should succeed after fund
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, allocation)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	/* -------------------------- assert liquidity pool ------------------------- */
	// pool created
	expectedPoolID := uint64(1)
	pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, expectedPoolID)
	s.Require().NoError(err)

	// pool price should be the same as the last price of the plan
	price, err := pool.SpotPrice(s.Ctx, "adym", rollappDenom)
	s.Require().NoError(err)

	plan = k.MustGetPlan(s.Ctx, planId)
	lastPrice := plan.SpotPrice()
	s.Require().Equal(lastPrice, price)

	// assert incentives
	poolCoins := pool.GetTotalPoolLiquidity(s.Ctx)
	gauges, err := s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(expectedPoolID))
	s.Require().NoError(err)
	found := false
	gauge := incentivestypes.Gauge{}
	for _, gauge = range gauges {
		if !gauge.IsPerpetual {
			found = true
			break
		}
	}
	s.Require().True(found)
	s.Require().False(gauge.Coins.IsZero())

	// expected tokens for incentives:
	// 		raisedDYM - poolCoins
	// 		totalAllocation - soldAmt - poolCoins
	expectedIncentives := preSettleCoins.Sub(poolCoins...)
	s.Assert().Equal(expectedIncentives, gauge.Coins)
}

func (s *KeeperTestSuite) TestSettleNothingSold() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	endTime := startTime.Add(time.Hour)
	amt := sdk.NewInt(1_000_000).MulRaw(1e18)
	rollappDenom := "rollapp_denom"

	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	_, err := k.CreatePlan(s.Ctx, amt, startTime, endTime, rollapp, curve, incentives)
	s.Require().NoError(err)
	// planDenom := k.MustGetPlan(s.Ctx, planId).TotalAllocation.Denom

	// Settle without any tokens sold
	s.Ctx = s.Ctx.WithBlockTime(endTime.Add(time.Minute))
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	/* -------------------------- assert liquidity pool ------------------------- */
	// pool created
	expectedPoolID := uint64(1)
	pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, expectedPoolID)
	s.Require().NoError(err)
	poolCoins := pool.GetTotalPoolLiquidity(s.Ctx)
	poolCoins.AmountOf("adym").Equal(s.App.IROKeeper.GetParams(s.Ctx).CreationFee)

	// incentives expected to have zero coins
	gauges, err := s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(expectedPoolID))
	s.Require().NoError(err)
	found := false
	gauge := incentivestypes.Gauge{}
	for _, gauge = range gauges {
		if !gauge.IsPerpetual {
			found = true
			break
		}
	}
	s.Require().True(found)
	s.Require().True(gauge.Coins.IsZero())
}

func (s *KeeperTestSuite) TestSettleAllSold() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	// setting curve with fixed price
	curve := types.BondingCurve{
		M: math.LegacyMustNewDecFromStr("0"),
		N: math.LegacyMustNewDecFromStr("1"),
		C: math.LegacyMustNewDecFromStr("0.00001"),
	}
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	endTime := startTime.Add(time.Hour)
	amt := sdk.NewInt(1_000_000).MulRaw(1e18)
	rollappDenom := "rollapp_denom"

	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, endTime, rollapp, curve, incentives)
	s.Require().NoError(err)

	// Buy all possible tokens
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	buyer := sample.Acc()
	buyAmt := amt.ToLegacyDec().Mul(keeper.AllocationSellLimit).TruncateInt()
	s.BuySomeTokens(planId, buyer, buyAmt)

	// Settle
	s.Ctx = s.Ctx.WithBlockTime(endTime.Add(time.Minute))
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	plan := k.MustGetPlan(s.Ctx, planId)

	pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, 1)
	s.Require().NoError(err)

	gauges, err := s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(1))
	s.Require().NoError(err)
	found := false
	gauge := incentivestypes.Gauge{}
	for _, gauge = range gauges {
		if !gauge.IsPerpetual {
			found = true
			break
		}
	}
	s.Require().True(found)

	// only few RA tokens left, so the pool should be quite small
	// most of the dym should be as incentive
	s.T().Log("Pool coins", pool.GetTotalPoolLiquidity(s.Ctx))
	s.T().Log("Gauge coins", gauge.Coins)
	s.Require().True(pool.GetTotalPoolLiquidity(s.Ctx).AmountOf("adym").LT(gauge.Coins.AmountOf("adym")))
	s.Require().Equal(pool.GetTotalPoolLiquidity(s.Ctx).AmountOf(plan.SettledDenom), amt.Sub(buyAmt))
}

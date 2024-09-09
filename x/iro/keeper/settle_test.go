package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

func (s *KeeperTestSuite) TestSettle() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	endTime := startTime.Add(time.Hour)
	amt := sdk.NewInt(1_000_000)
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
	soldAmt := sdk.NewInt(1_000)
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

func (s *KeeperTestSuite) TestClaim() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()
	rollappDenom := "dasdasdasdasdsa"

	startTime := time.Now()
	amt := sdk.NewInt(1_000_000)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, startTime.Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)
	planDenom := k.MustGetPlan(s.Ctx, planId).TotalAllocation.Denom
	balance := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().Equal(amt, balance.Amount)

	claimer := sample.Acc()
	// buy some tokens
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	soldAmt := sdk.NewInt(1_000)
	s.BuySomeTokens(planId, claimer, soldAmt)

	// claim should fail as not settled
	err = k.Claim(s.Ctx, planId, claimer)
	s.Require().Error(err)

	// settle
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	// claim should fail as no balance available (random address)
	err = k.Claim(s.Ctx, planId, sample.Acc())
	s.Require().Error(err)

	// fund. claim should succeed
	err = k.Claim(s.Ctx, planId, claimer)
	s.Require().NoError(err)

	// assert claimed amt
	balance = s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().True(balance.IsZero())
	balance = s.App.BankKeeper.GetBalance(s.Ctx, claimer, rollappDenom)
	s.Require().Equal(soldAmt, balance.Amount)
}

// Test liquidity pool bootstrap
func (s *KeeperTestSuite) TestBootstrapLiquidityPool() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	amt := sdk.NewInt(1_000_000)
	rollappDenom := "dasdasdasdasdsa"

	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)

	// create IRO plan
	apptesting.FundAccount(s.App, s.Ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), sdk.NewCoins(sdk.NewCoin(appparams.BaseDenom, k.GetParams(s.Ctx).CreationFee)))
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, startTime.Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)

	// buy some tokens
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(100_000)))
	s.FundAcc(buyer, buyersFunds)

	err = k.Buy(s.Ctx, planId, buyer, sdk.NewInt(1_000), sdk.NewInt(100_000))
	s.Require().NoError(err)

	plan := k.MustGetPlan(s.Ctx, planId)
	raisedDYM := k.BK.GetBalance(s.Ctx, plan.GetAddress(), appparams.BaseDenom)
	preSettleCoins := sdk.NewCoins(raisedDYM, sdk.NewCoin(rollappDenom, amt.Sub(plan.SoldAmt)))

	// settle should succeed after fund
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
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
	lastPrice := plan.BondingCurve.SpotPrice(plan.SoldAmt)
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

// test edge cases: nothing sold, all sold

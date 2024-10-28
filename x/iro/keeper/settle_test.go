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

func (s *KeeperTestSuite) TestBootstrapLiquidityPool() {
	curve := types.BondingCurve{
		M: math.LegacyMustNewDecFromStr("0"),
		N: math.LegacyMustNewDecFromStr("1"),
		C: math.LegacyMustNewDecFromStr("0.1"), // each token costs 0.1 DYM
	}

	startTime := time.Now()
	allocation := sdk.NewInt(1_000_000).MulRaw(1e18)
	rollappDenom := "dasdasdasdasdsa"

	testCases := []struct {
		name           string
		buyAmt         math.Int
		expectedDYM    math.Int
		expectedTokens math.Int
	}{
		// for small purchases, the raised dym is the limiting factor:
		// - the expected DYM in the pool is the buy amount * 0.1 (fixed price) + 10 DYM creation fee
		// for large purchases, the left tokens are the limiting factor:
		// - the expected DYM in the pool is the left tokens / 0.1 (fixed price)
		{
			name:           "Small purchase",
			buyAmt:         math.NewInt(1_000).MulRaw(1e18),
			expectedDYM:    math.NewInt(110).MulRaw(1e18),
			expectedTokens: math.NewInt(1_100).MulRaw(1e18),
		},
		{
			name:           "Large purchase - left tokens are limiting factor",
			buyAmt:         math.NewInt(800_000).MulRaw(1e18),
			expectedDYM:    math.NewInt(20_000).MulRaw(1e18),
			expectedTokens: math.NewInt(200_000).MulRaw(1e18),
		},
		{
			name:           "Nothing sold - pool contains only creation fee",
			buyAmt:         math.NewInt(0),
			expectedDYM:    math.NewInt(10).MulRaw(1e18), // creation fee
			expectedTokens: math.NewInt(100).MulRaw(1e18),
		},
		{
			name:           "All sold - pool contains only reserved tokens",
			buyAmt:         math.NewInt(999_999).MulRaw(1e18),
			expectedDYM:    math.NewInt(1).MulRaw(1e17), // 0.1 DYM
			expectedTokens: math.NewInt(1).MulRaw(1e18), // reserved tokens
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset the test state for each test case
			rollappId := s.CreateDefaultRollapp()
			rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
			k := s.App.IROKeeper

			// Create IRO plan
			apptesting.FundAccount(s.App, s.Ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), sdk.NewCoins(sdk.NewCoin(appparams.BaseDenom, k.GetParams(s.Ctx).CreationFee)))
			planId, err := k.CreatePlan(s.Ctx, allocation, startTime, startTime.Add(time.Hour), rollapp, curve, types.DefaultIncentivePlanParams())
			s.Require().NoError(err)
			reservedTokens := k.MustGetPlan(s.Ctx, planId).SoldAmt

			// Buy tokens
			if tc.buyAmt.GT(math.ZeroInt()) {
				s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
				buyer := sample.Acc()
				s.BuySomeTokens(planId, buyer, tc.buyAmt)
			}

			plan := k.MustGetPlan(s.Ctx, planId)
			raisedDYM := k.BK.GetBalance(s.Ctx, plan.GetAddress(), appparams.BaseDenom)
			unallocatedTokensAmt := allocation.Sub(plan.SoldAmt).Add(reservedTokens)

			// Settle
			s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, allocation)))
			err = k.Settle(s.Ctx, rollappId, rollappDenom)
			s.Require().NoError(err)

			// Assert liquidity pool
			poolId := uint64(1)
			pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			s.Require().NoError(err)

			poolCoins := pool.GetTotalPoolLiquidity(s.Ctx)
			s.Require().Equal(tc.expectedDYM, poolCoins.AmountOf("adym"))
			s.Require().Equal(tc.expectedTokens, poolCoins.AmountOf(rollappDenom))

			// Assert pool price
			lastIROPrice := plan.SpotPrice()
			price, err := pool.SpotPrice(s.Ctx, "adym", rollappDenom)
			s.Require().NoError(err)
			s.Require().Equal(lastIROPrice, price)

			// Assert incentives
			gauges, err := s.App.IncentivesKeeper.GetGaugesForDenom(s.Ctx, gammtypes.GetPoolShareDenom(poolId))
			s.Require().NoError(err)
			found := false
			var gauge incentivestypes.Gauge
			for _, g := range gauges {
				if !g.IsPerpetual {
					found = true
					gauge = g
					break
				}
			}
			s.Require().True(found)

			// expected tokens for incentives:
			// 		raisedDYM - poolCoins
			// 		unallocatedTokens - poolCoins
			expectedIncentives := sdk.NewCoins(raisedDYM, sdk.NewCoin(rollappDenom, unallocatedTokensAmt)).Sub(poolCoins...)
			s.Assert().Equal(expectedIncentives, gauge.Coins)
		})
	}
}

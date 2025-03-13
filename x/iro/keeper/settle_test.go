package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
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
	amt := math.NewInt(1_000_000).MulRaw(1e18)
	rollappDenom := "dasdasdasdasdsa"
	liquidityPart := types.DefaultParams().MinLiquidityPart

	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, "adym", amt, time.Hour, startTime, true, rollapp, curve, incentives, liquidityPart, time.Hour, 0)
	s.Require().NoError(err)
	planDenom := k.MustGetPlan(s.Ctx, planId).TotalAllocation.Denom

	// assert initial IRO balance
	balance := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().Equal(amt, balance.Amount)

	// buy some tokens
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	soldAmt := math.NewInt(1_000).MulRaw(1e18)
	s.BuySomeTokens(planId, sample.Acc(), soldAmt)

	// settle should fail as no rollappDenom balance available
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().Error(err)

	// should succeed after fund (mocks the genesis bridge transfer)
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	// settle again should fail as already settled
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().Error(err)

	// assert no IRO balance in the account
	balance = s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().True(balance.IsZero())

	// assert sold amount is kept in the account and not used for liquidity pool
	balance = s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), rollappDenom)
	s.Require().Equal(soldAmt, balance.Amount)
}

func (s *KeeperTestSuite) TestBootstrapLiquidityPool() {
	curve := types.BondingCurve{
		M:                      math.LegacyMustNewDecFromStr("0"),
		N:                      math.LegacyMustNewDecFromStr("1"),
		C:                      math.LegacyMustNewDecFromStr("0.1"), // each token costs 0.1 DYM
		RollappDenomDecimals:   18,
		LiquidityDenomDecimals: 18,
	}

	startTime := time.Now()
	allocation := math.NewInt(1_000_000).MulRaw(1e18)
	rollappDenom := "dasdasdasdasdsa"
	liquidityPart := types.DefaultParams().MinLiquidityPart
	maxToSell := types.FindEquilibrium(curve, allocation, liquidityPart)

	testCases := []struct {
		name           string
		buyAmt         math.Int
		expectedDYM    math.Int
		expectedTokens math.Int
	}{
		// - the expected DYM in the pool is the buy amount * 0.1 (fixed price) + 0.1 DYM creation fee
		{
			name:           "Small purchase",
			buyAmt:         math.NewInt(999).MulRaw(1e18),
			expectedDYM:    math.NewInt(100).MulRaw(1e18),   // 100 DYM
			expectedTokens: math.NewInt(1_000).MulRaw(1e18), // 1000 tokens
		},
		{
			name:           "Nothing sold - pool contains only creation fee",
			buyAmt:         math.NewInt(0),
			expectedDYM:    math.NewInt(1).MulRaw(1e17), // 0.1 dym creation fee
			expectedTokens: math.NewInt(1).MulRaw(1e18), // 1 token
		},
		{
			name:           "Large purchase",
			buyAmt:         math.NewInt(399_999).MulRaw(1e18),
			expectedDYM:    math.NewInt(40_000).MulRaw(1e18),
			expectedTokens: math.NewInt(400_000).MulRaw(1e18),
		},
		{
			name:           "All available tokens",
			buyAmt:         maxToSell.SubRaw(1e18), // 500_000 - 1
			expectedDYM:    maxToSell.ToLegacyDec().Mul(curve.C).TruncateInt(),
			expectedTokens: allocation.Sub(maxToSell),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset the test state for each test case
			rollappId := s.CreateDefaultRollapp()
			rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
			k := s.App.IROKeeper

			// Create IRO plan
			planDenom := "adym"
			apptesting.FundAccount(s.App, s.Ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), sdk.NewCoins(sdk.NewCoin(planDenom, k.GetParams(s.Ctx).CreationFee)))
			planId, err := k.CreatePlan(s.Ctx, "adym", allocation, time.Hour, startTime, true, rollapp, curve, types.DefaultIncentivePlanParams(), liquidityPart, time.Hour, 0)
			s.Require().NoError(err)
			reservedTokens := k.MustGetPlan(s.Ctx, planId).SoldAmt

			// Buy tokens
			if tc.buyAmt.GT(math.ZeroInt()) {
				s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
				buyer := sample.Acc()
				s.BuySomeTokens(planId, buyer, tc.buyAmt)
			}

			plan := k.MustGetPlan(s.Ctx, planId)
			raisedDYM := k.BK.GetBalance(s.Ctx, plan.GetAddress(), plan.LiquidityDenom)
			s.Require().Equal(tc.expectedDYM.String(), raisedDYM.Amount.String())

			unallocatedTokensAmt := allocation.Sub(plan.SoldAmt).Add(reservedTokens)

			// Settle
			s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, allocation)))
			err = k.Settle(s.Ctx, rollappId, rollappDenom)
			s.Require().NoError(err)

			// Assert liquidity pool
			expectedDYMInPool := tc.expectedDYM.ToLegacyDec().Mul(liquidityPart).TruncateInt()
			expectedTokensInPool := expectedDYMInPool.ToLegacyDec().Quo(curve.C).TruncateInt()

			poolId := uint64(1)
			pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			s.Require().NoError(err)

			poolCoins := pool.GetTotalPoolLiquidity(s.Ctx)
			s.Require().Equal(expectedDYMInPool.String(), poolCoins.AmountOf("adym").String())
			s.Require().Equal(expectedTokensInPool, poolCoins.AmountOf(rollappDenom))

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
			// 		raisedDYM  - founder share - actual poolCoins
			// 		unallocatedTokens - actual poolCoins
			expectedIncentives := sdk.NewCoins(sdk.NewCoin("adym", expectedDYMInPool), sdk.NewCoin(rollappDenom, unallocatedTokensAmt)).Sub(poolCoins...)
			s.Assert().Equal(expectedIncentives.String(), gauge.Coins.String())
		})
	}
}

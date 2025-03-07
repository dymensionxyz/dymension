package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// TestClaim tests that Claim works correctly.
//
// It creates a rollapp, then buys some tokens on it. It then tests that Claim fails
// if the plan is not settled. After settling the plan, it tests that Claim works
// and that the user gets the correct amount of tokens.
func (s *KeeperTestSuite) TestClaim() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()
	rollappDenom := "dasdasdasdasdsa"
	liquidityPart := types.DefaultParams().MinLiquidityPart

	startTime := time.Now()
	amt := math.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, startTime.Add(time.Hour), rollapp, curve, incentives, liquidityPart)
	s.Require().NoError(err)
	planDenom := k.MustGetPlan(s.Ctx, planId).TotalAllocation.Denom
	balance := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().Equal(amt, balance.Amount)

	claimer := sample.Acc()
	// buy some tokens
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	soldAmt := math.NewInt(1_000).MulRaw(1e18)
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

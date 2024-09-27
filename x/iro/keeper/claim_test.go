package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func (s *KeeperTestSuite) TestClaim() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()
	rollappDenom := "dasdasdasdasdsa"

	startTime := time.Now()
	amt := sdk.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, startTime.Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)
	planDenom := k.MustGetPlan(s.Ctx, planId).TotalAllocation.Denom
	balance := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().Equal(amt, balance.Amount)

	claimer := sample.Acc()
	// buy some tokens
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	soldAmt := sdk.NewInt(1_000).MulRaw(1e18)
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

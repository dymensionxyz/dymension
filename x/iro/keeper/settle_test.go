package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func (s *KeeperTestSuite) TestSettle() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()

	startTime := time.Now()
	amt := sdk.NewInt(1_000_000)
	rollappDenom := "dasdasdasdasdsa"

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, startTime.Add(time.Hour), rollapp, curve)
	s.Require().NoError(err)
	planDenom := k.MustGetPlan(s.Ctx, planId).TotalAllocation.Denom
	// assert initial FUT balance
	balance := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().Equal(amt, balance.Amount)

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

	// assert unsold amt is claimed
	balance = s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().True(balance.IsZero())
}

// TestClaim after settle
func (s *KeeperTestSuite) TestClaim() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	rollappDenom := "dasdasdasdasdsa"

	startTime := time.Now()
	amt := sdk.NewInt(1_000_000)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, startTime.Add(time.Hour), rollapp, curve)
	s.Require().NoError(err)
	planDenom := k.MustGetPlan(s.Ctx, planId).TotalAllocation.Denom
	balance := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().Equal(amt, balance.Amount)

	// claim should fail as not settled
	claimer := sample.Acc()
	err = k.Claim(s.Ctx, planId, claimer.String())
	s.Require().Error(err)

	// settle
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	// claim should fail as no balance available
	err = k.Claim(s.Ctx, planId, claimer.String())
	s.Require().Error(err)

	// fund. claim should succeed
	s.FundAcc(claimer, sdk.NewCoins(sdk.NewCoin(planDenom, amt)))
	err = k.Claim(s.Ctx, planId, claimer.String())
	s.Require().NoError(err)

	// assert claimed amt
	balance = s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().True(balance.IsZero())
	balance = s.App.BankKeeper.GetBalance(s.Ctx, claimer, rollappDenom)
	s.Require().Equal(amt, balance.Amount)
}

// Test liquidity pool bootstrap

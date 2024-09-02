package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	balance := s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().Equal(amt, balance.Amount)

	//assert initial FUT balance

	// should fail as no balance available
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().Error(err)

	// should succeed after fund
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	// settle again should fail as already settled
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().Error(err)

	// assert unsold amt is claimed and used for liquidity pool
	balance = s.App.BankKeeper.GetBalance(s.Ctx, k.AK.GetModuleAddress(types.ModuleName), planDenom)
	s.Require().True(balance.IsZero())
}

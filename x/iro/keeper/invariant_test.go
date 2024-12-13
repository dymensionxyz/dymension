package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	keeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func (s *KeeperTestSuite) TestInvariantAccounting() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	endTime := startTime.Add(time.Hour)
	amt := sdk.NewInt(1_000_000).MulRaw(1e18)
	rollappDenom := "test_rollapp_denom"

	// Create a plan
	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, amt, startTime, endTime, rollapp, curve, incentives)
	s.Require().NoError(err)
	plan := k.MustGetPlan(s.Ctx, planId)
	planDenom := plan.TotalAllocation.Denom

	// Buy some tokens to create a non-zero sold amount
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	soldAmt := sdk.NewInt(1_000).MulRaw(1e18)
	buyer := sample.Acc()
	s.BuySomeTokens(planId, buyer, soldAmt)

	// Check invariant before settlement - should pass
	inv := keeper.InvariantAccounting(*k)
	err = inv(s.Ctx)
	s.Require().NoError(err)

	// Fund module with RA tokens and settle the plan
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	// Check invariant after settlement but before claims - should pass
	err = inv(s.Ctx)
	s.Require().NoError(err)

	// Claim some tokens
	err = k.Claim(s.Ctx, planId, buyer)
	s.Require().NoError(err)

	// Check invariant after claims - should pass
	err = inv(s.Ctx)
	s.Require().NoError(err)

	// Artificially break invariant by minting IRO tokens to module
	err = s.App.BankKeeper.MintCoins(s.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(planDenom, sdk.NewInt(1))))
	s.Require().NoError(err)

	// Check invariant - should be broken due to IRO tokens in module after settlement
	err = inv(s.Ctx)
	s.Require().Error(err)

	// Clean up minted tokens
	err = s.App.BankKeeper.BurnCoins(s.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(planDenom, sdk.NewInt(1))))
	s.Require().NoError(err)
}

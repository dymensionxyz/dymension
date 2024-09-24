package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/osmosis-labs/osmosis/v15/x/txfees"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// FIXME: test trade after settled

// FIXME: test taker fee (+low values should fail)

// FIXME: add sell test

func (s *KeeperTestSuite) TestBuy() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	maxAmt := sdk.NewInt(1_000_000_000).MulRaw(1e18)
	totalAllocation := sdk.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, totalAllocation, startTime, startTime.Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := sdk.NewInt(1_000).MulRaw(1e18)

	// buy before plan start - should fail
	err = k.Buy(s.Ctx.WithBlockTime(startTime.Add(-time.Minute)), planId, buyer, buyAmt, maxAmt)
	s.Require().Error(err)

	// cost is higher than maxCost specified - should fail
	expectedCost := curve.Cost(math.ZeroInt(), buyAmt)
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, expectedCost.SubRaw(1))
	s.Require().Error(err)

	// buy more than user's balance - should fail
	err = k.Buy(s.Ctx, planId, buyer, sdk.NewInt(100_000).MulRaw(1e18), maxAmt)
	s.Require().Error(err)

	// buy very small amount - should fail
	err = k.Buy(s.Ctx, planId, buyer, sdk.NewInt(100), maxAmt)
	s.Require().Error(err)

	// assert nothing sold
	plan, _ := k.GetPlan(s.Ctx, planId)
	s.Assert().Equal(sdk.NewInt(0), plan.SoldAmt)
	buyerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, buyer).AmountOf("adym")
	s.Assert().Equal(buyersFunds.AmountOf("adym"), buyerBalance)

	// successful buy
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, maxAmt)
	s.Require().NoError(err)
	plan, _ = k.GetPlan(s.Ctx, planId)
	s.Assert().True(plan.SoldAmt.Equal(buyAmt))

	// check cost again - should be higher
	expectedCost2 := curve.Cost(plan.SoldAmt, plan.SoldAmt.Add(buyAmt))
	s.Require().NoError(err)
	s.Assert().True(expectedCost2.GT(expectedCost))

	// assert balance
	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, buyer)
	takerFee := s.App.BankKeeper.GetAllBalances(s.Ctx, authtypes.NewModuleAddress(txfees.ModuleName))
	expectedBalance := buyersFunds.AmountOf("adym").Sub(expectedCost).Sub(takerFee.AmountOf("adym"))
	s.Require().Equal(expectedBalance, balances.AmountOf("adym"))

	expectedBaseDenom := fmt.Sprintf("%s_%s", types.IROTokenPrefix, rollappId)
	s.Require().Equal(buyAmt, balances.AmountOf(expectedBaseDenom))
}

func (s *KeeperTestSuite) TestBuyAllocationLimit() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	// setting curve so prices will be cheap
	curve := types.BondingCurve{
		M: math.LegacyMustNewDecFromStr("0.0005"),
		N: math.LegacyMustNewDecFromStr("0.1"),
		C: math.LegacyZeroDec(),
	}
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	maxAmt := sdk.NewInt(1_000_000_000).MulRaw(1e18)
	totalAllocation := sdk.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, totalAllocation, startTime, startTime.Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)

	buyer := sample.Acc()
	s.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("adym", maxAmt)))

	// plan start
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	// buy more than total allocation limit - should fail
	err = k.Buy(s.Ctx, planId, buyer, totalAllocation, maxAmt)
	s.Require().Error(err)

	// buy less than total allocation limit - should pass
	maxSellAmt := totalAllocation.ToLegacyDec().Mul(keeper.AllocationSellLimit).TruncateInt()
	err = k.Buy(s.Ctx, planId, buyer, maxSellAmt, maxAmt)
	s.Require().NoError(err)
}

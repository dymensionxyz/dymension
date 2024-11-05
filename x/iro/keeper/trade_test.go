package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

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
	initialOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId))

	plan := k.MustGetPlan(s.Ctx, planId)
	reservedTokens := plan.SoldAmt
	s.Assert().True(reservedTokens.GT(sdk.ZeroInt()))
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := sdk.NewInt(1_000).MulRaw(1e18)
	expectedCost := curve.Cost(plan.SoldAmt, plan.SoldAmt.Add(buyAmt))

	// buy before plan start - should fail
	err = k.Buy(s.Ctx.WithBlockTime(startTime.Add(-time.Minute)), planId, buyer, buyAmt, maxAmt)
	s.Require().Error(err)

	// cost is higher than maxCost specified - should fail
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, expectedCost.SubRaw(1))
	s.Require().Error(err)

	// buy more than user's balance - should fail
	err = k.Buy(s.Ctx, planId, buyer, sdk.NewInt(100_000).MulRaw(1e18), maxAmt)
	s.Require().Error(err)

	// buy very small amount - should fail (as cost ~= 0)
	err = k.Buy(s.Ctx, planId, buyer, sdk.NewInt(100), maxAmt)
	s.Require().Error(err)

	// assert nothing sold
	plan = k.MustGetPlan(s.Ctx, planId)
	s.Assert().Equal(reservedTokens, plan.SoldAmt) // nothing sold, still reserved amount
	buyerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, buyer).AmountOf("adym")
	s.Assert().Equal(buyersFunds.AmountOf("adym"), buyerBalance)

	// successful buy
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, maxAmt)
	s.Require().NoError(err)
	plan, _ = k.GetPlan(s.Ctx, planId)
	s.Assert().True(plan.SoldAmt.Sub(reservedTokens).Equal(buyAmt))

	// check cost again - should be higher
	expectedCost2 := curve.Cost(plan.SoldAmt, plan.SoldAmt.Add(buyAmt))
	s.Require().NoError(err)
	s.Assert().True(expectedCost2.GT(expectedCost))

	// extract taker fee from buy event
	takerFeeAmt := s.TakerFeeAmtAfterBuy()

	// assert balance
	buyerFinalBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, buyer)
	expectedBalance := buyersFunds.AmountOf("adym").Sub(expectedCost).Sub(takerFeeAmt)
	s.Require().Equal(expectedBalance, buyerFinalBalance.AmountOf("adym"))
	s.Require().Equal(buyAmt, buyerFinalBalance.AmountOf(plan.GetIRODenom()))

	// assert owner is incentivized: it must get 50% of taker fee
	currentOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId))
	ownerBalanceChange := currentOwnerBalance.Sub(initialOwnerBalance...)
	ownerRevenue := takerFeeAmt.QuoRaw(2)
	s.Require().Equal(ownerRevenue, ownerBalanceChange.AmountOf("adym"))
}

func (s *KeeperTestSuite) TestTradeAfterSettled() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	endTime := startTime.Add(time.Hour)
	maxAmt := sdk.NewInt(1_000_000_000).MulRaw(1e18)
	totalAllocation := sdk.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, totalAllocation, startTime, endTime, rollapp, curve, incentives)
	s.Require().NoError(err)

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := sdk.NewInt(1_000).MulRaw(1e18)

	// Buy before settlement
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, maxAmt)
	s.Require().NoError(err)

	// settle
	rollappDenom := "dasdasdasdasdsa"
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, totalAllocation)))
	err = k.Settle(s.Ctx, rollappId, rollappDenom)
	s.Require().NoError(err)

	// Attempt to buy after settlement - should fail
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, maxAmt)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestTakerFee() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	// Bonding curve with fixed price (1 token = 1 adym)
	curve := types.BondingCurve{
		M: math.LegacyMustNewDecFromStr("0"),
		N: math.LegacyMustNewDecFromStr("1"),
		C: math.LegacyMustNewDecFromStr("1"),
	}
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	totalAllocation := sdk.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, totalAllocation, startTime, startTime.Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := sdk.NewInt(1_000).MulRaw(1e18)

	// Attempt to buy while ignoring taker fee - should fail
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, buyAmt)
	s.Require().Error(err)

	// Successful buy
	expectedTakerFee := s.App.IROKeeper.GetParams(s.Ctx).TakerFee.MulInt(buyAmt).TruncateInt()
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, buyAmt.Add(expectedTakerFee))
	s.Require().NoError(err)

	// Extract taker fee from buy event
	takerFeeAmtBuy := s.TakerFeeAmtAfterBuy()

	// Check taker fee
	s.Require().Equal(expectedTakerFee, takerFeeAmtBuy)
}

func (s *KeeperTestSuite) TestSell() {
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
	initialOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId))
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := sdk.NewInt(1_000).MulRaw(1e18)

	// Buy tokens first
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, maxAmt)
	s.Require().NoError(err)

	// Extract taker fee from buy event
	takerFeeAmtBuy := s.TakerFeeAmtAfterBuy()

	// Sell tokens
	sellAmt := sdk.NewInt(500).MulRaw(1e18)
	minReceive := sdk.NewInt(1) // Set a very low minReceive for testing purposes
	err = k.Sell(s.Ctx, planId, buyer, sellAmt, minReceive)
	s.Require().NoError(err)

	// Extract taker fee from sell event
	takerFeeAmtSell := s.TakerFeeAmtAfterSell()

	// Check balances after sell
	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, buyer)
	s.Require().Equal(buyAmt.Sub(sellAmt), balances.AmountOf(k.MustGetPlan(s.Ctx, planId).GetIRODenom()))

	// Assert owner is incentivized: it must get 50% of taker fee
	currentOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId))
	ownerBalanceChange := currentOwnerBalance.Sub(initialOwnerBalance...)
	// ownerRevenue = (takerFeeBuy + takerFeeSell) / 2
	ownerRevenue := takerFeeAmtBuy.Add(takerFeeAmtSell).QuoRaw(2)
	s.Require().Equal(ownerRevenue, ownerBalanceChange.AmountOf("adym"))

	// Attempt to sell more than owned - should fail
	err = k.Sell(s.Ctx, planId, buyer, buyAmt, minReceive)
	s.Require().Error(err)

	// Attempt to sell with minReceive higher than possible - should fail
	highMinReceive := maxAmt
	err = k.Sell(s.Ctx, planId, buyer, sellAmt, highMinReceive)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TakerFeeAmtAfterSell() sdk.Int {
	// Extract taker fee from event
	eventName := proto.MessageName(new(types.EventSell))
	takerFeeAmt, found := s.ExtractTakerFeeAmtFromEvents(s.Ctx.EventManager().Events(), eventName)
	s.Require().True(found)
	return takerFeeAmt
}

func (s *KeeperTestSuite) TakerFeeAmtAfterBuy() sdk.Int {
	// Extract taker fee from event
	eventName := proto.MessageName(new(types.EventBuy))
	takerFeeAmt, found := s.ExtractTakerFeeAmtFromEvents(s.Ctx.EventManager().Events(), eventName)
	s.Require().True(found)
	return takerFeeAmt
}

func (s *KeeperTestSuite) ExtractTakerFeeAmtFromEvents(events []sdk.Event, eventName string) (sdk.Int, bool) {
	event, found := s.FindLastEventOfType(events, eventName)
	if !found {
		return sdk.Int{}, false
	}
	attrs := s.ExtractAttributes(event)
	for key, value := range attrs {
		if key == "taker_fee" {
			fee, ok := sdk.NewIntFromString(value)
			s.Require().True(ok)
			return fee, true
		}
	}
	return sdk.ZeroInt(), false
}

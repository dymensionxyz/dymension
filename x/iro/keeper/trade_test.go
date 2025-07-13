package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func (s *KeeperTestSuite) TestTradeDisabled() {
	rollappId := s.CreateDefaultRollapp()
	owner := s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId)

	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	maxAmt := math.NewInt(1_000_000_000).MulRaw(1e18)
	totalAllocation := math.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, "adym", totalAllocation, time.Hour, startTime, false, rollapp, curve, incentives, types.DefaultParams().MinLiquidityPart, time.Hour, 0)
	s.Require().NoError(err)

	plan := k.MustGetPlan(s.Ctx, planId)
	s.Assert().False(plan.TradingEnabled)
	s.Assert().True(plan.StartTime.IsZero())
	s.Assert().True(plan.PreLaunchTime.IsZero())

	// Verify rollapp is not launchable (pre-launch time is far in the future)
	rollapp = s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	s.Require().NotNil(rollapp.PreLaunchTime)
	s.Assert().True(rollapp.PreLaunchTime.After(s.Ctx.BlockTime().Add(time.Hour * 24 * 365 * 9))) // at least 9 years in future

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)
	s.FundAcc(owner, buyersFunds)

	buyAmt := math.NewInt(1_000).MulRaw(1e18)

	// buy before plan start - should fail
	err = k.Buy(s.Ctx.WithBlockTime(startTime.Add(-time.Minute)), planId, buyer, buyAmt, maxAmt)
	s.Require().Error(err)

	// Plan is not yet enabled - should fail
	err = k.Buy(s.Ctx.WithBlockTime(startTime.Add(time.Minute)), planId, buyer, buyAmt, maxAmt)
	s.Require().Error(err)

	// owner can still buy
	err = k.Buy(s.Ctx.WithBlockTime(startTime.Add(-time.Minute)), planId, owner, buyAmt, maxAmt)
	s.Require().NoError(err)

	// Enable trading not as owner - should fail
	err = s.App.IROKeeper.EnableTrading(s.Ctx, planId, buyer)
	s.Require().Error(err)

	// Enable trading by owner
	enableTime := time.Now().Round(0).UTC()
	s.Ctx = s.Ctx.WithBlockTime(enableTime)
	err = s.App.IROKeeper.EnableTrading(s.Ctx, planId, owner)
	s.Require().NoError(err)

	// Verify plan trading is enabled and times are set correctly
	plan = k.MustGetPlan(s.Ctx, planId)
	s.Assert().True(plan.TradingEnabled)
	s.Assert().Equal(enableTime, plan.StartTime)
	s.Assert().Equal(enableTime.Add(plan.IroPlanDuration), plan.PreLaunchTime)

	// Verify rollapp pre-launch time is updated
	rollapp = s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	s.Require().NotNil(rollapp.PreLaunchTime)
	s.Assert().Equal(plan.PreLaunchTime, *rollapp.PreLaunchTime)

	// Buy should now succeed
	err = k.Buy(s.Ctx.WithBlockTime(enableTime.Add(2*time.Minute)), planId, buyer, buyAmt, maxAmt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestBuy() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	maxAmt := math.NewInt(1_000_000_000).MulRaw(1e18)
	totalAllocation := math.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, "adym", totalAllocation, time.Hour, startTime, true, rollapp, curve, incentives, types.DefaultParams().MinLiquidityPart, time.Hour, 0)
	s.Require().NoError(err)
	initialOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId))

	plan := k.MustGetPlan(s.Ctx, planId)
	reservedTokens := plan.SoldAmt
	s.Assert().True(reservedTokens.GT(math.ZeroInt()))
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := math.NewInt(1_000).MulRaw(1e18)
	expectedCost := curve.Cost(plan.SoldAmt, plan.SoldAmt.Add(buyAmt))

	// buy before plan start - should fail
	err = k.Buy(s.Ctx.WithBlockTime(startTime.Add(-time.Minute)), planId, buyer, buyAmt, maxAmt)
	s.Require().Error(err)

	// cost is higher than maxCost specified - should fail
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, expectedCost.SubRaw(1))
	s.Require().Error(err)

	// buy more than user's balance - should fail
	err = k.Buy(s.Ctx, planId, buyer, math.NewInt(100_000).MulRaw(1e18), maxAmt)
	s.Require().Error(err)

	// buy very small amount - should fail (as cost ~= 0)
	err = k.Buy(s.Ctx, planId, buyer, math.NewInt(100), maxAmt)
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
	maxAmt := math.NewInt(1_000_000_000).MulRaw(1e18)
	totalAllocation := math.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, "adym", totalAllocation, time.Hour, startTime, true, rollapp, curve, incentives, types.DefaultParams().MinLiquidityPart, time.Hour, 0)
	s.Require().NoError(err)

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := math.NewInt(1_000).MulRaw(1e18)

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
		M:                      math.LegacyMustNewDecFromStr("0"),
		N:                      math.LegacyMustNewDecFromStr("1"),
		C:                      math.LegacyMustNewDecFromStr("1"),
		RollappDenomDecimals:   18,
		LiquidityDenomDecimals: 18,
	}
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	totalAllocation := math.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, "adym", totalAllocation, time.Hour, startTime, true, rollapp, curve, incentives, types.DefaultParams().MinLiquidityPart, time.Hour, 0)
	s.Require().NoError(err)
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := math.NewInt(1_000).MulRaw(1e18)

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
	maxAmt := math.NewInt(1_000_000_000).MulRaw(1e18)
	totalAllocation := math.NewInt(1_000_000).MulRaw(1e18)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, "adym", totalAllocation, time.Hour, startTime, true, rollapp, curve, incentives, types.DefaultParams().MinLiquidityPart, time.Hour, 0)
	s.Require().NoError(err)
	initialOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId))
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18)))
	s.FundAcc(buyer, buyersFunds)

	buyAmt := math.NewInt(1_000).MulRaw(1e18)

	// Buy tokens first
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, maxAmt)
	s.Require().NoError(err)

	// Extract taker fee from buy event
	takerFeeAmtBuy := s.TakerFeeAmtAfterBuy()

	// Sell tokens
	sellAmt := math.NewInt(500).MulRaw(1e18)
	minReceive := math.NewInt(1) // Set a very low minReceive for testing purposes
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

func (s *KeeperTestSuite) TakerFeeAmtAfterSell() math.Int {
	// Extract taker fee from event
	eventName := proto.MessageName(new(types.EventSell))
	takerFeeAmt, found := s.ExtractTakerFeeAmtFromEvents(s.Ctx.EventManager().Events(), eventName)
	s.Require().True(found)
	return takerFeeAmt
}

func (s *KeeperTestSuite) TakerFeeAmtAfterBuy() math.Int {
	// Extract taker fee from event
	eventName := proto.MessageName(new(types.EventBuy))
	takerFeeAmt, found := s.ExtractTakerFeeAmtFromEvents(s.Ctx.EventManager().Events(), eventName)
	s.Require().True(found)
	return takerFeeAmt
}

func (s *KeeperTestSuite) ExtractTakerFeeAmtFromEvents(events []sdk.Event, eventName string) (math.Int, bool) {
	event, found := s.FindLastEventOfType(events, eventName)
	if !found {
		return math.Int{}, false
	}
	// Look for taker_fee attribute (structured coin object)
	for _, attr := range event.Attributes {
		if attr.GetKey() == "taker_fee" {
			var coin sdk.Coin
			err := s.App.AppCodec().UnmarshalJSON([]byte(attr.GetValue()), &coin)
			s.Require().NoError(err)
			return coin.Amount, true
		}
	}
	return math.ZeroInt(), false
}

func (s *KeeperTestSuite) TestBuyWithUSDC() {
	// Note: USDC has 6 decimals instead of 18
	s.PreparePoolWithCoins(sdk.NewCoins(
		sdk.NewCoin("usdc", math.NewInt(1_000_000).MulRaw(1e6)),
		sdk.NewCoin("adym", math.NewInt(1_000_000).MulRaw(1e18)),
	))

	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	// Bonding curve with fixed price (1 token = 1 usdc)
	curve := types.BondingCurve{
		M:                      math.LegacyMustNewDecFromStr("0"),
		N:                      math.LegacyMustNewDecFromStr("1"),
		C:                      math.LegacyMustNewDecFromStr("1"),
		RollappDenomDecimals:   18,
		LiquidityDenomDecimals: 6, // USDC has 6 decimals
	}
	incentives := types.DefaultIncentivePlanParams()

	startTime := time.Now()
	totalAllocation := math.NewInt(1_000_000).MulRaw(1e18) // 1M tokens with 18 decimals

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	owner := rollapp.Owner

	// Create plan with USDC as liquidity denom instead of DYM
	// Fund owner with USDC (6 decimals) for creation fee
	s.FundAcc(sdk.MustAccAddressFromBech32(owner), sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(100_000).MulRaw(1e6)))) // 100K USDC)
	planId, err := k.CreatePlan(s.Ctx, "usdc", totalAllocation, time.Hour, startTime, true, rollapp, curve, incentives, types.DefaultParams().MinLiquidityPart, time.Hour, 0)
	s.Require().NoError(err)

	initialOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId))
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Minute))

	// Fund buyer with USDC (6 decimals)
	buyer := sample.Acc()
	buyersFunds := sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(100_000).MulRaw(1e6))) // 100K USDC
	s.FundAcc(buyer, buyersFunds)

	// Buy 1000 tokens (with 18 decimals)
	buyAmt := math.NewInt(1_000).MulRaw(1e18)
	// Expected cost is 1000 USDC (with 6 decimals)
	expectedCost := math.NewInt(1_000).MulRaw(1e6)
	expectedTakerFee := s.App.IROKeeper.GetParams(s.Ctx).TakerFee.MulInt(expectedCost).TruncateInt()
	maxCost := expectedCost.Add(expectedTakerFee)

	// Successful buy
	err = k.Buy(s.Ctx, planId, buyer, buyAmt, maxCost)
	s.Require().NoError(err)

	// Extract taker fee from buy event
	takerFeeAmt := s.TakerFeeAmtAfterBuy()

	// Check taker fee
	s.Require().Equal(expectedTakerFee, takerFeeAmt)

	// Check buyer's balance after purchase
	buyerFinalBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, buyer)
	expectedUSDCBalance := buyersFunds.AmountOf("usdc").Sub(expectedCost).Sub(takerFeeAmt)
	s.Require().Equal(expectedUSDCBalance, buyerFinalBalance.AmountOf("usdc"))
	s.Require().Equal(buyAmt, buyerFinalBalance.AmountOf(k.MustGetPlan(s.Ctx, planId).GetIRODenom()))

	// Assert owner is incentivized: it must get 50% of taker fee
	currentOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.RollappKeeper.MustGetRollappOwner(s.Ctx, rollappId))
	ownerBalanceChange := currentOwnerBalance.Sub(initialOwnerBalance...)

	// taker fee is swapped to dym when charged
	ownerRoyalties := takerFeeAmt.QuoRaw(2)
	royaltiesInDym := ownerRoyalties.MulRaw(1e12)                                                // from usdc -> dym (assuming 1:1 price)
	err = approxEqualInt(royaltiesInDym, ownerBalanceChange.AmountOf("adym"), math.NewInt(1e16)) // 0.01 DYM tolre
	s.Require().NoError(err)
}

// approxEqualInt checks if two values of different types are approximately equal
func approxEqualInt(expected, actual, tolerance math.Int) error {
	diff := expected.Sub(actual).Abs()
	if diff.GTE(tolerance) {
		return fmt.Errorf("expected %s, got %s, diff %s", expected, actual, diff)
	}

	return nil
}

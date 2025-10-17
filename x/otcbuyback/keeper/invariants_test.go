package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/keeper"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

var defaultLinearDiscount = types.NewLinearDiscountType(
	math.LegacyNewDecWithPrec(2, 1), // 0.2 = 20% initial discount
	math.LegacyNewDecWithPrec(5, 1), // 0.5 = 50% max discount
	24*time.Hour,
)

func (suite *KeeperTestSuite) TestModuleAccountBalanceInvariant_ActiveAuction() {
	// Create and complete an auction
	allocation := sdk.NewCoin("adym", math.NewInt(10).MulRaw(1e18))
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(allocation))

	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation,
		suite.Ctx.BlockTime().Add(-1*time.Hour), // started 1 hour ago
		suite.Ctx.BlockTime().Add(1*time.Hour),  // ends in 1 hour
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	// Create some purchases
	buyer1 := suite.CreateRandomAccount()
	buyer2 := suite.CreateRandomAccount()

	// Purchase 3 tokens by buyer1
	purchase1 := newTestPurchase(math.NewInt(3).MulRaw(1e18))
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer1, purchase1)
	suite.Require().NoError(err)

	// Purchase 2 tokens by buyer2, claim 0.5
	purchase2 := newTestPurchase(math.NewInt(2).MulRaw(1e18))
	purchase2.ClaimTokens(math.NewInt(1).MulRaw(1e17)) // claim 0.5 tokens
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer2, purchase2)
	suite.Require().NoError(err)

	// Update auction sold amount
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found)
	auction.SoldAmount = math.NewInt(5).MulRaw(1e18) // 3 + 2
	raisedAmount := sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(1000).MulRaw(1e6)))
	auction.RaisedAmount = raisedAmount
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction)
	suite.Require().NoError(err)

	suite.FundModuleAcc(types.ModuleName, raisedAmount)

	// Test invariant - should pass
	invariant := keeper.ModuleAccountBalanceInvariant(*suite.App.OTCBuybackKeeper)
	msg, broken := invariant(suite.Ctx)
	suite.Require().False(broken, "invariant should not be broken: %s", msg)

	// The invariant should pass, which means the module account has sufficient balance
	// to cover all outstanding obligations (remaining allocation + unclaimed + raised amount)
}

func (suite *KeeperTestSuite) TestModuleAccountBalanceInvariant_CompletedAuction() {
	// Create and complete an auction
	allocation := sdk.NewCoin("adym", math.NewInt(10).MulRaw(1e18))
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(allocation))

	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation,
		suite.Ctx.BlockTime().Add(-2*time.Hour), // started 2 hours ago
		suite.Ctx.BlockTime().Add(-1*time.Hour), // ended 1 hour ago
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	// Create purchases
	buyer1 := suite.CreateRandomAccount()
	buyer2 := suite.CreateRandomAccount()

	// Purchase 4 tokens by buyer1, claim 1
	purchase1 := newTestPurchase(math.NewInt(4).MulRaw(1e18))
	purchase1.ClaimTokens(math.NewInt(1).MulRaw(1e18)) // claim 1 token
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer1, purchase1)
	suite.Require().NoError(err)

	// Purchase 3 tokens by buyer2, claim 2
	purchase2 := newTestPurchase(math.NewInt(3).MulRaw(1e18))
	purchase2.ClaimTokens(math.NewInt(2).MulRaw(1e18)) // claim 2 tokens
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer2, purchase2)
	suite.Require().NoError(err)

	// Update auction and mark as completed
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found)
	auction.SoldAmount = math.NewInt(7).MulRaw(1e18) // 4 + 3
	auction.Completed = true
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction)
	suite.Require().NoError(err)

	// Test invariant - should pass
	invariant := keeper.ModuleAccountBalanceInvariant(*suite.App.OTCBuybackKeeper)
	msg, broken := invariant(suite.Ctx)
	suite.Require().False(broken, "invariant should not be broken: %s", msg)

	// The invariant should pass, which means the module account has sufficient balance
	// to cover all outstanding obligations (only unclaimed tokens for completed auctions)
}

func (suite *KeeperTestSuite) TestModuleAccountBalanceInvariant_NegativeClaimed() {
	// Fund module account
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100).MulRaw(1e18))))

	// Create an auction
	allocation := sdk.NewCoin("adym", math.NewInt(10).MulRaw(1e18))
	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation,
		suite.Ctx.BlockTime().Add(-1*time.Hour),
		suite.Ctx.BlockTime().Add(1*time.Hour),
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	// Create a purchase with negative claimed amount (invalid state)
	buyer := suite.CreateRandomAccount()
	purchase := newTestPurchase(math.NewInt(1).MulRaw(1e18))
	purchase.Claimed = math.NewInt(-1).MulRaw(1e17) // negative claimed amount (-0.1)
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer, purchase)
	suite.Require().NoError(err)

	// Update auction sold amount
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found)
	auction.SoldAmount = math.NewInt(1).MulRaw(1e18)
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction)
	suite.Require().NoError(err)

	// Test invariant - should fail
	invariant := keeper.ModuleAccountBalanceInvariant(*suite.App.OTCBuybackKeeper)
	msg, broken := invariant(suite.Ctx)
	suite.Require().True(broken, "invariant should be broken")
	suite.Require().Contains(msg, "negative claimed amount")
}

func (suite *KeeperTestSuite) TestModuleAccountBalanceInvariant_ClaimedExceedsAmount() {
	// Fund module account
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100).MulRaw(1e18))))

	// Create an auction
	allocation := sdk.NewCoin("adym", math.NewInt(10).MulRaw(1e18))
	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation,
		suite.Ctx.BlockTime().Add(-1*time.Hour),
		suite.Ctx.BlockTime().Add(1*time.Hour),
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	// Create a purchase with claimed amount exceeding purchased amount
	buyer := suite.CreateRandomAccount()
	purchase := newTestPurchase(math.NewInt(1).MulRaw(1e18))
	purchase.Claimed = math.NewInt(2).MulRaw(1e18) // claimed > amount (2 > 1)
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer, purchase)
	suite.Require().NoError(err)

	// Update auction sold amount
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found)
	auction.SoldAmount = math.NewInt(1).MulRaw(1e18)
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction)
	suite.Require().NoError(err)

	// Test invariant - should fail
	invariant := keeper.ModuleAccountBalanceInvariant(*suite.App.OTCBuybackKeeper)
	msg, broken := invariant(suite.Ctx)
	suite.Require().True(broken, "invariant should be broken")
	suite.Require().Contains(msg, "claimed 2000000000000000000 exceeds purchased 1000000000000000000")
}

func (suite *KeeperTestSuite) TestModuleAccountBalanceInvariant_TotalPurchasedNotEqualSoldAmount() {
	// Fund module account
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100).MulRaw(1e18))))

	// Create an auction
	allocation := sdk.NewCoin("adym", math.NewInt(10).MulRaw(1e18))
	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation,
		suite.Ctx.BlockTime().Add(-1*time.Hour),
		suite.Ctx.BlockTime().Add(1*time.Hour),
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	// Create purchases totaling 2 tokens
	buyer1 := suite.CreateRandomAccount()
	buyer2 := suite.CreateRandomAccount()

	purchase1 := newTestPurchase(math.NewInt(1).MulRaw(1e18))
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer1, purchase1)
	suite.Require().NoError(err)

	purchase2 := newTestPurchase(math.NewInt(1).MulRaw(1e18))
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer2, purchase2)
	suite.Require().NoError(err)

	// Set auction sold amount to 3 (inconsistent with purchases)
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found)
	auction.SoldAmount = math.NewInt(3).MulRaw(1e18) // should be 2
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction)
	suite.Require().NoError(err)

	// Test invariant - should fail
	invariant := keeper.ModuleAccountBalanceInvariant(*suite.App.OTCBuybackKeeper)
	msg, broken := invariant(suite.Ctx)
	suite.Require().True(broken, "invariant should be broken")
	suite.Require().Contains(msg, "total purchased 2000000000000000000 != sold amount 3000000000000000000")
}

func (suite *KeeperTestSuite) TestModuleAccountBalanceInvariant_SoldAmountExceedsAllocation() {
	// Fund module account
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100).MulRaw(1e18))))

	// Create an auction
	allocation := sdk.NewCoin("adym", math.NewInt(10).MulRaw(1e18))
	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation,
		suite.Ctx.BlockTime().Add(-1*time.Hour),
		suite.Ctx.BlockTime().Add(1*time.Hour),
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	// Set auction sold amount to exceed allocation
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found)
	auction.SoldAmount = math.NewInt(15).MulRaw(1e18) // exceeds allocation of 10
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction)
	suite.Require().NoError(err)

	// Test invariant - should fail
	invariant := keeper.ModuleAccountBalanceInvariant(*suite.App.OTCBuybackKeeper)
	msg, broken := invariant(suite.Ctx)
	suite.Require().True(broken, "invariant should be broken")
	suite.Require().Contains(msg, "sold amount 15000000000000000000 > allocation 10000000000000000000")
}

func (suite *KeeperTestSuite) TestModuleAccountBalanceInvariant_InsufficientModuleBalance() {
	// Fund module account
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100).MulRaw(1e18))))

	// Create an auction
	allocation := sdk.NewCoin("adym", math.NewInt(10).MulRaw(1e18))
	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation,
		suite.Ctx.BlockTime().Add(-1*time.Hour),
		suite.Ctx.BlockTime().Add(1*time.Hour),
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	// Create purchases
	buyer := suite.CreateRandomAccount()
	purchase := newTestPurchase(math.NewInt(5).MulRaw(1e18))
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID, buyer, purchase)
	suite.Require().NoError(err)

	// Update auction
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found)
	auction.SoldAmount = math.NewInt(5).MulRaw(1e18)
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction)
	suite.Require().NoError(err)

	// Manually reduce module account balance to simulate insufficient funds
	err = suite.App.BankKeeper.SendCoinsFromModuleToAccount(
		suite.Ctx,
		types.ModuleName,
		suite.CreateRandomAccount(), // send to random account
		sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(95).MulRaw(1e18))), // send most of the tokens to make invariant fail
	)
	suite.Require().NoError(err)

	// Test invariant - should fail
	invariant := keeper.ModuleAccountBalanceInvariant(*suite.App.OTCBuybackKeeper)
	msg, broken := invariant(suite.Ctx)
	suite.Require().True(broken, "invariant should be broken")
	suite.Require().Contains(msg, "insufficient module balance")
}

func (suite *KeeperTestSuite) TestModuleAccountBalanceInvariant_MultipleAuctions() {
	// Create two active auctions
	allocation1 := sdk.NewCoin("adym", math.NewInt(10).MulRaw(1e18))
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(allocation1))
	auctionID1, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation1,
		suite.Ctx.BlockTime().Add(-1*time.Hour),
		suite.Ctx.BlockTime().Add(1*time.Hour),
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	allocation2 := sdk.NewCoin("adym", math.NewInt(5).MulRaw(1e18))
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(allocation2))
	auctionID2, err := suite.App.OTCBuybackKeeper.CreateAuction(
		suite.Ctx,
		allocation2,
		suite.Ctx.BlockTime().Add(-1*time.Hour),
		suite.Ctx.BlockTime().Add(1*time.Hour),
		defaultLinearDiscount,
		types.Auction_VestingParams{
			VestingDelay: time.Hour,
		},
		types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	)
	suite.Require().NoError(err)

	// Create purchases for both auctions
	buyer1 := suite.CreateRandomAccount()
	buyer2 := suite.CreateRandomAccount()

	// Auction 1: 3 tokens purchased, 1 claimed
	purchase1 := newTestPurchase(math.NewInt(3).MulRaw(1e18))
	purchase1.ClaimTokens(math.NewInt(1).MulRaw(1e18))
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID1, buyer1, purchase1)
	suite.Require().NoError(err)

	// Auction 2: 2 tokens purchased, 0.5 claimed
	purchase2 := newTestPurchase(math.NewInt(2).MulRaw(1e18))
	purchase2.ClaimTokens(math.NewInt(1).MulRaw(1e17)) // 0.5 tokens
	err = suite.App.OTCBuybackKeeper.SetPurchase(suite.Ctx, auctionID2, buyer2, purchase2)
	suite.Require().NoError(err)

	// Update auctions
	auction1, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID1)
	suite.Require().True(found)
	auction1.SoldAmount = math.NewInt(3).MulRaw(1e18)
	raisedAmount1 := sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(600).MulRaw(1e6)))
	auction1.RaisedAmount = raisedAmount1
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction1)
	suite.Require().NoError(err)

	auction2, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID2)
	suite.Require().True(found)
	auction2.SoldAmount = math.NewInt(2).MulRaw(1e18)
	raisedAmount2 := sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(400).MulRaw(1e6)))
	auction2.RaisedAmount = raisedAmount2
	err = suite.App.OTCBuybackKeeper.SetAuction(suite.Ctx, auction2)
	suite.Require().NoError(err)

	// Test invariant - should pass
	suite.FundModuleAcc(types.ModuleName, raisedAmount1.Add(raisedAmount2...))
	invariant := keeper.ModuleAccountBalanceInvariant(*suite.App.OTCBuybackKeeper)
	msg, broken := invariant(suite.Ctx)
	suite.Require().False(broken, "invariant should not be broken: %s", msg)

	// The invariant should pass, which means the module account has sufficient balance
	// to cover all outstanding obligations (remaining allocation + unclaimed + raised amount)
}

func newTestPurchase(amt math.Int) types.Purchase {
	return types.NewPurchase(types.NewPurchaseEntry(amt, time.Time{}, 0))
}

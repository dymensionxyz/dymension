package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// TestAuctionLifecycle tests the complete lifecycle of an OTC buyback auction.
// It verifies:
// 1. Auction state transitions (upcoming -> active -> completed)
// 2. Remaining unsold tokens are returned to the streamer module
// 3. Pump streams are created with correct parameters for raised funds
func (suite *KeeperTestSuite) TestAuctionLifecycle() {
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	// Create auction and verify initial state
	auctionID := suite.CreateDefaultAuction()
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found, "auction should be found")
	suite.Require().False(auction.IsCompleted(), "auction should not be completed")

	// Test auction state transitions over time

	// Time before auction starts - should be upcoming
	before := suite.Ctx.BlockTime().Add(-1 * time.Second)
	suite.Require().False(auction.IsActive(before), "auction should not be active before start time")
	suite.Require().True(auction.IsUpcoming(before), "auction should be upcoming before start time")

	// Time after auction starts - should be active
	after := suite.Ctx.BlockTime().Add(1 * time.Second)
	suite.Require().True(auction.IsActive(after), "auction should be active after start time")
	suite.Require().False(auction.IsUpcoming(after), "auction should not be upcoming after start time")

	// Record initial balances before auction completion
	initialStreamerBalance := suite.App.BankKeeper.GetBalance(suite.Ctx,
		authtypes.NewModuleAddress(streamertypes.ModuleName), "adym")
	initialOTCBalance := suite.App.BankKeeper.GetBalance(suite.Ctx,
		authtypes.NewModuleAddress(types.ModuleName), "adym")

	// Get initial auction allocation to calculate expected remaining funds
	originalAllocation := auction.Allocation

	// Advance time to end the auction without any purchases (all tokens remain)
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(24 * time.Hour).Add(1 * time.Second))
	err := suite.App.OTCBuybackKeeper.BeginBlock(suite.Ctx)
	suite.Require().NoError(err, "begin block should not error")

	// Verify auction is completed
	auction, found = suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found, "auction should be found after completion")
	suite.Require().True(auction.IsCompleted(), "auction should be completed after end time")
	suite.Require().False(auction.IsActive(suite.Ctx.BlockTime()), "auction should not be active after completion")
	suite.Require().False(auction.IsUpcoming(suite.Ctx.BlockTime()), "auction should not be upcoming after completion")

	// Test 1: Verify remaining funds are returned to streamer module
	// Since no purchases were made, all original allocation should be returned
	finalStreamerBalance := suite.App.BankKeeper.GetBalance(suite.Ctx,
		authtypes.NewModuleAddress(streamertypes.ModuleName), "adym")
	finalOTCBalance := suite.App.BankKeeper.GetBalance(suite.Ctx,
		authtypes.NewModuleAddress(types.ModuleName), "adym")

	// The streamer module should have received back the original allocation
	expectedStreamerIncrease := originalAllocation
	actualStreamerIncrease := finalStreamerBalance.Amount.Sub(initialStreamerBalance.Amount)
	suite.Require().Equal(expectedStreamerIncrease, actualStreamerIncrease,
		"streamer module should receive back all unsold tokens (%s), got %s",
		expectedStreamerIncrease, actualStreamerIncrease)

	// OTC module should have no remaining tokens from this auction
	suite.Require().True(finalOTCBalance.Amount.LTE(initialOTCBalance.Amount),
		"OTC module balance should not increase after returning unsold tokens")

	// For this test (no purchases), verify that the completion was handled properly
	// and auction state reflects no raised amount and no pump streams
	suite.Require().True(auction.RaisedAmount.IsZero(),
		"auction with no purchases should have empty raised amount")
	suite.Require().Equal(math.ZeroInt(), auction.SoldAmount,
		"auction with no purchases should have zero sold amount")

	// assert no pump streams were created
	streams := suite.App.StreamerKeeper.GetStreams(suite.Ctx)
	suite.Require().Equal(0, len(streams))
}

// TestAuctionLifecycleWithPurchases tests the complete auction lifecycle including purchases
// to verify pump stream creation with raised funds.
func (suite *KeeperTestSuite) TestPumpStreamsCreation() {
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	// Create auction and make some purchases
	auctionID := suite.CreateDefaultAuction()

	// Create a buyer and make a purchase
	buyer := suite.CreateRandomAccount()
	buyerFunds := sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(1000).MulRaw(1e6)))
	suite.FundAcc(buyer, buyerFunds)

	// Make a purchase (buy 50 DYM tokens)
	amountToBuy := math.NewInt(50).MulRaw(1e18)
	paymentCoin, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc")
	suite.Require().NoError(err)

	auction, _ := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	raisedAmount := auction.RaisedAmount.AmountOf("usdc")
	suite.Require().Equal(raisedAmount, paymentCoin.Amount)

	// Advance time to complete the auction
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(24 * time.Hour).Add(1 * time.Second))
	err = suite.App.OTCBuybackKeeper.BeginBlock(suite.Ctx)
	suite.Require().NoError(err)

	// Verify auction completion
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found)
	suite.Require().True(auction.IsCompleted())

	// Test 2: Verify pump streams created for raised funds
	// The raised USDC should be transferred back to streamer for pump distribution
	finalStreamerUSDCBalance := suite.App.BankKeeper.GetBalance(suite.Ctx,
		authtypes.NewModuleAddress(streamertypes.ModuleName), "usdc")

	suite.Require().Equal(raisedAmount, finalStreamerUSDCBalance.Amount,
		"streamer should receive raised amount for pump streams: expected %s USDC, got %s",
		raisedAmount, finalStreamerUSDCBalance)

	// assert pump streams were created
	streams := suite.App.StreamerKeeper.GetStreams(suite.Ctx)
	suite.Require().Equal(1, len(streams))
	stream := streams[0]
	suite.Require().True(stream.IsPumpStream())
	suite.Require().Equal(sdk.NewCoins(sdk.NewCoin("usdc", raisedAmount)), stream.Coins)

	streamPumpParams := *stream.PumpParams
	expectedCoins := sdk.NewCoins(sdk.NewCoin("usdc", raisedAmount)).QuoInt(math.NewIntFromUint64(auction.PumpParams.NumEpochsPaidOver))
	suite.Require().Equal(expectedCoins, streamPumpParams.EpochCoinsLeft)
	suite.Require().Equal(auction.PumpParams.NumOfPumpsPerEpoch, streamPumpParams.NumPumps)
	suite.Require().Equal(streamertypes.PumpDistr_PUMP_DISTR_UNIFORM, streamPumpParams.PumpDistr)

	pumpPool := streamPumpParams.GetPool()
	suite.Require().Equal(uint64(1), pumpPool.PoolId)
	suite.Require().Equal("adym", pumpPool.TokenOut)
}

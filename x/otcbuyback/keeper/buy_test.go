package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
	testutil "github.com/dymensionxyz/dymension/v3/testutil/math"
)

// FIXME: BuyExactSpend

func (suite *KeeperTestSuite) TestBuyPriceDiscount() {
	suite.Run("simple buy and claim flow without dynamic pricing", func() {
		suite.SetupTest()

		// Create a buyer account with USDC balance
		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(10_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		// Create an auction
		auctionID := suite.CreateDefaultAuction()
		auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)

		discount := auction.GetCurrentDiscount(suite.Ctx.BlockTime())
		suite.Require().Equal(discount, auction.InitialDiscount) // 10% default

		amountToBuy := math.NewInt(100).MulRaw(1e18)
		paymentCoin, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc")
		suite.Require().NoError(err)

		// should have spent ~22.5 usdc (pool price is 1:4 (25 usdc for 100 dym) + 10% discount)
		err = testutil.ApproxEqualRatio(math.NewInt(225).MulRaw(1e5), paymentCoin.Amount, 0.01)
		suite.Require().NoError(err, "Should have spent 22.5 usdc")
		suite.Require().Equal("usdc", paymentCoin.Denom)
	})
}

func (suite *KeeperTestSuite) TestMultipleBuyersAndClaims() {
	suite.Run("multiple buyers and claims", func() {
		suite.SetupTest()

		usdcDenom := "usdc"

		auctionID := suite.CreateDefaultAuction()

		// Create multiple buyers
		buyers := []sdk.AccAddress{suite.CreateRandomAccount(), suite.CreateRandomAccount(), suite.CreateRandomAccount()}
		amountsToBuy := []math.Int{
			math.NewInt(10).MulRaw(1e18), // Buyer 1: 10 tokens
			math.NewInt(15).MulRaw(1e18), // Buyer 2: 15 tokens
			math.NewInt(20).MulRaw(1e18), // Buyer 3: 20 tokens
		}

		// Fund all buyers
		for _, buyer := range buyers {
			suite.FundAcc(buyer, sdk.NewCoins(
				sdk.NewCoin(usdcDenom, math.NewInt(50000).MulRaw(1e6)),
			))
		}

		// Execute purchases from all buyers
		var (
			totalPurchased    math.Int  = math.ZeroInt()
			totalRaisedAmount sdk.Coins = sdk.NewCoins()
		)
		for buyerIdx, buyer := range buyers {
			paymentCoin, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountsToBuy[buyerIdx], usdcDenom)
			suite.Require().NoError(err)
			totalPurchased = totalPurchased.Add(amountsToBuy[buyerIdx])
			totalRaisedAmount = totalRaisedAmount.Add(paymentCoin)

			// Verify each purchase
			purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer.String())
			suite.Require().True(found)
			suite.Require().Equal(amountsToBuy[buyerIdx], purchase.Amount)
			suite.Require().True(purchase.Claimed.IsZero())
		}

		// Verify total auction sold amount
		auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)
		suite.Require().Equal(totalPurchased, auction.SoldAmount)
		suite.Require().Equal(totalRaisedAmount, auction.RaisedAmount)

		// End auction and move to vesting period
		suite.Ctx = suite.Ctx.WithBlockTime(auction.EndTime.Add(1 * time.Hour))
		err := suite.App.OTCBuybackKeeper.EndAuction(suite.Ctx, auctionID, "time_ended")
		suite.Require().NoError(err)

		// Move to vesting end
		auction, _ = suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Ctx = suite.Ctx.WithBlockTime(auction.GetVestingEndTime())

		// All buyers claim tokens
		var totalClaimed math.Int = math.ZeroInt()
		for buyerIdx, buyer := range buyers {
			claimedAmount, err := suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
			suite.Require().NoError(err)
			totalClaimed = totalClaimed.Add(claimedAmount)

			// Verify each purchase
			purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer.String())
			suite.Require().True(found)
			suite.Require().Equal(amountsToBuy[buyerIdx], purchase.Claimed)

			// verify buyer received the tokens
			balance := suite.App.BankKeeper.GetBalance(suite.Ctx, buyer, params.BaseDenom)
			suite.Require().Equal(claimedAmount, balance.Amount)
		}

		suite.Require().Equal(totalPurchased, totalClaimed)
	})
}

func (suite *KeeperTestSuite) TestBuyInvalidScenarios() {
	suite.Run("buy with invalid scenarios", func() {
		suite.SetupTest()

		// Create a buyer account with USDC balance
		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(10_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		auctionID := suite.CreateDefaultAuction()
		auction, _ := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)

		amountToBuy := math.NewInt(100).MulRaw(1e18)
		// Test 1: Buy with unaccepted token
		_, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "unaccepted_token")
		suite.Require().Error(err)

		// Test 2: Buy more than available allocation
		tooMuch := auction.Allocation.Add(math.NewInt(1))
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, tooMuch, "usdc")
		suite.Require().Error(err)

		// Test 3: Buy on non-active auction
		suite.Ctx = suite.Ctx.WithBlockTime(auction.StartTime.Add(-1 * time.Hour))
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc")
		suite.Require().Error(err)
		suite.Ctx = suite.Ctx.WithBlockTime(auction.StartTime.Add(1 * time.Hour))

		// Test 4: Buy with invalid amount
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(-1).MulRaw(1e18), "usdc")
		suite.Require().Error(err)

		// Happy path
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(100).MulRaw(1e18), "usdc")
		suite.Require().NoError(err)
	})
}

func (suite *KeeperTestSuite) TestBuyAll() {
	suite.Run("buy all", func() {
		suite.SetupTest()

		auctionID := suite.CreateDefaultAuction()
		auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)

		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(10_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		amountToBuy := auction.Allocation
		_, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc")
		suite.Require().NoError(err)

		// Verify purchase was created
		purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer.String())
		suite.Require().True(found)
		suite.Require().Equal(amountToBuy, purchase.Amount)
		suite.Require().True(purchase.Claimed.IsZero())

		// Verify auction is completed
		completedAuction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)
		suite.Require().True(completedAuction.IsCompleted(), "Auction should be completed")
	})
}

func (suite *KeeperTestSuite) TestMultipleClaims() {
	suite.Run("multiple claims", func() {
		suite.SetupTest()

		// Create a buyer account with USDC balance
		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(10_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		auctionID := suite.CreateDefaultAuction()
		auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)

		expectedVestingStartTime := auction.GetVestingStartTime()
		expectedVestingEndTime := auction.GetVestingEndTime()

		// buyer should not have a purchase yet
		_, found = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer.String())
		suite.Require().False(found)

		amountToBuy := math.NewInt(10).MulRaw(1e18)
		_, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc")
		suite.Require().NoError(err)

		// Test: Try to claim from non-completed auction (should fail)
		_, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "auction must be completed")

		// End the auction manually
		err = suite.App.OTCBuybackKeeper.EndAuction(suite.Ctx, auctionID, "auction_ended_time")
		suite.Require().NoError(err)

		// Test: Try to claim before vesting starts (should fail with no claimable tokens)
		suite.Ctx = suite.Ctx.WithBlockTime(expectedVestingStartTime.Add(-1 * time.Second))
		_, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "no tokens available to claim", "Should not be able to claim before vesting starts")

		// Claim in the middle of vesting period for partial vesting
		suite.Ctx = suite.Ctx.WithBlockTime(expectedVestingStartTime.Add(12 * time.Hour))
		claimedAmount, err := suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().NoError(err)
		suite.Require().True(claimedAmount.IsPositive(), "Should have claimed some tokens")

		// claim again without changing the time, should not be able to claim again
		_, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "no tokens available to claim")

		// move time forward to the end of the vesting period
		suite.Ctx = suite.Ctx.WithBlockTime(expectedVestingEndTime)
		_, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().NoError(err)

		updatedPurchase, _ := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer.String())
		suite.Require().Equal(amountToBuy.String(), updatedPurchase.Claimed.String(), "Purchase should show tokens as claimed")
	})
}

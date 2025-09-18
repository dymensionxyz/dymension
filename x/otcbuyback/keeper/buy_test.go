package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
	testutil "github.com/dymensionxyz/dymension/v3/testutil/math"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

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

	suite.Run("BuyExactSpend - spend exact amount and get calculated tokens", func() {
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

		// Spend exactly 11.25 USDC to buy ~50 DYM tokens (within auction allocation limit)
		exactPayment := sdk.NewCoin("usdc", math.NewInt(1125).MulRaw(1e4)) // 11.25 USDC
		tokensPurchased, err := suite.App.OTCBuybackKeeper.BuyExactSpend(suite.Ctx, buyer, auctionID, exactPayment)
		suite.Require().NoError(err)

		// At pool price 1:4 (0.25 USDC per DYM) with 10% discount = 0.225 USDC per DYM
		// 11.25 USDC / 0.225 = 50 DYM tokens exactly
		expectedTokens := math.NewInt(50).MulRaw(1e18)
		suite.Require().True(tokensPurchased.Equal(expectedTokens), "Should have purchased 50 DYM tokens, got: %s", tokensPurchased)

		// Verify buyer balance was debited exactly 11.25 USDC
		buyerBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, buyer, "usdc")
		expectedBalance := buyerFunds[0].Amount.Sub(exactPayment.Amount)
		suite.Require().Equal(expectedBalance, buyerBalance.Amount)

		// Verify the buyer has a purchase record
		purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer.String())
		suite.Require().True(found)
		suite.Require().Equal(tokensPurchased, purchase.Amount)
	})

	suite.Run("BuyExactSpend - payment too small", func() {
		suite.SetupTest()

		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(1_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		auctionID := suite.CreateDefaultAuction()

		// Try to spend a very small amount that won't buy any tokens
		tinyPayment := sdk.NewCoin("usdc", math.NewInt(1)) // 0.000001 USDC
		_, err := suite.App.OTCBuybackKeeper.BuyExactSpend(suite.Ctx, buyer, auctionID, tinyPayment)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "payment amount too small to purchase any tokens")
	})

	suite.Run("BuyExactSpend - different discount levels", func() {
		suite.SetupTest()

		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(10_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		// Fund the module account
		suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(1000)))

		// Create auction with higher initial discount (20%)
		vestingParams := types.Auction_VestingParams{
			VestingPeriod:               24 * time.Hour,
			VestingStartAfterAuctionEnd: 0,
		}
		pumpParams := types.Auction_PumpParams{
			StartTimeAfterAuctionEnd: time.Hour,
			EpochIdentifier:          "day",
			NumEpochsPaidOver:        30,
			NumOfPumpsPerEpoch:       1,
		}

		auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(suite.Ctx,
			common.DymUint64(1000),                  // allocation
			suite.Ctx.BlockTime(),                   // start time
			suite.Ctx.BlockTime().Add(24*time.Hour), // end time
			math.LegacyNewDecWithPrec(2, 1),         // 0.2 = 20% initial discount
			math.LegacyNewDecWithPrec(5, 1),         // 0.5 = 50% max discount
			vestingParams,
			pumpParams,
		)
		suite.Require().NoError(err)

		// Spend exactly 25 USDC with 20% discount
		exactPayment := sdk.NewCoin("usdc", math.NewInt(25).MulRaw(1e6))
		tokensPurchased, err := suite.App.OTCBuybackKeeper.BuyExactSpend(suite.Ctx, buyer, auctionID, exactPayment)
		suite.Require().NoError(err)

		// At pool price 1:4 (0.25 USDC per DYM) with 20% discount = 0.2 USDC per DYM
		// 25 USDC / 0.2 = 125 DYM tokens
		expectedTokens := math.NewInt(125).MulRaw(1e18)
		suite.Require().True(tokensPurchased.Equal(expectedTokens), "Should have purchased 125 DYM tokens with 20%% discount, got: %s", tokensPurchased)
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
			totalPurchased    = math.ZeroInt()
			totalRaisedAmount = sdk.NewCoins()
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
		totalClaimed := math.ZeroInt()
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

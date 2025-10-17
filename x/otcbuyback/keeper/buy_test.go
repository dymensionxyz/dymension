package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
	testutil "github.com/dymensionxyz/dymension/v3/testutil/math"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
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
		auctionID := suite.CreateDefaultLinearAuction()
		auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)

		discount, _, err := auction.GetDiscount(suite.Ctx.BlockTime(), 0)
		suite.Require().NoError(err)
		suite.Require().Equal(discount, auction.DiscountType.GetLinear().InitialDiscount) // 10% default

		amountToBuy := math.NewInt(100).MulRaw(1e18)
		paymentCoin, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc", 0)
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
		auctionID := suite.CreateDefaultLinearAuction()
		auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)

		discount, _, err := auction.GetDiscount(suite.Ctx.BlockTime(), 0)
		suite.Require().NoError(err)
		suite.Require().Equal(discount, auction.DiscountType.GetLinear().InitialDiscount) // 10% default

		// Spend exactly 11.25 USDC to buy ~50 DYM tokens (within auction allocation limit)
		exactPayment := sdk.NewCoin("usdc", math.NewInt(1125).MulRaw(1e4)) // 11.25 USDC
		tokensPurchased, err := suite.App.OTCBuybackKeeper.BuyExactSpend(suite.Ctx, buyer, auctionID, exactPayment, 0)
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
		purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().True(found)
		suite.Require().Equal(tokensPurchased, purchase.TotalAmount())
	})

	suite.Run("BuyExactSpend - payment too small", func() {
		suite.SetupTest()

		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(1_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		auctionID := suite.CreateDefaultLinearAuction()

		// Try to spend a very small amount that won't buy any tokens
		tinyPayment := sdk.NewCoin("usdc", math.NewInt(1)) // 0.000001 USDC
		_, err := suite.App.OTCBuybackKeeper.BuyExactSpend(suite.Ctx, buyer, auctionID, tinyPayment, 0)
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
			VestingDelay: 0,
		}
		pumpParams := types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		}

		discount := types.NewLinearDiscountType(
			math.LegacyNewDecWithPrec(2, 1), // 0.2 = 20% initial discount
			math.LegacyNewDecWithPrec(5, 1), // 0.5 = 50% max discount
			24*time.Hour,
		)

		auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(suite.Ctx,
			common.DymUint64(1000),                  // allocation
			suite.Ctx.BlockTime(),                   // start time
			suite.Ctx.BlockTime().Add(24*time.Hour), // end time
			discount,
			vestingParams,
			pumpParams,
		)
		suite.Require().NoError(err)

		// Spend exactly 25 USDC with 20% discount
		exactPayment := sdk.NewCoin("usdc", math.NewInt(25).MulRaw(1e6))
		tokensPurchased, err := suite.App.OTCBuybackKeeper.BuyExactSpend(suite.Ctx, buyer, auctionID, exactPayment, 0)
		suite.Require().NoError(err)

		// At pool price 1:4 (0.25 USDC per DYM) with 20% discount = 0.2 USDC per DYM
		// 25 USDC / 0.2 = 125 DYM tokens
		expectedTokens := math.NewInt(125).MulRaw(1e18)
		suite.Require().True(tokensPurchased.Equal(expectedTokens), "Should have purchased 125 DYM tokens with 20%% discount, got: %s", tokensPurchased)
	})

	suite.Run("fixed discount - verify correct discount applied", func() {
		suite.SetupTest()

		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(10_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		// Fund the module account
		suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(1000)))

		// Create fixed discount auction with two tiers
		discountType := types.NewFixedDiscountType([]types.FixedDiscount_Discount{
			{Discount: math.LegacyNewDecWithPrec(10, 2), VestingPeriod: 30 * 24 * time.Hour}, // 10%, 30d
			{Discount: math.LegacyNewDecWithPrec(30, 2), VestingPeriod: 90 * 24 * time.Hour}, // 30%, 90d
		})

		auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
			suite.Ctx,
			common.DymUint64(1000),
			suite.Ctx.BlockTime(),
			suite.Ctx.BlockTime().Add(24*time.Hour),
			discountType,
			types.Auction_VestingParams{VestingDelay: 0},
			types.DefaultPumpParams,
		)
		suite.Require().NoError(err)

		// Buy 100 DYM with 30-day vesting (10% discount)
		// Pool price: 1:4 = 0.25 USDC per DYM
		// With 10% discount: 0.25 * 0.9 = 0.225 USDC per DYM
		// For 100 DYM: 100 * 0.225 = 22.5 USDC
		amountToBuy := math.NewInt(100).MulRaw(1e18)
		paymentCoin1, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc", 30*24*time.Hour)
		suite.Require().NoError(err)

		expectedPayment1 := math.NewInt(225).MulRaw(1e5) // 22.5 USDC
		err = testutil.ApproxEqualRatio(expectedPayment1, paymentCoin1.Amount, 0.01)
		suite.Require().NoError(err, "Should have paid ~22.5 USDC with 10%% discount, got %s", paymentCoin1.Amount)

		// Buy 100 DYM with 90-day vesting (30% discount)
		// With 30% discount: 0.25 * 0.7 = 0.175 USDC per DYM
		// For 100 DYM: 100 * 0.175 = 17.5 USDC
		paymentCoin2, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc", 90*24*time.Hour)
		suite.Require().NoError(err)

		expectedPayment2 := math.NewInt(175).MulRaw(1e5) // 17.5 USDC
		err = testutil.ApproxEqualRatio(expectedPayment2, paymentCoin2.Amount, 0.01)
		suite.Require().NoError(err, "Should have paid ~17.5 USDC with 30%% discount, got %s", paymentCoin2.Amount)

		// Verify total purchase
		purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().True(found)
		suite.Require().Equal(2, len(purchase.Entries))
		suite.Require().Equal(math.NewInt(200).MulRaw(1e18), purchase.TotalAmount())
		suite.Require().Equal(30*24*time.Hour, purchase.Entries[0].VestingDuration)
		suite.Require().Equal(90*24*time.Hour, purchase.Entries[1].VestingDuration)
	})
}

func (suite *KeeperTestSuite) TestMultipleBuyersAndClaims() {
	suite.Run("multiple buyers and claims", func() {
		suite.SetupTest()

		usdcDenom := "usdc"

		auctionID := suite.CreateDefaultLinearAuction()

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
			paymentCoin, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountsToBuy[buyerIdx], usdcDenom, 0)
			suite.Require().NoError(err)
			totalPurchased = totalPurchased.Add(amountsToBuy[buyerIdx])
			totalRaisedAmount = totalRaisedAmount.Add(paymentCoin)

			// Verify each purchase
			purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
			suite.Require().True(found)
			suite.Require().Equal(amountsToBuy[buyerIdx], purchase.TotalAmount())
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
		// Vesting end is when the last auction purchase becomes claimable
		auction, _ = suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		l := auction.DiscountType.GetLinear()
		suite.Require().NotNil(l)
		suite.Ctx = suite.Ctx.WithBlockTime(auction.EndTime.Add(l.VestingPeriod))

		// All buyers claim tokens
		totalClaimed := math.ZeroInt()
		for buyerIdx, buyer := range buyers {
			claimedAmount, err := suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
			suite.Require().NoError(err)
			totalClaimed = totalClaimed.Add(claimedAmount)

			// Verify each purchase
			purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
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

		auctionID := suite.CreateDefaultLinearAuction()
		auction, _ := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)

		amountToBuy := math.NewInt(100).MulRaw(1e18)
		// Test 1: Buy with unaccepted token
		_, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "unaccepted_token", 0)
		suite.Require().Error(err)

		// Test 2: Buy more than available allocation
		tooMuch := auction.Allocation.Add(math.NewInt(1))
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, tooMuch, "usdc", 0)
		suite.Require().Error(err)

		// Test 3: Buy on non-active auction
		suite.Ctx = suite.Ctx.WithBlockTime(auction.StartTime.Add(-1 * time.Hour))
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc", 0)
		suite.Require().Error(err)
		suite.Ctx = suite.Ctx.WithBlockTime(auction.StartTime.Add(1 * time.Hour))

		// Test 4: Buy with invalid amount
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(-1).MulRaw(1e18), "usdc", 0)
		suite.Require().Error(err)

		// Happy path
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(100).MulRaw(1e18), "usdc", 0)
		suite.Require().NoError(err)
	})
}

func (suite *KeeperTestSuite) TestBuyAll() {
	suite.Run("buy all", func() {
		suite.SetupTest()

		auctionID := suite.CreateDefaultLinearAuction()
		auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)

		buyer := suite.CreateRandomAccount()
		buyerFunds := sdk.NewCoins(
			sdk.NewCoin("usdc", math.NewInt(10_000).MulRaw(1e6)),
		)
		suite.FundAcc(buyer, buyerFunds)

		amountToBuy := auction.Allocation
		_, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc", 0)
		suite.Require().NoError(err)

		// Verify purchase was created
		purchase, found := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().True(found)
		suite.Require().Equal(amountToBuy, purchase.TotalAmount())
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

		auctionID := suite.CreateDefaultLinearAuction()
		auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(found)

		expectedVestingStartTime := auction.StartTime
		auction, _ = suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		l := auction.DiscountType.GetLinear()
		suite.Require().NotNil(l)
		expectedVestingEndTime := auction.EndTime.Add(l.VestingPeriod)

		// buyer should not have a purchase yet
		_, found = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().False(found)

		amountToBuy := math.NewInt(10).MulRaw(1e18)
		_, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc", 0)
		suite.Require().NoError(err)

		// Test: Try to claim immediately after buy (should fail)
		_, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), types.ErrNoClaimableTokens.Error())

		// End the auction manually
		suite.Ctx = suite.Ctx.WithBlockTime(auction.EndTime)
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

		updatedPurchase, _ := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().Equal(amountToBuy.String(), updatedPurchase.Claimed.String(), "Purchase should show tokens as claimed")
	})
}

func (suite *KeeperTestSuite) TestFixedDiscountBuy() {
	suite.Run("buy with fixed discount", func() {
		suite.SetupTest()

		suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(1000)))

		discountType := types.NewFixedDiscountType([]types.FixedDiscount_Discount{
			{Discount: math.LegacyNewDecWithPrec(10, 2), VestingPeriod: 30 * 24 * time.Hour},
			{Discount: math.LegacyNewDecWithPrec(30, 2), VestingPeriod: 90 * 24 * time.Hour},
		})

		auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
			suite.Ctx,
			common.DymUint64(1000),
			suite.Ctx.BlockTime(),
			suite.Ctx.BlockTime().Add(24*time.Hour),
			discountType,
			types.DefaultVestingParams,
			types.DefaultPumpParams,
		)
		suite.Require().NoError(err)

		buyer := suite.CreateRandomAccount()
		suite.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(10000).MulRaw(1e6))))

		// Buy with 30-day vesting (10% discount)
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(100).MulRaw(1e18), "usdc", 30*24*time.Hour)
		suite.Require().NoError(err)

		purchase, _ := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().Equal(1, len(purchase.Entries))
		suite.Require().Equal(30*24*time.Hour, purchase.Entries[0].VestingDuration)

		// Buy with 90-day vesting (30% discount)
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(50).MulRaw(1e18), "usdc", 90*24*time.Hour)
		suite.Require().NoError(err)

		purchase, _ = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().Equal(2, len(purchase.Entries))

		// Try invalid vesting period
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(10).MulRaw(1e18), "usdc", 60*24*time.Hour)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "vesting period not found")
	})
}

func (suite *KeeperTestSuite) TestGovernanceParams() {
	suite.Run("max purchase number enforcement", func() {
		suite.SetupTest()

		params, _ := suite.App.OTCBuybackKeeper.GetParams(suite.Ctx)
		params.MaxPurchaseNumber = 3
		err := suite.App.OTCBuybackKeeper.SetParams(suite.Ctx, params)
		suite.Require().NoError(err)

		auctionID := suite.CreateDefaultLinearAuction()
		buyer := suite.CreateRandomAccount()
		suite.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(100000).MulRaw(1e6))))

		// Make 3 purchases (at limit)
		for i := 0; i < 3; i++ {
			_, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(1).MulRaw(1e18), "usdc", 0)
			suite.Require().NoError(err)
		}

		// 4th purchase should fail
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(1).MulRaw(1e18), "usdc", 0)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "maximum purchases")
	})

	suite.Run("min purchase amount enforcement", func() {
		suite.SetupTest()

		params, _ := suite.App.OTCBuybackKeeper.GetParams(suite.Ctx)
		params.MinPurchaseAmount = math.NewInt(10).MulRaw(1e18)
		err := suite.App.OTCBuybackKeeper.SetParams(suite.Ctx, params)
		suite.Require().NoError(err)

		auctionID := suite.CreateDefaultLinearAuction()
		buyer := suite.CreateRandomAccount()
		suite.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(10000).MulRaw(1e6))))

		// Try to buy less than minimum
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(5).MulRaw(1e18), "usdc", 0)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "less than minimum")

		// Buy exactly minimum (should work)
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(10).MulRaw(1e18), "usdc", 0)
		suite.Require().NoError(err)
	})

	suite.Run("auction end when remaining below minimum", func() {
		suite.SetupTest()

		params, _ := suite.App.OTCBuybackKeeper.GetParams(suite.Ctx)
		params.MinPurchaseAmount = math.NewInt(10).MulRaw(1e18)
		err := suite.App.OTCBuybackKeeper.SetParams(suite.Ctx, params)
		suite.Require().NoError(err)

		auctionID := suite.CreateDefaultLinearAuction()
		auction, _ := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)

		buyer := suite.CreateRandomAccount()
		suite.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(100000).MulRaw(1e6))))

		// Buy almost all, leaving 5 DYM (below minimum)
		amountToBuy := auction.Allocation.Sub(math.NewInt(5).MulRaw(1e18))
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, "usdc", 0)
		suite.Require().NoError(err)

		// Verify auction ended
		auction, _ = suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
		suite.Require().True(auction.IsCompleted())
	})
}

func (suite *KeeperTestSuite) TestOverlappingVestingBuyAndClaim() {
	suite.Run("multiple purchases and overlapping vesting", func() {
		suite.SetupTest()

		// Create auction with instant vesting (VestingDelay = 0)
		suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(1000)))
		discountType := types.NewLinearDiscountType(
			math.LegacyNewDecWithPrec(1, 1),
			math.LegacyNewDecWithPrec(5, 1),
			24*time.Hour,
		)

		auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
			suite.Ctx,
			common.DymUint64(1000),
			suite.Ctx.BlockTime(),
			suite.Ctx.BlockTime().Add(24*time.Hour),
			discountType,
			types.Auction_VestingParams{VestingDelay: 0},
			types.DefaultPumpParams,
		)
		suite.Require().NoError(err)

		buyer := suite.CreateRandomAccount()
		suite.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(100000).MulRaw(1e6))))

		// First purchase
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(100).MulRaw(1e18), "usdc", 0)
		suite.Require().NoError(err)

		// Advance time by 6 hours and make second purchase
		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(6 * time.Hour))
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(50).MulRaw(1e18), "usdc", 0)
		suite.Require().NoError(err)

		purchase, _ := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().Equal(2, len(purchase.Entries))
		suite.Require().Equal(math.NewInt(150).MulRaw(1e18), purchase.TotalAmount())

		// Advance to 12 hours from start (first entry 50% vested, second entry 25% vested)
		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(6 * time.Hour))

		// Claim during active auction
		claimedAmount, err := suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().NoError(err)
		// Should claim 50% of first entry == 50 DYM and 25% of second entry == 12.5 DYM
		suite.Require().Equal(math.NewInt(50).MulRaw(1e18).Add(math.NewInt(125).MulRaw(1e17)), claimedAmount)

		// Advance to end of all vesting
		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(24 * time.Hour))
		_, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().NoError(err)

		purchase, _ = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().Equal(purchase.TotalAmount(), purchase.Claimed)
	})

	suite.Run("fixed discount with different vesting periods", func() {
		suite.SetupTest()

		suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(1000)))

		const v30 = 30 * 24 * time.Hour
		const v90 = 90 * 24 * time.Hour

		// Create fixed discount auction with 30d and 90d vesting
		discountType := types.NewFixedDiscountType([]types.FixedDiscount_Discount{
			{Discount: math.LegacyNewDecWithPrec(10, 2), VestingPeriod: v30}, // 10%, 30d
			{Discount: math.LegacyNewDecWithPrec(30, 2), VestingPeriod: v90}, // 30%, 90d
		})

		auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(
			suite.Ctx,
			common.DymUint64(1000),
			suite.Ctx.BlockTime(),
			suite.Ctx.BlockTime().Add(24*time.Hour),
			discountType,
			types.Auction_VestingParams{VestingDelay: 0}, // instant vesting
			types.DefaultPumpParams,
		)
		suite.Require().NoError(err)

		buyer := suite.CreateRandomAccount()
		suite.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(100000).MulRaw(1e6))))

		// 1. First buy with 30-day vesting: 100 DYM
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(100).MulRaw(1e18), "usdc", v30)
		suite.Require().NoError(err)

		purchase, _ := suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().Equal(1, len(purchase.Entries))

		// 2. Check claiming right after first buy (vesting just started)
		claimable := purchase.ClaimableAmount(suite.Ctx.BlockTime())
		suite.Require().True(claimable.IsZero(), "No vesting yet at exact start time")

		// 3. Advance 15 days (50% of first vesting)
		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(15 * 24 * time.Hour))

		// 4. Second buy with 90-day vesting: 180 DYM
		_, err = suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, math.NewInt(180).MulRaw(1e18), "usdc", v90)
		suite.Require().NoError(err)

		purchase, _ = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		suite.Require().Equal(2, len(purchase.Entries))

		// 5. Check claiming right after second buy (should only have from first)
		claimable = purchase.ClaimableAmount(suite.Ctx.BlockTime())
		expected := math.NewInt(50).MulRaw(1e18) // 50% of 100 DYM from first entry
		suite.Require().Equal(expected, claimable, "Should have 50 DYM vested from first entry")

		claimedAmount, err := suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().NoError(err)
		suite.Require().Equal(expected, claimedAmount)

		// 6. Advance to 30 days total (first vesting ends, 16.67% of second vesting)
		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(15 * 24 * time.Hour))

		purchase, _ = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		claimable = purchase.ClaimableAmount(suite.Ctx.BlockTime())
		// First: 100% of 100 = 100 DYM (but 50 already claimed)
		// Second: 15 days / 90 days = 16.67% of 180 ≈ 30 DYM (with small truncation from LegacyDec)
		// Total unlocked: ≈ 130 DYM, minus 50 claimed ≈ 80 claimable
		// Allow small tolerance for LegacyDec truncation (1/6 * 180 has precision loss)
		expected = math.NewInt(80).MulRaw(1e18)
		err = testutil.ApproxEqualRatio(expected, claimable, 0.0001)
		suite.Require().NoError(err, "Claimable should be approximately 80 DYM, got %s", claimable)

		claimedAmount, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().NoError(err)
		// Claimable and claimed should match exactly
		suite.Require().Equal(claimable, claimedAmount)

		// 7. Advance to 60 days total (first done, 50% of second vesting)
		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(30 * 24 * time.Hour))

		purchase, _ = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		claimable = purchase.ClaimableAmount(suite.Ctx.BlockTime())
		// First: 100% = 100 DYM
		// Second: 45 days / 90 days = 50% of 180 = 90 DYM (exact)
		// Total unlocked: 190 DYM, claimed so far: ≈ 130 DYM (with precision loss from checkpoint 6)
		// Claimable: ≈ 60 DYM
		expected = math.NewInt(60).MulRaw(1e18)
		err = testutil.ApproxEqualRatio(expected, claimable, 0.0001)
		suite.Require().NoError(err, "Claimable should be approximately 60 DYM, got %s", claimable)

		claimedAmount, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().NoError(err)
		suite.Require().Equal(claimable, claimedAmount)

		// 8. Advance to 105 days total (both fully vested)
		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(45 * 24 * time.Hour))

		purchase, _ = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		claimable = purchase.ClaimableAmount(suite.Ctx.BlockTime())
		// Total: 280 DYM, claimed: ≈ 190 DYM (with cumulative precision effects)
		// Remaining: ≈ 90 DYM
		expected = math.NewInt(90).MulRaw(1e18)
		err = testutil.ApproxEqualRatio(expected, claimable, 0.0001)
		suite.Require().NoError(err, "Claimable should be approximately 90 DYM, got %s", claimable)

		claimedAmount, err = suite.App.OTCBuybackKeeper.ClaimVestedTokens(suite.Ctx, buyer, auctionID)
		suite.Require().NoError(err)
		suite.Require().Equal(claimable, claimedAmount)

		// Verify all claimed (with tolerance for precision)
		purchase, _ = suite.App.OTCBuybackKeeper.GetPurchase(suite.Ctx, auctionID, buyer)
		err = testutil.ApproxEqualRatio(purchase.TotalAmount(), purchase.Claimed, 0.0001)
		suite.Require().NoError(err, "All tokens should be claimed (within tolerance), total: %s, claimed: %s",
			purchase.TotalAmount(), purchase.Claimed)
	})
}

package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

func (suite *KeeperTestSuite) TestMovingAverageUpdates() {
	suite.SetupTest()

	initialPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, 1, "usdc", "adym")
	suite.Require().NoError(err)

	tokenData, err := suite.App.OTCBuybackKeeper.GetAcceptedTokenData(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	suite.Require().Equal(initialPrice, tokenData.LastAveragePrice)
	suite.Require().Equal(initialPrice, math.LegacyMustNewDecFromStr("0.25").QuoInt64(1e12)) // 1e12 is the precision of the price

	// call update moving average price without any changes
	err = suite.App.OTCBuybackKeeper.UpdateMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	price, err := suite.App.OTCBuybackKeeper.GetMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	suite.Require().Equal(initialPrice, price)

	// change the pool price
	// instead of overriding the pool itself, we create new one with 1:10 ratio
	poolID := suite.PreparePoolWithCoins(sdk.NewCoins(
		sdk.NewCoin("usdc", math.NewInt(100_000).MulRaw(1e6)),
		sdk.NewCoin("adym", math.NewInt(1_000_000).MulRaw(1e18)),
	))

	newTokenData := types.TokenData{
		PoolId:           poolID,
		LastAveragePrice: tokenData.LastAveragePrice,
	}
	err = suite.App.OTCBuybackKeeper.SetAcceptedToken(suite.Ctx, "usdc", newTokenData)
	suite.Require().NoError(err)
	err = suite.App.OTCBuybackKeeper.UpdateMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)

	// Assert that the price change follows the EMA formula with smoothing factor
	// new_avg = alpha * current_price + (1 - alpha) * old_avg
	// expeted new price is 0.1 * 0.1 + 0.9 * 0.25 = 0.235
	newPrice, err := suite.App.OTCBuybackKeeper.GetMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	expectedNewPrice := math.LegacyMustNewDecFromStr("0.235").QuoInt64(1e12)
	suite.Require().Equal(expectedNewPrice, newPrice, "new moving average should match EMA calculation: alpha * spot_price + (1-alpha) * old_avg")

	// update the smoothing factor to 0.2
	params := types.Params{
		MovingAverageSmoothingFactor: math.LegacyMustNewDecFromStr("0.2"),
	}
	err = suite.App.OTCBuybackKeeper.SetParams(suite.Ctx, params)
	suite.Require().NoError(err)
	err = suite.App.OTCBuybackKeeper.UpdateMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	newPrice, err = suite.App.OTCBuybackKeeper.GetMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)

	// expeted new price is 0.2 * 0.1 + 0.8 * 0.235 = 0.208
	expectedNewPrice = math.LegacyMustNewDecFromStr("0.208").QuoInt64(1e12)
	suite.Require().Equal(expectedNewPrice, newPrice, "new moving average should match EMA calculation: alpha * spot_price + (1-alpha) * old_avg")

	// update the smoothing factor to 0.5
	params = types.Params{
		MovingAverageSmoothingFactor: math.LegacyMustNewDecFromStr("0.5"),
	}
	err = suite.App.OTCBuybackKeeper.SetParams(suite.Ctx, params)
	suite.Require().NoError(err)

	// update the pool price back to 1:4
	tokenData, err = suite.App.OTCBuybackKeeper.GetAcceptedTokenData(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	poolID = suite.PreparePoolWithCoins(sdk.NewCoins(
		sdk.NewCoin("usdc", math.NewInt(500_000).MulRaw(1e6)),
		sdk.NewCoin("adym", math.NewInt(2_000_000).MulRaw(1e18)),
	))
	err = suite.App.OTCBuybackKeeper.SetAcceptedToken(suite.Ctx, "usdc", types.TokenData{
		PoolId:           poolID,
		LastAveragePrice: tokenData.LastAveragePrice,
	})
	suite.Require().NoError(err)
	err = suite.App.OTCBuybackKeeper.UpdateMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	newPrice, err = suite.App.OTCBuybackKeeper.GetMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)

	// expeted new price is 0.5 * 0.25 + 0.5 * 0.208 = 0.229
	expectedNewPrice = math.LegacyMustNewDecFromStr("0.229").QuoInt64(1e12)
	suite.Require().Equal(expectedNewPrice, newPrice, "new moving average should match EMA calculation: alpha * spot_price + (1-alpha) * old_avg")
}

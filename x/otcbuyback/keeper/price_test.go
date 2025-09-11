package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

// FIXME: test it as funct of smoothing factor

func (suite *KeeperTestSuite) TestMovingAverageUpdates() {
	suite.SetupTest()

	initialPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, 1, "usdc", "adym")
	suite.Require().NoError(err)

	tokenData, err := suite.App.OTCBuybackKeeper.GetAcceptedTokenData(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	suite.Require().Equal(initialPrice, tokenData.LastAveragePrice)

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

	// we expect the new price to be lower, as the price reduced from 1:4 to 1:10
	// but still higher than the actual price
	newPrice, err := suite.App.OTCBuybackKeeper.GetMovingAveragePrice(suite.Ctx, "usdc")
	suite.Require().NoError(err)
	suite.Require().True(newPrice.LT(price), "new price %s should be less than old price %s", newPrice, price)

	currSpotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolID, "usdc", "adym")
	suite.Require().NoError(err)
	suite.Require().True(currSpotPrice.LT(newPrice), "current spot price %s should be less than new price %s", currSpotPrice, newPrice)
}

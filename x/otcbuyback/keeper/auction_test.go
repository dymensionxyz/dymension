package keeper_test

import (
	"time"
)

// UT auction ends due to time passing
func (suite *KeeperTestSuite) TestAuctionLifecycle() {

	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	auctionID := suite.CreateDefaultAuction()
	auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found, "auction should be found")
	suite.Require().False(auction.IsCompleted(), "auction should not be completed")

	// few seconds before the auction starts
	before := suite.Ctx.BlockTime().Add(-1 * time.Second)
	suite.Require().False(auction.IsActive(before), "auction should be active")
	suite.Require().True(auction.IsUpcoming(before), "auction should be upcoming")

	// few seconds after the auction starts
	after := suite.Ctx.BlockTime().Add(1 * time.Second)
	suite.Require().True(auction.IsActive(after), "auction should be active")
	suite.Require().False(auction.IsUpcoming(after), "auction should be upcoming")

	// Advance time to end the auction
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(24 * time.Hour).Add(1 * time.Second))
	err := suite.App.OTCBuybackKeeper.BeginBlock(suite.Ctx)
	suite.Require().NoError(err, "begin block should not error")

	auction, found = suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, auctionID)
	suite.Require().True(found, "auction should be found")
	suite.Require().True(auction.IsCompleted(), "auction should be completed")
	suite.Require().False(auction.IsActive(suite.Ctx.BlockTime()), "auction should be active")
	suite.Require().False(auction.IsUpcoming(suite.Ctx.BlockTime()), "auction should be upcoming")

	// FIXME: test auction remaining funds are returned to streamer

	// FIXME: pump streams are created
}

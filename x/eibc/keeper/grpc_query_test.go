package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestParamsQuery() {
	suite.SetupTest()
	wctx := sdk.WrapSDKContext(suite.Ctx)
	params := types.DefaultParams()
	suite.App.EIBCKeeper.SetParams(suite.Ctx, params)

	response, err := suite.queryClient.Params(wctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}

func (suite *KeeperTestSuite) TestQueryDemandOrderById() {
	suite.SetupTest()
	keeper := suite.App.EIBCKeeper

	// Validate demand order query with empty request
	res, err := suite.queryClient.DemandOrderById(sdk.WrapSDKContext(suite.Ctx), &types.QueryGetDemandOrderRequest{})
	suite.Require().Error(err)
	suite.Require().Nil(res)

	// Create a demand order with status pending
	recipientAddress := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(1000))[0]
	demandOrder, err := types.NewDemandOrder(*rollappPacket, "150", "50", "stake", recipientAddress.String())
	suite.Require().NoError(err)
	keeper.SetDemandOrder(suite.Ctx, demandOrder)

	// Query the demand order by its ID
	res, err = suite.queryClient.DemandOrderById(sdk.WrapSDKContext(suite.Ctx), &types.QueryGetDemandOrderRequest{Id: demandOrder.Id})
	suite.Require().NoError(err)
	suite.Require().NotNil(res.DemandOrder)
	suite.Require().Equal(demandOrder, res.DemandOrder)

}

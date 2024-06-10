package keeper_test

import (
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
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
	demandOrder := types.NewDemandOrder(*rollappPacket, math.NewIntFromUint64(150), math.NewIntFromUint64(50), "stake", recipientAddress.String())
	err = keeper.SetDemandOrder(suite.Ctx, demandOrder)
	suite.Require().NoError(err)

	// Query the demand order by its ID
	res, err = suite.queryClient.DemandOrderById(sdk.WrapSDKContext(suite.Ctx), &types.QueryGetDemandOrderRequest{Id: demandOrder.Id})
	suite.Require().NoError(err)
	suite.Require().NotNil(res.DemandOrder)
	suite.Require().Equal(demandOrder, res.DemandOrder)
}

func (suite *KeeperTestSuite) TestQueryDemandOrdersByStatus() {
	suite.SetupTest()
	keeper := suite.App.EIBCKeeper

	// Define the number of demand orders and create addresses
	demandOrdersNum := 3
	demandOrderAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, demandOrdersNum, math.NewInt(1000))

	// Define statuses to test
	statuses := []commontypes.Status{commontypes.Status_PENDING, commontypes.Status_REVERTED, commontypes.Status_FINALIZED}

	// Create and set demand orders for each status
	for i, status := range statuses {

		rollappPacket := &commontypes.RollappPacket{
			RollappId:   "testRollappId" + strconv.Itoa(i),
			Status:      status,
			ProofHeight: 2,
			Packet:      &packet,
		}

		// Use a unique address for each demand order
		recipientAddress := demandOrderAddresses[i].String()

		demandOrder := types.NewDemandOrder(*rollappPacket, math.NewIntFromUint64(150), math.NewIntFromUint64(50), "stake", recipientAddress)
		// Assert needed type of status for packet
		demandOrder.TrackingPacketStatus = status

		err := keeper.SetDemandOrder(suite.Ctx, demandOrder)
		suite.Require().NoError(err)

		// Query demand orders by status
		res, err := suite.queryClient.DemandOrdersByStatus(sdk.WrapSDKContext(suite.Ctx), &types.QueryDemandOrdersByStatusRequest{Status: status})
		suite.Require().NoError(err)
		suite.Require().NotNil(res.DemandOrders)
		suite.Require().Len(res.DemandOrders, 1, fmt.Sprintf("Expected 1 demand order for status %s, but got %d", status, len(res.DemandOrders)))
		suite.Require().Equal(demandOrder.Id, res.DemandOrders[0].Id)
	}

	// Query with invalid status should return an error
	res, err := suite.queryClient.DemandOrdersByStatus(sdk.WrapSDKContext(suite.Ctx), &types.QueryDemandOrdersByStatusRequest{Status: -1})
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

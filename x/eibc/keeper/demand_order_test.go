package keeper_test

import (
	"strconv"

	"cosmossdk.io/math"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestListDemandOrdersByStatus() {
	keeper := suite.App.EIBCKeeper
	ctx := suite.Ctx
	demandOrdersNum := 5
	demandOrderAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, demandOrdersNum, math.NewInt(1000))
	// Create and set some demand orders with status pending
	for i := 0; i < demandOrdersNum; i++ {
		rollappPacket := &commontypes.RollappPacket{
			RollappId:   "testRollappId" + strconv.Itoa(i),
			Status:      commontypes.Status_PENDING,
			ProofHeight: 2,
			Packet:      &packet,
		}
		demandOrder := types.NewDemandOrder(*rollappPacket, math.NewIntFromUint64(150), math.NewIntFromUint64(50), "stake", demandOrderAddresses[i].String())
		err := keeper.SetDemandOrder(ctx, demandOrder)
		suite.Require().NoError(err)
	}
	// Get the demand orders with status active
	demandOrders, err := keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_PENDING)
	suite.Require().NoError(err)
	suite.Equal(demandOrdersNum, len(demandOrders))

	// Update 3 of the demand orders to status finalized
	for _, demandOrder := range demandOrders[:3] {
		_, err = keeper.UpdateDemandOrderWithStatus(ctx, demandOrder, commontypes.Status_FINALIZED)
		suite.Require().NoError(err)
	}
	// Retrieve the updated demand orders after status change
	updatedDemandOrders, err := keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_FINALIZED)
	suite.Require().NoError(err)
	// Validate that there are exactly demandOrderNum packets in total
	pendingDemandOrders, err := keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_PENDING)
	suite.Require().NoError(err)
	totalDemandOrders := len(updatedDemandOrders) + len(pendingDemandOrders)
	suite.Require().Equal(demandOrdersNum, totalDemandOrders)
}

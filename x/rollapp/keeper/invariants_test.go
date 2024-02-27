package keeper_test

import (
	"strconv"

	"cosmossdk.io/math"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *RollappTestSuite) TestInvariants() {
	keeper := suite.App.EIBCKeeper
	ctx := suite.Ctx
	demandOrdersNum := 10
	demandOrderAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, demandOrdersNum, math.NewInt(1000))
	// Create and set some demand orders with status pending
	for i := 0; i < demandOrdersNum; i++ {
		var status commontypes.Status
		switch i % 3 {
		case 0:
			status = commontypes.Status_PENDING
		case 1:
			status = commontypes.Status_REVERTED
		case 2:
			status = commontypes.Status_FINALIZED
		}
		rollappPacket := &commontypes.RollappPacket{
			RollappId:   "testRollappId" + strconv.Itoa(i),
			Status:      status,
			ProofHeight: 2,
			Packet:      &packet,
		}
		err := suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
		suite.Require().NoError(err)
		demandOrder, err := types.NewDemandOrder(*rollappPacket, "150", "50", "stake", demandOrderAddresses[i].String())
		suite.Require().NoError(err)
		keeper.SetDemandOrder(ctx, demandOrder)
	}

	// check invariant
	suite.Require().NotPanics(func() {
		eibckeeper.DemandOrderCountInvariant(suite.App.EIBCKeeper)(ctx)
		eibckeeper.UnderlyingPacketExistInvariant(suite.App.EIBCKeeper)(ctx)
	})
}

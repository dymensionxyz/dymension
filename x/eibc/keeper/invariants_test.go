package keeper_test

import (
	"strconv"

	"cosmossdk.io/math"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestInvariants() {
	keeper := suite.App.EIBCKeeper
	ctx := suite.Ctx
	demandOrdersNum := 10
	demandOrderAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, demandOrdersNum, math.NewInt(1000))
	// Create and set some demand orders with status pending
	for i := 0; i < demandOrdersNum; i++ {
		var status commontypes.Status
		switch i % 2 {
		case 0:
			status = commontypes.Status_PENDING
		case 1:
			status = commontypes.Status_FINALIZED
		}
		rollappPacket := &commontypes.RollappPacket{
			RollappId:   "testRollappId" + strconv.Itoa(i),
			Status:      status,
			ProofHeight: 2,
			Packet:      &packet,
		}
		suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
		demandOrder := types.NewDemandOrder(*rollappPacket, math.NewIntFromUint64(150), math.NewIntFromUint64(50), "stake", demandOrderAddresses[i].String(), 1, nil)
		err := keeper.SetDemandOrder(ctx, demandOrder)
		suite.Require().NoError(err)
	}

	suite.Require().NotPanics(func() {
		_, broken := eibckeeper.AllInvariants(suite.App.EIBCKeeper)(ctx)
		suite.False(broken)
	})
}

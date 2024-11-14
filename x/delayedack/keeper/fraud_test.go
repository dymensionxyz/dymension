package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func (suite *DelayedAckTestSuite) TestHandleFraud() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	rollappId := "testRollappId"
	pkts := apptesting.GenerateRollappPackets(suite.T(), rollappId, 10)
	rollappId2 := "testRollappId2"
	pkts2 := apptesting.GenerateRollappPackets(suite.T(), rollappId2, 10)
	prefixPending1 := types.ByRollappIDByStatus(rollappId, commontypes.Status_PENDING)
	prefixPending2 := types.ByRollappIDByStatus(rollappId2, commontypes.Status_PENDING)
	prefixFinalized1 := types.ByRollappIDByStatus(rollappId, commontypes.Status_FINALIZED)
	prefixFinalized2 := types.ByRollappIDByStatus(rollappId, commontypes.Status_FINALIZED)

	for _, pkt := range append(pkts, pkts2...) {
		keeper.SetRollappPacket(ctx, pkt)
		keeper.MustSetPendingPacketByAddress(ctx, apptesting.TestPacketReceiver, pkt.RollappPacketKey())
	}

	suite.Require().Equal(10, len(keeper.ListRollappPackets(ctx, prefixPending1)))
	suite.Require().Equal(10, len(keeper.ListRollappPackets(ctx, prefixPending2)))
	pktsByAddr, err := keeper.GetPendingPacketsByAddress(ctx, apptesting.TestPacketReceiver)
	suite.Require().NoError(err)
	suite.Require().Equal(20, len(pktsByAddr))

	// finalize one packet
	_, err = keeper.UpdateRollappPacketAfterFinalization(ctx, pkts[0])
	suite.Require().Nil(err)
	_, err = keeper.UpdateRollappPacketAfterFinalization(ctx, pkts2[0])
	suite.Require().Nil(err)

	// call fraud on the 4 packet
	err = keeper.OnHardFork(ctx, rollappId, 4)
	suite.Require().NoError(err)

	// expected result:
	// rollappId:
	// - packet 1 are finalized
	// - packet 2-3 are still pending
	// - packets 4-10 are deleted
	// rollappId2:
	// - packet 1 are finalized
	// - packets 2-10 are still pending

	suite.Require().Equal(1, len(keeper.ListRollappPackets(ctx, prefixFinalized1)))
	suite.Require().Equal(2, len(keeper.ListRollappPackets(ctx, prefixPending1)))

	pktsByAddr, err = keeper.GetPendingPacketsByAddress(ctx, apptesting.TestPacketReceiver)
	suite.Require().NoError(err)
	suite.Require().Equal(11, len(pktsByAddr)) // 2 from rollappId, 9 from rollappId2

	suite.Require().Equal(1, len(keeper.ListRollappPackets(ctx, prefixFinalized2)))
	suite.Require().Equal(9, len(keeper.ListRollappPackets(ctx, prefixPending2)))
}

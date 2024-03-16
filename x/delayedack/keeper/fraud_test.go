package keeper_test

import (
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func (suite *KeeperTestSuite) TestHandleFraud() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	rollappId := "testRollappId"
	pkts := generatePackets(rollappId, 5)
	rollappId2 := "testRollappId2"
	pkts2 := generatePackets(rollappId2, 5)

	for _, pkt := range append(pkts, pkts2...) {
		keeper.SetRollappPacket(ctx, pkt)
	}

	suite.Require().Equal(10, len(keeper.GetAllRollappPackets(ctx)))
	suite.Require().Equal(10, len(keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_PENDING, 0)))

	//finalize some packets
	keeper.UpdateRollappPacketWithStatus(ctx, pkts[0], commontypes.Status_FINALIZED)
	keeper.UpdateRollappPacketWithStatus(ctx, pkts2[0], commontypes.Status_FINALIZED)

	err := keeper.HandleFraud(ctx, rollappId)
	suite.Require().Nil(err)

	suite.Require().Equal(10, len(keeper.GetAllRollappPackets(ctx)))
	suite.Require().Equal(4, len(keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_PENDING, 0)))
	suite.Require().Equal(4, len(keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_REVERTED, 0)))
	suite.Require().Equal(2, len(keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_FINALIZED, 0)))
}

/* ---------------------------------- utils --------------------------------- */

func generatePackets(rollappId string, num uint64) []commontypes.RollappPacket {
	var packets []commontypes.RollappPacket
	for i := uint64(0); i < num; i++ {
		packets = append(packets, commontypes.RollappPacket{
			RollappId: rollappId,
			Packet: &channeltypes.Packet{
				SourcePort:         "testSourcePort",
				SourceChannel:      "testSourceChannel",
				DestinationPort:    "testDestinationPort",
				DestinationChannel: "testDestinationChannel",
				Data:               []byte("testData"),
				Sequence:           i,
			},
			Status:      commontypes.Status_PENDING,
			ProofHeight: i,
		})
	}
	return packets
}

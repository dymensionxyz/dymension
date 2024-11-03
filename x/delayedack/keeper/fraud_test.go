package keeper_test

import (
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func (suite *DelayedAckTestSuite) TestHandleFraud() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	rollappId := "testRollappId"
	pkts := generatePackets(rollappId, 10)
	rollappId2 := "testRollappId2"
	pkts2 := generatePackets(rollappId2, 10)
	prefixPending1 := types.ByRollappIDByStatus(rollappId, commontypes.Status_PENDING)
	prefixPending2 := types.ByRollappIDByStatus(rollappId2, commontypes.Status_PENDING)
	prefixFinalized1 := types.ByRollappIDByStatus(rollappId, commontypes.Status_FINALIZED)
	prefixFinalized2 := types.ByRollappIDByStatus(rollappId, commontypes.Status_FINALIZED)

	for _, pkt := range append(pkts, pkts2...) {
		keeper.SetRollappPacket(ctx, pkt)
	}

	suite.Require().Equal(10, len(keeper.ListRollappPackets(ctx, prefixPending1)))
	suite.Require().Equal(10, len(keeper.ListRollappPackets(ctx, prefixPending2)))

	// finalize one packet
	_, err := keeper.UpdateRollappPacketWithStatus(ctx, pkts[0], commontypes.Status_FINALIZED)
	suite.Require().Nil(err)
	_, err = keeper.UpdateRollappPacketWithStatus(ctx, pkts2[0], commontypes.Status_FINALIZED)
	suite.Require().Nil(err)

	// call fraud on the 4 packet
	err = keeper.OnHardFork(ctx, rollappId, 4)
	suite.Require().Nil(err)

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

	suite.Require().Equal(1, len(keeper.ListRollappPackets(ctx, prefixFinalized2)))
	suite.Require().Equal(9, len(keeper.ListRollappPackets(ctx, prefixPending2)))
}

/* ---------------------------------- utils --------------------------------- */

func generatePackets(rollappId string, num uint64) []commontypes.RollappPacket {
	var packets []commontypes.RollappPacket
	for i := uint64(1); i <= num; i++ {
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

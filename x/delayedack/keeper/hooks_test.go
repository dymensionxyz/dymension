package keeper_test

import (
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// TestAfterEpochEnd tests that the finalized of rollapp packets
// are deleted given the correct epoch identifier
func (suite *KeeperTestSuite) TestAfterEpochEnd() {
	tests := []struct {
		name                 string
		pendingPacketsNum    int
		finalizePacketsNum   int
		epochIdentifierParam string
		epochIdentifier      string
		expectedDeleted      int
		expectedTotal        int
	}{
		{
			name:                 "delete rollapp packets after epoch end",
			pendingPacketsNum:    5,
			finalizePacketsNum:   3,
			epochIdentifierParam: "minute",
			epochIdentifier:      "minute",
			expectedDeleted:      3,
			expectedTotal:        2,
		},
		{
			name:                 "fail delete rollapp packets after epoch end - invalid epoch identifier",
			pendingPacketsNum:    5,
			finalizePacketsNum:   3,
			epochIdentifierParam: "minute",
			epochIdentifier:      "hour",
			expectedDeleted:      0,
			expectedTotal:        5,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
			for i := 1; i <= tc.pendingPacketsNum; i++ {
				rollappPacket := &commontypes.RollappPacket{
					RollappId: "testRollappId",
					Packet: &channeltypes.Packet{
						SourcePort:         "testSourcePort",
						SourceChannel:      "testSourceChannel",
						DestinationPort:    "testDestinationPort",
						DestinationChannel: "testDestinationChannel",
						Data:               []byte("testData"),
						Sequence:           uint64(i),
					},
					Status:      commontypes.Status_PENDING,
					ProofHeight: uint64(i * 2),
				}
				keeper.SetRollappPacket(ctx, *rollappPacket)
			}

			rollappPackets := keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_PENDING, 0)
			suite.Require().Equal(tc.pendingPacketsNum, len(rollappPackets))

			for _, rollappPacket := range rollappPackets[:tc.finalizePacketsNum] {
				keeper.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_FINALIZED)
			}
			finalizedRollappPackets := keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_FINALIZED, 0)
			suite.Require().Equal(tc.finalizePacketsNum, len(finalizedRollappPackets))

			keeper.SetParams(ctx, types.Params{EpochIdentifier: tc.epochIdentifierParam})
			epochHooks := keeper.GetEpochHooks()
			epochHooks.AfterEpochEnd(ctx, tc.epochIdentifier, 1)

			finalizedRollappPackets = keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_FINALIZED, 0)
			suite.Require().Equal(tc.finalizePacketsNum-tc.expectedDeleted, len(finalizedRollappPackets))

			totalRollappPackets := len(finalizedRollappPackets) + len(keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_PENDING, 0))
			suite.Require().Equal(tc.expectedTotal, totalRollappPackets)
		})
	}
}

func (suite *KeeperTestSuite) TestDeletionOfRevertedPackets() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx

	rollappId := "testRollappId"
	pkts := generatePackets(rollappId, 5)
	rollappId2 := "testRollappId2"
	pkts2 := generatePackets(rollappId2, 5)

	for _, pkt := range append(pkts, pkts2...) {
		keeper.SetRollappPacket(ctx, pkt)
	}

	err := keeper.HandleFraud(ctx, rollappId)
	suite.Require().Nil(err)

	suite.Require().Equal(10, len(keeper.GetAllRollappPackets(ctx)))

	keeper.SetParams(ctx, types.Params{EpochIdentifier: "minute"})
	epochHooks := keeper.GetEpochHooks()
	epochHooks.AfterEpochEnd(ctx, "minute", 1)

	suite.Require().Equal(5, len(keeper.GetAllRollappPackets(ctx)))
}

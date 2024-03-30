package keeper_test

import (
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	dkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// TestAfterEpochEnd tests that the finalized of rollapp packets
// are deleted given the correct epoch identifier
func (suite *DelayedAckTestSuite) TestAfterEpochEnd() {
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

	const rollappID = "testRollappId"

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
			for i := 1; i <= tc.pendingPacketsNum; i++ {
				rollappPacket := &commontypes.RollappPacket{
					RollappId: rollappID,
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
				err := keeper.SetRollappPacket(ctx, *rollappPacket)
				suite.Require().NoError(err)
			}

			rollappPackets := keeper.ListRollappPackets(ctx, dkeeper.ByRollappIDAndStatus(rollappID, commontypes.Status_PENDING))
			suite.Require().Equal(tc.pendingPacketsNum, len(rollappPackets))

			for _, rollappPacket := range rollappPackets[:tc.finalizePacketsNum] {
				_, err := keeper.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_FINALIZED)
				suite.Require().NoError(err)
			}
			finalizedRollappPackets := keeper.ListRollappPackets(ctx, dkeeper.ByRollappIDAndStatus(rollappID, commontypes.Status_FINALIZED))
			suite.Require().Equal(tc.finalizePacketsNum, len(finalizedRollappPackets))

			keeper.SetParams(ctx, types.Params{EpochIdentifier: tc.epochIdentifierParam})
			epochHooks := keeper.GetEpochHooks()
			err := epochHooks.AfterEpochEnd(ctx, tc.epochIdentifier, 1)
			suite.Require().NoError(err)

			finalizedRollappPackets = keeper.ListRollappPackets(ctx, dkeeper.ByRollappIDAndStatus(rollappID, commontypes.Status_FINALIZED))
			suite.Require().Equal(tc.finalizePacketsNum-tc.expectedDeleted, len(finalizedRollappPackets))

			pendingPackets := keeper.ListRollappPackets(ctx, dkeeper.ByRollappIDAndStatus(rollappID, commontypes.Status_PENDING))
			totalRollappPackets := len(finalizedRollappPackets) + len(pendingPackets)
			suite.Require().Equal(tc.expectedTotal, totalRollappPackets)
		})
	}
}

func (suite *DelayedAckTestSuite) TestDeletionOfRevertedPackets() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx

	rollappId := "testRollappId"
	pkts := generatePackets(rollappId, 5)
	rollappId2 := "testRollappId2"
	pkts2 := generatePackets(rollappId2, 5)
	prefixAll := dkeeper.AllRollappPackets()

	for _, pkt := range append(pkts, pkts2...) {
		err := keeper.SetRollappPacket(ctx, pkt)
		suite.Require().NoError(err)
	}

	err := keeper.HandleFraud(ctx, rollappId)
	suite.Require().Nil(err)

	suite.Require().Equal(10, len(keeper.ListRollappPackets(ctx, prefixAll)))

	keeper.SetParams(ctx, types.Params{EpochIdentifier: "minute"})
	epochHooks := keeper.GetEpochHooks()
	err = epochHooks.AfterEpochEnd(ctx, "minute", 1)
	suite.Require().NoError(err)

	suite.Require().Equal(5, len(keeper.ListRollappPackets(ctx, prefixAll)))
}

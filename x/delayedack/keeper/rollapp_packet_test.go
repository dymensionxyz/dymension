package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func (suite *DelayedAckTestSuite) TestRollappPacketEvents() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	tests := []struct {
		name                               string
		rollappPacket                      commontypes.RollappPacket
		rollappUpdatedStatus               commontypes.Status
		rollappUpdateError                 error
		expectedEventType                  string
		expectedEventsCountPreUpdate       int
		expectedEventsAttributesPreUpdate  []sdk.Attribute
		expectedEventsCountPostUpdate      int
		expectedEventsAttributesPostUpdate []sdk.Attribute
	}{
		{
			name: "Test demand order fulfillment - success",
			rollappPacket: commontypes.RollappPacket{
				RollappId: "testRollappID",
				Packet: &channeltypes.Packet{
					SourcePort:         "testSourcePort",
					SourceChannel:      "testSourceChannel",
					DestinationPort:    "testDestinationPort",
					DestinationChannel: "testDestinationChannel",
					Data:               []byte("testData"),
					Sequence:           1,
				},
				Status:      commontypes.Status_PENDING,
				ProofHeight: 1,
			},
			rollappUpdatedStatus:         commontypes.Status_FINALIZED,
			rollappUpdateError:           types.ErrRollappPacketDoesNotExist,
			expectedEventType:            delayedAckEventType,
			expectedEventsCountPreUpdate: 1,
			expectedEventsAttributesPreUpdate: []sdk.Attribute{
				sdk.NewAttribute(commontypes.AttributeKeyRollappId, "testRollappID"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
				sdk.NewAttribute(commontypes.AttributeKeyPacketSourcePort, "testSourcePort"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketSourceChannel, "testSourceChannel"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketDestinationPort, "testDestinationPort"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketSequence, "1"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketError, ""),
			},
			expectedEventsCountPostUpdate: 2,
			expectedEventsAttributesPostUpdate: []sdk.Attribute{
				sdk.NewAttribute(commontypes.AttributeKeyRollappId, "testRollappID"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketStatus, commontypes.Status_FINALIZED.String()),
				sdk.NewAttribute(commontypes.AttributeKeyPacketSourcePort, "testSourcePort"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketSourceChannel, "testSourceChannel"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketDestinationPort, "testDestinationPort"),
				sdk.NewAttribute(commontypes.AttributeKeyPacketSequence, "1"),
			},
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Set the rpllapp packet
			keeper.SetRollappPacket(ctx, tc.rollappPacket)
			// Check the events
			suite.AssertEventEmitted(ctx, tc.expectedEventType, tc.expectedEventsCountPreUpdate)
			lastEvent, ok := suite.FindLastEventOfType(ctx.EventManager().Events(), tc.expectedEventType)
			suite.Require().True(ok)
			suite.AssertAttributes(lastEvent, tc.expectedEventsAttributesPreUpdate)
			// Update the rollapp packet
			tc.rollappPacket.Error = tc.rollappUpdateError.Error()
			_, err := keeper.UpdateRollappPacketWithStatus(ctx, tc.rollappPacket, tc.rollappUpdatedStatus)
			suite.Require().NoError(err)
			// Check the events
			suite.AssertEventEmitted(ctx, tc.expectedEventType, tc.expectedEventsCountPostUpdate)
			lastEvent, ok = suite.FindLastEventOfType(ctx.EventManager().Events(), tc.expectedEventType)
			suite.Require().True(ok)
			suite.AssertAttributes(lastEvent, tc.expectedEventsAttributesPostUpdate)
		})
	}
}

func (suite *DelayedAckTestSuite) TestListRollappPackets() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	rollappIDs := []string{"testRollappID1", "testRollappID2", "testRollappID3"}

	sm := map[int]commontypes.Status{
		0: commontypes.Status_PENDING,
		1: commontypes.Status_FINALIZED,
		2: commontypes.Status_REVERTED,
	}

	var packetsToSet []commontypes.RollappPacket
	// Create and set some RollappPackets
	for _, rollappID := range rollappIDs {
		for i := 1; i < 6; i++ {
			packet := commontypes.RollappPacket{
				RollappId: rollappID,
				Packet: &channeltypes.Packet{
					SourcePort:         "testSourcePort",
					SourceChannel:      "testSourceChannel",
					DestinationPort:    "testDestinationPort",
					DestinationChannel: "testDestinationChannel",
					Data:               []byte("testData"),
					Sequence:           uint64(i),
				},
				Status:      sm[i%3],
				Type:        commontypes.RollappPacket_Type(i % 3),
				ProofHeight: uint64(6 - i),
			}
			packetsToSet = append(packetsToSet, packet)
		}
	}
	totalLength := len(packetsToSet)

	for _, packet := range packetsToSet {
		keeper.SetRollappPacket(ctx, packet)
	}

	// Get all rollapp packets by rollapp id
	packets := keeper.ListRollappPackets(ctx, types.ByRollappID(rollappIDs[0]))
	suite.Require().Equal(5, len(packets))

	expectPendingLength := 3
	pendingPackets := keeper.ListRollappPackets(ctx, types.ByStatus(commontypes.Status_PENDING))
	suite.Require().Equal(expectPendingLength, len(pendingPackets))

	expectFinalizedLength := 6
	finalizedPackets := keeper.ListRollappPackets(ctx, types.ByStatus(commontypes.Status_FINALIZED))
	suite.Require().Equal(expectFinalizedLength, len(finalizedPackets))

	expectRevertedLength := 6
	revertedPackets := keeper.ListRollappPackets(ctx, types.ByStatus(commontypes.Status_REVERTED))
	suite.Require().Equal(expectRevertedLength, len(revertedPackets))

	expectRevertedLengthLimit := 4
	revertedPacketsLimit := keeper.ListRollappPackets(ctx, types.ByStatus(commontypes.Status_REVERTED).Take(4))
	suite.Require().Equal(expectRevertedLengthLimit, len(revertedPacketsLimit))

	suite.Require().Equal(totalLength, len(pendingPackets)+len(finalizedPackets)+len(revertedPackets))

	rollappPacket1Finalized := keeper.ListRollappPackets(ctx, types.ByRollappIDByStatus(rollappIDs[0], commontypes.Status_FINALIZED))
	rollappPacket2Pending := keeper.ListRollappPackets(ctx, types.ByRollappIDByStatus(rollappIDs[1], commontypes.Status_PENDING))
	rollappPacket3Reverted := keeper.ListRollappPackets(ctx, types.ByRollappIDByStatus(rollappIDs[2], commontypes.Status_REVERTED))
	suite.Require().Equal(2, len(rollappPacket1Finalized))
	suite.Require().Equal(1, len(rollappPacket2Pending))
	suite.Require().Equal(2, len(rollappPacket3Reverted))

	rollappPacket1MaxHeight4 := keeper.ListRollappPackets(ctx, types.PendingByRollappIDByMaxHeight(rollappIDs[0], 4))
	suite.Require().Equal(1, len(rollappPacket1MaxHeight4))

	rollappPacket2MaxHeight3 := keeper.ListRollappPackets(ctx, types.PendingByRollappIDByMaxHeight(rollappIDs[1], 3))
	suite.Require().Equal(1, len(rollappPacket2MaxHeight3))

	expectOnRecvLength := 3
	onRecvPackets := keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_RECV, commontypes.Status_PENDING))
	suite.Equal(expectOnRecvLength, len(onRecvPackets))

	expectOnAckLength := 6
	onAckPackets := keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_ACK, commontypes.Status_FINALIZED))
	suite.Equal(expectOnAckLength, len(onAckPackets))

	expectOnTimeoutLength := 6
	onTimeoutPackets := keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_TIMEOUT, commontypes.Status_REVERTED))
	suite.Equal(expectOnTimeoutLength, len(onTimeoutPackets))

	suite.Require().Equal(totalLength, len(onRecvPackets)+len(onAckPackets)+len(onTimeoutPackets))
}

func (suite *DelayedAckTestSuite) TestUpdateRollappPacketWithStatus() {
	var err error
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	packet := commontypes.RollappPacket{
		RollappId: "testRollappID",
		Packet: &channeltypes.Packet{
			SourcePort:         "testSourcePort",
			SourceChannel:      "testSourceChannel",
			DestinationPort:    "testDestinationPort",
			DestinationChannel: "testDestinationChannel",
			Data:               []byte("testData"),
			Sequence:           1,
		},
		Status:      commontypes.Status_PENDING,
		ProofHeight: 1,
	}
	keeper.SetRollappPacket(ctx, packet)
	// Update the packet status
	packet, err = keeper.UpdateRollappPacketWithStatus(ctx, packet, commontypes.Status_FINALIZED)
	suite.Require().NoError(err)
	suite.Require().Equal(commontypes.Status_FINALIZED, packet.Status)
	packets := keeper.GetAllRollappPackets(ctx)
	suite.Require().Equal(1, len(packets))
	// Set the packet and make sure there is only one packet in the store
	keeper.SetRollappPacket(ctx, packet)
	packets = keeper.GetAllRollappPackets(ctx)
	suite.Require().Equal(1, len(packets))
}

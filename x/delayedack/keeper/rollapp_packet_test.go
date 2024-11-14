package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func (suite *DelayedAckTestSuite) TestRollappPacketEvents() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	tests := []struct {
		name                               string
		rollappPacket                      commontypes.RollappPacket
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
				RollappId:   "testRollappID",
				Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
				Status:      commontypes.Status_PENDING,
				ProofHeight: 1,
			},
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
			_, err := keeper.UpdateRollappPacketAfterFinalization(ctx, tc.rollappPacket)
			suite.Require().NoError(err)
			// Check the events
			suite.AssertEventEmitted(ctx, tc.expectedEventType, tc.expectedEventsCountPostUpdate)
			lastEvent, ok = suite.FindLastEventOfType(ctx.EventManager().Events(), tc.expectedEventType)
			suite.Require().True(ok)
			suite.AssertAttributes(lastEvent, tc.expectedEventsAttributesPostUpdate)
		})
	}
}

// TestListRollappPackets tests the ListRollappPackets function
// we have 3 rollapps
// 2 pending packets, 3 finalized packets
// 2 onRecv packets, 2 onAck packets, 1 onTimeout packets
func (suite *DelayedAckTestSuite) TestListRollappPackets() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	rollappIDs := []string{"testRollappID1", "testRollappID2", "testRollappID3"}

	sm := map[int]commontypes.Status{
		0: commontypes.Status_PENDING,
		1: commontypes.Status_FINALIZED,
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
				Status:      sm[i%2],
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

	expectPendingLength := 6
	pendingPackets := keeper.ListRollappPackets(ctx, types.ByStatus(commontypes.Status_PENDING))
	suite.Require().Equal(expectPendingLength, len(pendingPackets))

	expectFinalizedLength := 9
	finalizedPackets := keeper.ListRollappPackets(ctx, types.ByStatus(commontypes.Status_FINALIZED))
	suite.Require().Equal(expectFinalizedLength, len(finalizedPackets))

	expectFinalizedLengthLimit := 4
	finalizedPacketsLimit := keeper.ListRollappPackets(ctx, types.ByStatus(commontypes.Status_FINALIZED).Take(4))
	suite.Require().Equal(expectFinalizedLengthLimit, len(finalizedPacketsLimit))

	suite.Require().Equal(totalLength, len(pendingPackets)+len(finalizedPackets))

	rollappPacket1Finalized := keeper.ListRollappPackets(ctx, types.ByRollappIDByStatus(rollappIDs[0], commontypes.Status_FINALIZED))
	rollappPacket2Pending := keeper.ListRollappPackets(ctx, types.ByRollappIDByStatus(rollappIDs[1], commontypes.Status_PENDING))
	suite.Require().Equal(3, len(rollappPacket1Finalized))
	suite.Require().Equal(2, len(rollappPacket2Pending))

	rollappPacket1MaxHeight4 := keeper.ListRollappPackets(ctx, types.PendingByRollappIDByMaxHeight(rollappIDs[0], 4))
	suite.Require().Equal(2, len(rollappPacket1MaxHeight4))

	rollappPacket2MaxHeight3 := keeper.ListRollappPackets(ctx, types.PendingByRollappIDByMaxHeight(rollappIDs[1], 3))
	suite.Require().Equal(1, len(rollappPacket2MaxHeight3))

	expectOnRecvLength := 0 // i % 2 == 0 AND i % 3 == 0
	onRecvPackets := keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_RECV, commontypes.Status_PENDING))
	suite.Equal(expectOnRecvLength, len(onRecvPackets))

	expectOnAckLength := 3 // i % 2 == 1 AND i % 3 == 1 (per rollapp)
	onAckPackets := keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_ACK, commontypes.Status_FINALIZED))
	suite.Equal(expectOnAckLength, len(onAckPackets))

	expectOnTimeoutLength := 3 // i % 2 == 1 AND i % 3 == 2 (per rollapp)
	onTimeoutPackets := keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_TIMEOUT, commontypes.Status_FINALIZED))
	suite.Equal(expectOnTimeoutLength, len(onTimeoutPackets))

	var totalCount int
	for _, status := range sm {
		onRecvPackets = keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_RECV, status))
		onAckPackets = keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_ACK, status))
		onTimeoutPackets = keeper.ListRollappPackets(ctx, types.ByTypeByStatus(commontypes.RollappPacket_ON_TIMEOUT, status))
		totalCount += len(onRecvPackets) + len(onAckPackets) + len(onTimeoutPackets)
	}

	suite.Require().Equal(totalLength, totalCount)
}

func (suite *DelayedAckTestSuite) TestUpdateRollappPacketWithStatus_PendingToFinalized() {
	var err error
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	oldPacket := commontypes.RollappPacket{
		RollappId:   "testRollappID",
		Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
		Status:      commontypes.Status_PENDING,
		ProofHeight: 1,
	}
	keeper.SetRollappPacket(ctx, oldPacket)
	err = keeper.SetPendingPacketByAddress(ctx, apptesting.TestPacketReceiver, oldPacket.RollappPacketKey())
	suite.Require().NoError(err)

	// Update the packet status
	packet, err := keeper.UpdateRollappPacketAfterFinalization(ctx, oldPacket)
	suite.Require().NoError(err)
	suite.Require().Equal(commontypes.Status_FINALIZED, packet.Status)

	// Check the updated packet is not in the receiver's index anymore
	byReceiver, err := keeper.GetPendingPacketsByAddress(ctx, apptesting.TestPacketReceiver)
	suite.Require().NoError(err)
	suite.Require().Empty(len(byReceiver))

	packets := keeper.GetAllRollappPackets(ctx)
	suite.Require().Equal(1, len(packets))
	// Set the packet and make sure there is only one packet in the store
	keeper.SetRollappPacket(ctx, packet)
	packets = keeper.GetAllRollappPackets(ctx)
	suite.Require().Equal(1, len(packets))
}

func (suite *DelayedAckTestSuite) TestUpdateRollappPacketTransferAddress_ON_RECV() {
	var err error
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	packet := commontypes.RollappPacket{
		RollappId:   "testRollappID",
		Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
		Type:        commontypes.RollappPacket_ON_RECV,
		Status:      commontypes.Status_PENDING,
		ProofHeight: 1,
	}
	keeper.SetRollappPacket(ctx, packet)
	err = keeper.SetPendingPacketByAddress(ctx, apptesting.TestPacketReceiver, packet.RollappPacketKey())
	suite.Require().NoError(err)

	// Update the packet receiver
	const newReceiver = "newReceiver"
	err = keeper.UpdateRollappPacketTransferAddress(ctx, string(packet.RollappPacketKey()), newReceiver)
	suite.Require().NoError(err)

	// Check the state
	packets := keeper.GetAllRollappPackets(ctx)
	suite.Require().Equal(1, len(packets))
	pd1, err := packets[0].GetTransferPacketData()
	suite.Require().NoError(err)
	suite.Require().Equal(newReceiver, pd1.Receiver)

	// Check the packet key is the same
	actualPacket, err := keeper.GetRollappPacket(ctx, string(packet.RollappPacketKey()))
	suite.Require().NoError(err)
	pd2, err := actualPacket.GetTransferPacketData()
	suite.Require().NoError(err)
	suite.Require().Equal(newReceiver, pd2.Receiver)

	// Check the index
	// Check the packet is in the receiver's index
	byReceiverNew, err := keeper.GetPendingPacketsByAddress(ctx, newReceiver)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(byReceiverNew))
	suite.Require().Equal(packet.RollappPacketKey(), byReceiverNew[0].RollappPacketKey())
	pd3, err := byReceiverNew[0].GetTransferPacketData()
	suite.Require().NoError(err)
	suite.Require().Equal(newReceiver, pd3.Receiver)

	// Check the packet is not in the receiver's index
	byReceiverOld, err := keeper.GetPendingPacketsByAddress(ctx, apptesting.TestPacketReceiver)
	suite.Require().NoError(err)
	suite.Require().Empty(byReceiverOld)

	// Set the packet and make sure there is only one packet in the store
	keeper.SetRollappPacket(ctx, packet)
	packets = keeper.GetAllRollappPackets(ctx)
	suite.Require().Equal(1, len(packets))
}

func (suite *DelayedAckTestSuite) TestUpdateRollappPacketTransferAddress_ON_ACK() {
	var err error
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	packet := commontypes.RollappPacket{
		RollappId:   "testRollappID",
		Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
		Type:        commontypes.RollappPacket_ON_ACK,
		Status:      commontypes.Status_PENDING,
		ProofHeight: 1,
	}
	keeper.SetRollappPacket(ctx, packet)
	err = keeper.SetPendingPacketByAddress(ctx, apptesting.TestPacketSender, packet.RollappPacketKey())
	suite.Require().NoError(err)

	// Update the packet receiver
	const newSender = "newSender"
	err = keeper.UpdateRollappPacketTransferAddress(ctx, string(packet.RollappPacketKey()), newSender)
	suite.Require().NoError(err)

	// Check the state
	packets := keeper.GetAllRollappPackets(ctx)
	suite.Require().Equal(1, len(packets))
	pd1, err := packets[0].GetTransferPacketData()
	suite.Require().NoError(err)
	suite.Require().Equal(newSender, pd1.Sender)

	// Check the packet key is the same
	actualPacket, err := keeper.GetRollappPacket(ctx, string(packet.RollappPacketKey()))
	suite.Require().NoError(err)
	pd2, err := actualPacket.GetTransferPacketData()
	suite.Require().NoError(err)
	suite.Require().Equal(newSender, pd2.Sender)

	// Check the index
	// Check the new packet is in the sender's index
	bySenderNew, err := keeper.GetPendingPacketsByAddress(ctx, newSender)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(bySenderNew))
	suite.Require().Equal(packet.RollappPacketKey(), bySenderNew[0].RollappPacketKey())
	pd3, err := bySenderNew[0].GetTransferPacketData()
	suite.Require().NoError(err)
	suite.Require().Equal(newSender, pd3.Sender)

	// Check the old packet is not in the sender's index
	bySenderOld, err := keeper.GetPendingPacketsByAddress(ctx, apptesting.TestPacketSender)
	suite.Require().NoError(err)
	suite.Require().Empty(bySenderOld)

	// Set the packet and make sure there is only one packet in the store
	keeper.SetRollappPacket(ctx, packet)
	packets = keeper.GetAllRollappPackets(ctx)
	suite.Require().Equal(1, len(packets))
}

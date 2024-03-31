package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	dkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
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
			err := keeper.SetRollappPacket(ctx, tc.rollappPacket)
			suite.Require().NoError(err)
			// Check the events
			suite.AssertEventEmitted(ctx, tc.expectedEventType, tc.expectedEventsCountPreUpdate)
			lastEvent, ok := suite.FindLastEventOfType(ctx.EventManager().Events(), tc.expectedEventType)
			suite.Require().True(ok)
			suite.AssertAttributes(lastEvent, tc.expectedEventsAttributesPreUpdate)
			// Update the rollapp packet
			tc.rollappPacket.Error = tc.rollappUpdateError.Error()
			_, err = keeper.UpdateRollappPacketWithStatus(ctx, tc.rollappPacket, tc.rollappUpdatedStatus)
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
	rollappID := "testRollappID"

	// Create and set some RollappPackets
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
			Status: commontypes.Status_PENDING,
			// Intentionally set the proof height in descending order to test sorting
			ProofHeight: uint64(6 - i),
		}
		err := keeper.SetRollappPacket(ctx, packet)
		suite.Require().NoError(err)
	}

	// Get all rollapp packets by rollapp id
	packets := keeper.ListRollappPackets(ctx, dkeeper.ByRollappID(rollappID))
	suite.Require().Equal(5, len(packets))

	// Get the packets until height 4
	packets = keeper.ListRollappPackets(ctx, dkeeper.ByRollappIDAndStatusAndMaxHeight(rollappID, 4, true, commontypes.Status_PENDING))
	suite.Require().Equal(4, len(packets))

	// Update the packet status to finalized
	for _, packet := range packets {
		_, err := keeper.UpdateRollappPacketWithStatus(ctx, packet, commontypes.Status_FINALIZED)
		suite.Require().NoError(err)
	}

	finalizedPackets := keeper.ListRollappPackets(ctx, dkeeper.ByStatus(commontypes.Status_FINALIZED))
	suite.Require().Equal(4, len(finalizedPackets))

	// Get the packets until height 5
	packets = keeper.ListRollappPackets(ctx, dkeeper.ByRollappIDAndStatusAndMaxHeight(rollappID, 5, true, commontypes.Status_PENDING))
	suite.Require().Equal(1, len(packets))
}

func (suite *DelayedAckTestSuite) TestUpdateRollappPacketWithStatus() {
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
	err := keeper.SetRollappPacket(ctx, packet)
	suite.Require().NoError(err)
	// Update the packet status
	packet, err = keeper.UpdateRollappPacketWithStatus(ctx, packet, commontypes.Status_FINALIZED)
	suite.Require().NoError(err)
	suite.Require().Equal(commontypes.Status_FINALIZED, packet.Status)
	packets := keeper.ListRollappPackets(ctx, dkeeper.AllRollappPackets())
	suite.Require().Equal(1, len(packets))
	// Set the packet and make sure there is only one packet in the store
	err = keeper.SetRollappPacket(ctx, packet)
	suite.Require().NoError(err)
	packets = keeper.ListRollappPackets(ctx, dkeeper.AllRollappPackets())
	suite.Require().Equal(1, len(packets))
}

package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (s *DelayedAckTestSuite) TestFinalizePacket() {
	rollapp := "rollapp_1234-1"

	cases := []struct {
		name          string
		packet        commontypes.RollappPacket
		rollappHeight uint64
		expectErr     bool
		errContains   string
	}{
		{
			name: "success",
			packet: commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 8,
				Packet:      apptesting.GenerateTestPacket(s.T(), 2),
			},
			rollappHeight: 10,
			expectErr:     false,
			errContains:   "",
		},
		{
			name: "error: packet not found: packet status is not pending",
			packet: commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 8,
				Packet:      apptesting.GenerateTestPacket(s.T(), 2),
			},
			rollappHeight: 10,
			expectErr:     true,
			errContains:   "get rollapp packet:",
		},
		{
			name: "error: packet proof height is greater than rollapp height",
			packet: commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 15,
				Packet:      apptesting.GenerateTestPacket(s.T(), 2),
			},
			rollappHeight: 10,
			expectErr:     true,
			errContains:   "verify height",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()

			s.CreateRollappByName(rollapp)
			proposer := s.CreateDefaultSequencer(s.Ctx, rollapp)

			// create state info
			stateInfo := rollapptypes.StateInfo{
				StateInfoIndex: rollapptypes.StateInfoIndex{
					RollappId: rollapp,
					Index:     1,
				},
				StartHeight: 1,
				NumBlocks:   tc.rollappHeight,
				Status:      commontypes.Status_FINALIZED,
				Sequencer:   proposer,
			}

			// save state info
			s.App.RollappKeeper.SetStateInfo(s.Ctx, stateInfo)
			s.App.RollappKeeper.SetLatestFinalizedStateIndex(s.Ctx, rollapptypes.StateInfoIndex{
				RollappId: rollapp,
				Index:     stateInfo.GetIndex().Index,
			})

			// save rollapp packets
			s.App.DelayedAckKeeper.SetRollappPacket(s.Ctx, tc.packet)

			// try to finalize a packet
			handler := s.App.MsgServiceRouter().Handler(new(types.MsgFinalizePacket))
			resp, err := handler(s.Ctx, &types.MsgFinalizePacket{
				Sender:            apptesting.CreateRandomAccounts(1)[0].String(),
				RollappId:         tc.packet.RollappId,
				PacketProofHeight: tc.packet.ProofHeight,
				PacketType:        tc.packet.Type,
				PacketSrcChannel:  tc.packet.Packet.SourceChannel,
				PacketSequence:    tc.packet.Packet.Sequence,
			})

			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)

				// verify that the corresponding finalized packet exists
				packetKey := commontypes.RollappPacketKey(
					commontypes.Status_FINALIZED,
					tc.packet.RollappId,
					tc.packet.ProofHeight,
					tc.packet.Type,
					tc.packet.Packet.SourceChannel,
					tc.packet.Packet.Sequence,
				)
				_, err = s.App.DelayedAckKeeper.GetRollappPacket(s.Ctx, string(packetKey))
				s.Require().NoError(err)
			}
		})
	}
}

func (s *DelayedAckTestSuite) TestFinalizePacketByPacketKey() {
	rollapp := "rollapp_1234-1"

	cases := []struct {
		name                 string
		packet               commontypes.RollappPacket
		rollappHeight        uint64
		expectedPacketStatus commontypes.Status
		expectFinalize       bool
	}{
		{
			name: "packet is finalized",
			packet: commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 8,
				Packet:      apptesting.GenerateTestPacket(s.T(), 2),
			},
			rollappHeight:        10,
			expectedPacketStatus: commontypes.Status_FINALIZED,
			expectFinalize:       true,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()

			s.CreateRollappByName(rollapp)
			proposer := s.CreateDefaultSequencer(s.Ctx, rollapp)

			// create state info
			stateInfo := rollapptypes.StateInfo{
				StateInfoIndex: rollapptypes.StateInfoIndex{
					RollappId: rollapp,
					Index:     1,
				},
				StartHeight: 1,
				NumBlocks:   tc.rollappHeight,
				Status:      commontypes.Status_FINALIZED,
				Sequencer:   proposer,
			}

			// save state info
			s.App.RollappKeeper.SetStateInfo(s.Ctx, stateInfo)
			s.App.RollappKeeper.SetLatestFinalizedStateIndex(s.Ctx, rollapptypes.StateInfoIndex{
				RollappId: rollapp,
				Index:     stateInfo.GetIndex().Index,
			})

			s.App.DelayedAckKeeper.SetRollappPacket(s.Ctx, tc.packet)

			// try to finalize a packet
			handler := s.App.MsgServiceRouter().Handler(new(types.MsgFinalizePacketByPacketKey))
			resp, err := handler(s.Ctx, &types.MsgFinalizePacketByPacketKey{
				Sender:    apptesting.CreateRandomAccounts(1)[0].String(),
				PacketKey: commontypes.EncodePacketKey(tc.packet.RollappPacketKey()),
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp)

			// verify that the corresponding finalized packet exists
			packetKey := commontypes.RollappPacketKey(
				tc.expectedPacketStatus,
				tc.packet.RollappId,
				tc.packet.ProofHeight,
				tc.packet.Type,
				tc.packet.Packet.SourceChannel,
				tc.packet.Packet.Sequence,
			)
			_, err = s.App.DelayedAckKeeper.GetRollappPacket(s.Ctx, string(packetKey))
			s.Require().NoError(err)
		})
	}
}

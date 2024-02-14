package keeper_test

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/stretchr/testify/require"
)

func TestListRollappPacketsForRollappAtHeight(t *testing.T) {
	keeper, ctx := keepertest.DelayedackKeeper(t)
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
			Status:      commontypes.Status_PENDING,
			ProofHeight: uint64(i * 2),
		}
		err := keeper.SetRollappPacket(ctx, packet)
		require.NoError(t, err)
	}

	// Get all rollapp packets
	packets := keeper.GetAllRollappPackets(ctx)
	require.Equal(t, 5, len(packets))

	// Get the packets until height 6
	packets = keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_PENDING, 6)
	require.Equal(t, 3, len(packets))

	// Update the packet status to finalized
	for _, packet := range packets {
		_, err := keeper.UpdateRollappPacketWithStatus(ctx, packet, commontypes.Status_FINALIZED)
		require.NoError(t, err)
	}
	finalizedPackets := keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_FINALIZED, 0)
	require.Equal(t, 3, len(finalizedPackets))

	// Get the packets until height 14
	packets = keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_PENDING, 14)
	require.Equal(t, 2, len(packets))
}

func TestUpdateRollappPacketWithStatus(t *testing.T) {
	keeper, ctx := keepertest.DelayedackKeeper(t)
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
	require.NoError(t, err)
	// Update the packet status
	packet, err = keeper.UpdateRollappPacketWithStatus(ctx, packet, commontypes.Status_FINALIZED)
	require.NoError(t, err)
	packets := keeper.GetAllRollappPackets(ctx)
	require.Equal(t, commontypes.Status_FINALIZED, packet.Status)
	require.Equal(t, 1, len(packets))
	// Set the packet and make sure there is only one packet in the store
	err = keeper.SetRollappPacket(ctx, packet)
	require.NoError(t, err)
	packets = keeper.GetAllRollappPackets(ctx)
	require.Equal(t, 1, len(packets))
}

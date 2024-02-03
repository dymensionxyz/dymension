package keeper_test

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/stretchr/testify/require"
)

func TestListRollappPacketsForRollappAtHeight(t *testing.T) {
	keeper, ctx := keepertest.DelayedackKeeper(t)
	rollappID := "testRollappID"

	// Create and set some RollappPackets
	for i := 1; i < 6; i++ {
		packet := types.RollappPacket{
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
		keeper.SetRollappPacket(ctx, rollappID, packet)
	}

	// Get the packets until height 6
	packets := keeper.ListRollappPendingPackets(ctx, rollappID, 6)
	require.Equal(t, 3, len(packets))

	// Update the packet status to finalized
	for _, packet := range packets {
		keeper.UpdateRollappPacketWithStatus(ctx, rollappID, packet, commontypes.Status_FINALIZED)
	}

	// Get the packets until height 14
	packets = keeper.ListRollappPendingPackets(ctx, rollappID, 14)
	require.Equal(t, 2, len(packets))
}

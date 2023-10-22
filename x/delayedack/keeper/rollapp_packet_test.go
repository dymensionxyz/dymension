package keeper_test

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/x/delayedack/types"
	"github.com/stretchr/testify/require"
)

func TestListRollappPacketsForRollappAtHeight(t *testing.T) {
	keeper, ctx := keepertest.DelayedackKeeper(t)
	rollappID := "testRollappID"

	// Create and set some RollappPackets
	for i := 0; i < 5; i++ {
		packet := types.RollappPacket{
			Packet: &channeltypes.Packet{
				SourcePort:         "testSourcePort",
				SourceChannel:      "testSourceChannel",
				DestinationPort:    "testDestinationPort",
				DestinationChannel: "testDestinationChannel",
				Data:               []byte("testData"),
				Sequence:           uint64(i),
			},
			Status: types.RollappPacket_PENDING,
		}
		// Set the context BlockHeight
		ctx = ctx.WithBlockHeight(int64(i%2) + 1) // This should create 3 packets with height 1 and 2 packets with sequence 2
		keeper.SetRollappPacket(ctx, rollappID, packet)
	}

	// Get the packets with height 1
	packets := keeper.ListRollappPacketsForRollappAtHeight(ctx, rollappID, 1)
	require.Equal(t, 3, len(packets))

	// Get the packets with height 2
	packets = keeper.ListRollappPacketsForRollappAtHeight(ctx, rollappID, 2)
	require.Equal(t, 2, len(packets))
}

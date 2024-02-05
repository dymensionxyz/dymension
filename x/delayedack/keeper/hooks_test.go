package keeper_test

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/stretchr/testify/require"
)

// TestAfterEpochEnd tests that the finalized of rollapp packets
// are deleted given the correct epoch identifier
func TestAfterEpochEnd(t *testing.T) {
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
			name:                 "epoch identifier matches params set",
			pendingPacketsNum:    5,
			finalizePacketsNum:   3,
			epochIdentifierParam: "minute",
			epochIdentifier:      "minute",
			expectedDeleted:      3,
			expectedTotal:        2,
		},
		{
			name:                 "epoch identifer does not match params set",
			pendingPacketsNum:    5,
			finalizePacketsNum:   3,
			epochIdentifierParam: "minute",
			epochIdentifier:      "hour",
			expectedDeleted:      0,
			expectedTotal:        5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := keepertest.DelayedackKeeper(t)
			for i := 1; i <= tc.pendingPacketsNum; i++ {
				rollappPacket := &types.RollappPacket{
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
			require.Equal(t, tc.pendingPacketsNum, len(rollappPackets))

			for _, rollappPacket := range rollappPackets[:tc.finalizePacketsNum] {
				keeper.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_FINALIZED)
			}
			finalizedRollappPackets := keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_FINALIZED, 0)
			require.Equal(t, tc.finalizePacketsNum, len(finalizedRollappPackets))

			keeper.SetParams(ctx, types.Params{EpochIdentifier: tc.epochIdentifierParam})
			epochHooks := keeper.GetEpochHooks()
			epochHooks.AfterEpochEnd(ctx, tc.epochIdentifier, 1)

			finalizedRollappPackets = keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_FINALIZED, 0)
			require.Equal(t, tc.finalizePacketsNum-tc.expectedDeleted, len(finalizedRollappPackets))

			totalRollappPackets := len(finalizedRollappPackets) + len(keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_PENDING, 0))
			require.Equal(t, tc.expectedTotal, totalRollappPackets)
		})
	}
}

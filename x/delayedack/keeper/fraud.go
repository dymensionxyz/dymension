package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func (k Keeper) HandleHardFork(ctx sdk.Context, rollappID string, height uint64, ibc porttypes.IBCModule) error {
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")

	// Get all the pending packets from fork height inclusive
	rollappPendingPackets := k.ListRollappPackets(ctx, types.PendingByRollappIDFromHeight(rollappID, height))

	// Iterate over all the pending packets and revert them
	for _, rollappPacket := range rollappPendingPackets {
		logContext := []interface{}{
			"rollappID", rollappID,
			"sourceChannel", rollappPacket.Packet.SourceChannel,
			"destChannel", rollappPacket.Packet.DestinationChannel,
			"type", rollappPacket.Type,
			"sequence", rollappPacket.Packet.Sequence,
		}

		// refund all pending outgoing packets
		if rollappPacket.Type == commontypes.RollappPacket_ON_ACK || rollappPacket.Type == commontypes.RollappPacket_ON_TIMEOUT {
			// we don't have access directly to `refundPacketToken` function, so we'll use the `OnTimeoutPacket` function
			err := ibc.OnTimeoutPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
			if err != nil {
				logger.Error("failed to refund reverted packet", append(logContext, "error", err.Error())...)
			}
		} else {
			// for incoming packets, we need to reset the packet receipt
			ibcPacket := rollappPacket.Packet
			k.channelKeeper.SetPacketReceipt(ctx, ibcPacket.GetDestPort(), ibcPacket.GetDestChannel(), ibcPacket.GetSequence())
		}

		// delete the packet
		k.DeleteRollappPacket(ctx, &rollappPacket)
		logger.Debug("reverted IBC rollapp packet", logContext...)
	}

	logger.Info("reverting IBC rollapp packets", "rollappID", rollappID, "numPackets", len(rollappPendingPackets))

	return nil
}

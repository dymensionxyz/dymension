package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func (k Keeper) HandleFraud(ctx sdk.Context, rollappID string, ibc porttypes.IBCModule) error {
	// Get all the pending packets
	rollappPendingPackets := k.ListRollappPackets(ctx, types.ByRollappIDByStatus(rollappID, commontypes.Status_PENDING))
	if len(rollappPendingPackets) == 0 {
		return nil
	}
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	logger.Info("Reverting IBC rollapp packets.", "rollappID", rollappID)

	// Iterate over all the pending packets and revert them
	for _, rollappPacket := range rollappPendingPackets {
		logContext := []interface{}{
			"rollappID", rollappID,
			"sourceChannel", rollappPacket.Packet.SourceChannel,
			"destChannel", rollappPacket.Packet.DestinationChannel,
			"type", rollappPacket.Type,
			"sequence", rollappPacket.Packet.Sequence,
		}

		// these are outgoing transfers Hub->RA
		if rollappPacket.Type == commontypes.RollappPacket_ON_ACK || rollappPacket.Type == commontypes.RollappPacket_ON_TIMEOUT {
			// refund all pending outgoing packets
			// we don't have access directly to `refundPacketToken` function, so we'll use the `OnTimeoutPacket` function
			err := ibc.OnTimeoutPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
			if err != nil {
				logger.Error("Refund reverted packet.", append(logContext, "error", err.Error())...)
			}
		}
		// Update status to reverted
		_, err := k.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_REVERTED)
		if err != nil {
			logger.Error("Reverting IBC rollapp packet", append(logContext, "error", err.Error())...)
			return err
		}

		logger.Debug("reverted IBC rollapp packet", logContext...)
	}
	return nil
}

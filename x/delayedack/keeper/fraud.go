package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func (k Keeper) HandleFraud(ctx sdk.Context, rollappID string) error {
	// Get all the pending packets
	rollappPendingPackets := k.ListRollappPackets(ctx, ByRollappIDByStatus(rollappID, commontypes.Status_PENDING))
	if len(rollappPendingPackets) == 0 {
		return nil
	}
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	logger.Debug("Reverting IBC rollapp packets", "rollappID", rollappID)
	for _, rollappPacket := range rollappPendingPackets {

		// setup log context
		logContext := []interface{}{
			"rollappID", rollappID,
			"type", rollappPacket.Type,
			"source channel", rollappPacket.Packet.SourceChannel,
			"sequence", rollappPacket.Packet.Sequence,
		}

		logger.Debug("Reverting IBC rollapp packet", logContext...)

		if rollappPacket.Type == commontypes.RollappPacket_ON_RECV {
			err := k.writeFailedAck(ctx, rollappPacket, commontypes.Status_REVERTED.String())
			if err != nil {
				logger.Error("failed to write failed ack", append(logContext, "error", err.Error())...)
				// don't return here as it's nice to have
			}
		}
		// Update status to reverted
		_, err := k.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_REVERTED)
		if err != nil {
			logger.Error("Error reverting IBC rollapp packet", append(logContext, "error", err.Error())...)
			return err
		}
	}
	return nil
}

func (k Keeper) writeFailedAck(ctx sdk.Context, rollappPacket commontypes.RollappPacket, msg string) error {
	failedAck := channeltypes.NewErrorAcknowledgement(fmt.Errorf(msg))
	// Write the acknowledgement to the chain
	_, chanCap, err := k.LookupModuleByChannel(ctx, rollappPacket.Packet.DestinationPort, rollappPacket.Packet.DestinationChannel)
	if err != nil {
		return err
	}
	err = k.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, failedAck)
	if err != nil {
		return err
	}

	return nil
}

package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

// AfterStateFinalized implements the RollappHooks interface
func (im IBCMiddleware) AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *rollapptypes.StateInfo) error {
	// Finalize the packets for the rollapp at the given height
	stateEndHeight := stateInfo.StartHeight + stateInfo.NumBlocks - 1
	im.FinalizeRollappPackets(ctx, rollappID, stateEndHeight)
	return nil
}

// FinalizeRollappPackets finalizes the packets for the given rollapp until the given height which is
// the end height of the latest finalized state
func (im IBCMiddleware) FinalizeRollappPackets(ctx sdk.Context, rollappID string, stateEndHeight uint64) {
	rollappPendingPackets := im.keeper.ListRollappPendingPackets(ctx, rollappID, stateEndHeight)
	if len(rollappPendingPackets) == 0 {
		return
	}

	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	// Get the packets for the rollapp until height
	logger.Debug("Finalizing IBC rollapp packets", "rollappID", rollappID, "state end height", stateEndHeight, "num packets", len(rollappPendingPackets))
	for _, rollappPacket := range rollappPendingPackets {
		logger.Debug("Finalizing IBC rollapp packet", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel())
		// Update the packet status
		im.keeper.UpdateRollappPacketStatus(ctx, rollappID, rollappPacket, types.RollappPacket_ACCEPTED)
		// Call the onRecvPacket callback for each packet
		ack := im.app.OnRecvPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
		// Write the acknowledgement to the chain only if it is synchronous
		if ack != nil {
			_, chanCap, err := im.keeper.LookupModuleByChannel(ctx, rollappPacket.Packet.DestinationPort, rollappPacket.Packet.DestinationChannel)
			if err != nil {
				logger.Error("Error looking up module by channel", "rollappID", rollappID, "error", "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				continue
			}
			err = im.keeper.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, ack)
			if err != nil {
				logger.Error("Error writing acknowledgement", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				continue
			}

		}
	}
}

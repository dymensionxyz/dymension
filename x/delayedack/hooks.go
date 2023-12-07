package delayedack

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	"github.com/dymensionxyz/dymension/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

var _ rollapptypes.RollappHooks = &IBCMiddleware{}

func (im IBCMiddleware) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error {
	return nil
}

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

		var err error
		switch rollappPacket.PacketType {
		case types.RollappPacket_OnRecv:
			// Call the OnRecvPacket callback for each packet
			ack := im.app.OnRecvPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
			// Write the acknowledgement to the chain only if it is synchronous
			if ack != nil {
				var chanCap *capabilitytypes.Capability
				_, chanCap, err = im.keeper.LookupModuleByChannel(ctx, rollappPacket.Packet.DestinationPort, rollappPacket.Packet.DestinationChannel)
				if err != nil {
					logger.Error("Error looking up module by channel", "rollappID", rollappID, "error", "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
					break //break out of switch
				}
				err = im.keeper.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, ack)
				if err != nil {
					logger.Error("Error writing acknowledgement", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				}
			}
		case types.RollappPacket_OnAcknowledgement:
			// Call the OnAcknowledgementPacket callback for each packet
			err = im.app.OnAcknowledgementPacket(ctx, *rollappPacket.Packet, rollappPacket.Ack, rollappPacket.Relayer)
			if err != nil {
				logger.Error("Error calling OnAcknowledgementPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
			}
		case types.RollappPacket_OnTimeout:
			// Call the OnTimeoutPacket callback for each packet
			err = im.app.OnTimeoutPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
			if err != nil {
				logger.Error("Error calling OnTimeoutPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
			}
		default:
			err = fmt.Errorf("unknown packet type")
			logger.Error(err.Error(), "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "packet type", rollappPacket.PacketType)
		}

		// Update the packet status
		if err != nil {
			im.keeper.UpdateRollappPacketStatus(ctx, rollappID, rollappPacket, types.RollappPacket_REJECTED)
		} else {
			im.keeper.UpdateRollappPacketStatus(ctx, rollappID, rollappPacket, types.RollappPacket_ACCEPTED)
		}
	}
}

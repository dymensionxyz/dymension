package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ rollapptypes.RollappHooks = &Keeper{}

func (k Keeper) OnHardFork(ctx sdk.Context, rollappID string, fraudHeight uint64) {
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")

	// Get all the pending packets from fork height inclusive
	rollappPendingPackets := k.ListRollappPackets(ctx, types.PendingByRollappIDFromHeight(rollappID, fraudHeight))

	// Iterate over all the pending packets and revert them
	for _, rollappPacket := range rollappPendingPackets {
		logContext := []interface{}{
			"rollappID", rollappID,
			"sourceChannel", rollappPacket.Packet.SourceChannel,
			"destChannel", rollappPacket.Packet.DestinationChannel,
			"type", rollappPacket.Type,
			"sequence", rollappPacket.Packet.Sequence,
		}

		pendingAddr := ""
		transfer := rollappPacket.MustGetTransferPacketData()
		if rollappPacket.Type == commontypes.RollappPacket_ON_ACK || rollappPacket.Type == commontypes.RollappPacket_ON_TIMEOUT {
			// for sent packets, we restore the packet commitment
			// the packet will be handled over the new rollapp revision
			commitment := channeltypes.CommitPacket(k.cdc, rollappPacket.Packet)
			k.channelKeeper.SetPacketCommitment(ctx, rollappPacket.Packet.SourcePort, rollappPacket.Packet.SourceChannel, rollappPacket.Packet.Sequence, commitment)
			pendingAddr = transfer.Sender
		} else {
			// for incoming packets, we need to reset the packet receipt
			ibcPacket := rollappPacket.Packet
			k.deletePacketReceipt(ctx, ibcPacket.GetDestPort(), ibcPacket.GetDestChannel(), ibcPacket.GetSequence())
			pendingAddr = transfer.Receiver
		}

		// delete the packet
		err := k.DeleteRollappPacket(ctx, &rollappPacket)
		if err != nil {
			logger.Error("failed to delete reverted packet", append(logContext, "error", err.Error())...)
			continue
		}

		// delete the pending packet
		err = k.DeletePendingPacketByAddress(ctx, pendingAddr, rollappPacket.RollappPacketKey())
		if err != nil {
			logger.Error("failed to delete reverted pending packet", append(logContext, "error", err.Error())...)
			continue
		}

		logger.Debug("reverted IBC rollapp packet", logContext...)
	}

	logger.Info("reverting IBC rollapp packets", "rollappID", rollappID, "numPackets", len(rollappPendingPackets))
}

// DeleteRollappPacket deletes a packet receipt from the store
func (k Keeper) deletePacketReceipt(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := ctx.KVStore(k.channelKeeperStoreKey)
	store.Delete(host.PacketReceiptKey(portID, channelID, sequence))
}

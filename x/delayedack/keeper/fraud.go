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

func (k Keeper) OnHardFork(ctx sdk.Context, rollappID string, newRevisionHeight uint64) error {
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")

	// Get all the pending packets from fork height inclusive
	rollappPendingPackets := k.ListRollappPackets(ctx, types.PendingByRollappIDFromHeight(rollappID, newRevisionHeight))

	// Iterate over all the pending packets and revert them
	for _, rollappPacket := range rollappPendingPackets {
		logContext := []interface{}{
			"rollappID", rollappID,
			"sourceChannel", rollappPacket.Packet.SourceChannel,
			"destChannel", rollappPacket.Packet.DestinationChannel,
			"type", rollappPacket.Type,
			"sequence", rollappPacket.Packet.Sequence,
		}

		if rollappPacket.Type == commontypes.RollappPacket_ON_ACK || rollappPacket.Type == commontypes.RollappPacket_ON_TIMEOUT {
			// for sent packets, we restore the packet commitment
			// the packet will be handled over the new rollapp revision
			// we update the packet to the original transfer target and restore the packet commitment
			commitment := channeltypes.CommitPacket(k.cdc, rollappPacket.RestoreOriginalTransferTarget().Packet)
			k.channelKeeper.SetPacketCommitment(ctx, rollappPacket.Packet.SourcePort, rollappPacket.Packet.SourceChannel, rollappPacket.Packet.Sequence, commitment)
		} else {
			// for incoming packets, we need to reset the packet receipt
			ibcPacket := rollappPacket.Packet
			k.deletePacketReceipt(ctx, ibcPacket.GetDestPort(), ibcPacket.GetDestChannel(), ibcPacket.GetSequence())
		}

		// delete the packet
		k.DeleteRollappPacket(ctx, &rollappPacket)

		logger.Debug("reverted IBC rollapp packet", logContext...)
	}

	logger.Info("reverting IBC rollapp packets", "rollappID", rollappID, "numPackets", len(rollappPendingPackets))

	return nil
}

// DeleteRollappPacket deletes a packet receipt from the store
func (k Keeper) deletePacketReceipt(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := ctx.KVStore(k.channelKeeperStoreKey)
	store.Delete(host.PacketReceiptKey(portID, channelID, sequence))
}

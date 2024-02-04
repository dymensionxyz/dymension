package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// SetRollappPacket stores a rollapp packet in the KVStore.
// It logs the saving of the packet and marshals the packet into bytes before storing.
// The key for the packet is generated using the rollappID, proofHeight and the packet itself.
func (k Keeper) SetRollappPacket(ctx sdk.Context, rollappPacket types.RollappPacket) {
	logger := ctx.Logger()
	logger.Debug("Saving rollapp packet", "rollappID", rollappPacket.RollappId, "channel", rollappPacket.Packet.DestinationChannel,
		"sequence", rollappPacket.Packet.Sequence, "proofHeight", rollappPacket.ProofHeight, "type", rollappPacket.Type)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))
	b := k.cdc.MustMarshal(&rollappPacket)
	store.Set(types.GetRollappPacketKey(
		rollappPacket.RollappId,
		types.RollappPacket_PENDING,
		rollappPacket.ProofHeight,
		*rollappPacket.Packet,
	), b)
}

// UpdateRollappPacketStatus deletes the current rollapp packet and creates a new one with and updated status under a new key.
// It assumes that the packet has been previously stored with the pending status.
func (k Keeper) UpdateRollappPacketStatus(ctx sdk.Context, rollappID string, rollappPacket types.RollappPacket, newStatus types.RollappPacket_Status) types.RollappPacket {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))

	// Delete the old rollapp packet
	oldKey := types.GetRollappPacketKey(rollappID, types.RollappPacket_PENDING, rollappPacket.ProofHeight, *rollappPacket.Packet)
	store.Delete(oldKey)

	// Update the packet
	rollappPacket.Status = newStatus

	// Create a new rollapp packet with the updated status
	newKey := types.GetRollappPacketKey(rollappID, newStatus, rollappPacket.ProofHeight, *rollappPacket.Packet)
	b := k.cdc.MustMarshal(&rollappPacket)
	store.Set(newKey, b)

	return rollappPacket
}

// ListRollappPendingPackets retrieves a list of pending rollapp packets from the KVStore.
// It builds a prefix using the rollappID and the pending status, and iterates over the range from lastProofHeight to proofHeight.
// If the packet's proofHeight is less than or equal to the maxProofHeight, it is added to the list.
// The function returns the list of pending packets.
func (k Keeper) ListRollappPendingPackets(
	ctx sdk.Context,
	rollappId string,
	maxProofHeight uint64,
) (list []types.RollappPacket) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))

	// Build the prefix which is composed of the rollappID and the status
	var prefix []byte
	prefix = append(prefix, []byte(rollappId)...)
	prefix = append(prefix, []byte("/")...)

	// Append the pending status to the prefix
	pendingStatusBytes := []byte(fmt.Sprint(types.RollappPacket_PENDING))
	prefix = append(prefix, pendingStatusBytes...)
	prefix = append(prefix, []byte("/")...)

	// Iterate over the range from lastProofHeight to proofHeight
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.RollappPacket
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.ProofHeight <= maxProofHeight {
			list = append(list, val)
		} else {
			break
		}
	}

	return list
}

func (k Keeper) GetAllRollappPackets(ctx sdk.Context) (list []types.RollappPacket) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))

	// Iterate over the range from lastProofHeight to proofHeight
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.RollappPacket
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return list
}

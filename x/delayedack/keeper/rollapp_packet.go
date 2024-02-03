package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// SetRollappPacket stores a rollapp packet in the KVStore.
// It logs the saving of the packet and marshals the packet into bytes before storing.
// The key for the packet is generated using the rollappID, proofHeight and the packet itself.
func (k Keeper) SetRollappPacket(ctx sdk.Context, rollappID string, rollappPacket types.RollappPacket) {
	logger := ctx.Logger()
	logger.Debug("Saving rollapp packet", "rollappID", rollappID, "channel", rollappPacket.Packet.DestinationChannel,
		"sequence", rollappPacket.Packet.Sequence, "proofHeight", rollappPacket.ProofHeight, "type", rollappPacket.Type)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))
	b := k.cdc.MustMarshal(&rollappPacket)
	store.Set(types.GetRollappPacketKey(
		rollappID,
		commontypes.Status_PENDING,
		rollappPacket.ProofHeight,
		*rollappPacket.Packet,
	), b)
}

// GetRollappPacket retrieves a rollapp packet from the KVStore.
func (k Keeper) GetRollappPacket(ctx sdk.Context, rollappPacketKey string) *types.RollappPacket {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))
	b := store.Get([]byte(rollappPacketKey))
	if b == nil {
		return nil
	}

	var rollappPacket types.RollappPacket
	k.cdc.MustUnmarshal(b, &rollappPacket)
	return &rollappPacket
}

// UpdateRollappPacketRecipient updates the recipient of the underlying packet.
// Only pending packets can be updated.
func (k Keeper) UpdateRollappPacketRecipient(
	ctx sdk.Context,
	rollappPacketKey string,
	newRecipient string,
) error {
	rollappPacket := k.GetRollappPacket(ctx, rollappPacketKey)
	if rollappPacket.Status != commontypes.Status_PENDING {
		return types.ErrCanOnlyUpdatePendingPacket
	}
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(rollappPacket.Packet.GetData(), &data); err != nil {
		return err
	}
	// Create a copy of the packet with the new recipient
	newPacketData := transfertypes.NewFungibleTokenPacketData(
		data.Denom,
		data.Amount,
		data.Sender,
		newRecipient,
		data.Memo,
	)
	// Marshall to binary and update the packet with this data
	packetBytes := newPacketData.GetBytes()
	packet := rollappPacket.Packet
	packet.Data = packetBytes
	// Update rollapp packet with the new updated packet and save in the store
	rollappPacket.Packet = packet
	b := k.cdc.MustMarshal(rollappPacket)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))
	store.Set([]byte(rollappPacketKey), b)
	return nil
}

// UpdateRollappPacketWithStatus deletes the current rollapp packet and creates a new one with and updated status under a new key.
// Updating the status should be called only with this method as it effects the key of the packet.
// The assumption is that the passed rollapp packet status field is not updated directly.
func (k *Keeper) UpdateRollappPacketWithStatus(ctx sdk.Context, rollappID string, rollappPacket types.RollappPacket, newStatus commontypes.Status) types.RollappPacket {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))

	// Delete the old rollapp packet
	oldKey := types.GetRollappPacketKey(rollappID, rollappPacket.Status, rollappPacket.ProofHeight, *rollappPacket.Packet)
	store.Delete(oldKey)

	// Update the packet
	rollappPacket.Status = newStatus

	// Create a new rollapp packet with the updated status
	newKey := types.GetRollappPacketKey(rollappID, newStatus, rollappPacket.ProofHeight, *rollappPacket.Packet)
	b := k.cdc.MustMarshal(&rollappPacket)
	store.Set(newKey, b)

	// Call hook subscribers
	keeperHooks := k.GetHooks()
	err := keeperHooks.AfterPacketStatusUpdated(ctx, &rollappPacket, string(oldKey), string(newKey))
	if err != nil {
		panic("Error after updating packet status: " + err.Error())
	}
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
	pendingStatusBytes := []byte(fmt.Sprint(commontypes.Status_PENDING))
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

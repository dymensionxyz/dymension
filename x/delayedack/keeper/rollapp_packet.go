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
func (k Keeper) SetRollappPacket(ctx sdk.Context, rollappPacket commontypes.RollappPacket) error {
	logger := ctx.Logger()
	logger.Debug("Saving rollapp packet", "rollappID", rollappPacket.RollappId, "channel", rollappPacket.Packet.DestinationChannel,
		"sequence", rollappPacket.Packet.Sequence, "proofHeight", rollappPacket.ProofHeight, "type", rollappPacket.Type)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(commontypes.RollappPacketKeyPrefix))
	b, err := k.cdc.Marshal(&rollappPacket)
	if err != nil {
		return err
	}
	store.Set(commontypes.GetRollappPacketKey(
		rollappPacket.RollappId,
		rollappPacket.Status,
		rollappPacket.ProofHeight,
		*rollappPacket.Packet,
	), b)
	return nil
}

// GetRollappPacket retrieves a rollapp packet from the KVStore.
func (k Keeper) GetRollappPacket(ctx sdk.Context, rollappPacketKey string) (*commontypes.RollappPacket, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(commontypes.RollappPacketKeyPrefix))
	b := store.Get([]byte(rollappPacketKey))
	if b == nil {
		return nil, types.ErrRollappPacketDoesNotExist
	}

	var rollappPacket commontypes.RollappPacket
	if err := k.cdc.Unmarshal(b, &rollappPacket); err != nil {
		return nil, err
	}
	return &rollappPacket, nil
}

// UpdateRollappPacketRecipient updates the recipient of the underlying packet.
// Only pending packets can be updated.
func (k Keeper) UpdateRollappPacketRecipient(
	ctx sdk.Context,
	rollappPacketKey string,
	newRecipient string,
) error {
	rollappPacket, err := k.GetRollappPacket(ctx, rollappPacketKey)
	if err != nil {
		return err
	}
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
	err = k.SetRollappPacket(ctx, *rollappPacket)
	if err != nil {
		return err
	}
	return nil
}

// UpdateRollappPacketWithStatus deletes the current rollapp packet and creates a new one with and updated status under a new key.
// Updating the status should be called only with this method as it effects the key of the packet.
// The assumption is that the passed rollapp packet status field is not updated directly.
func (k *Keeper) UpdateRollappPacketWithStatus(ctx sdk.Context, rollappPacket commontypes.RollappPacket, newStatus commontypes.Status) (commontypes.RollappPacket, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(commontypes.RollappPacketKeyPrefix))

	// Delete the old rollapp packet
	oldKey := commontypes.GetRollappPacketKey(rollappPacket.RollappId, rollappPacket.Status, rollappPacket.ProofHeight, *rollappPacket.Packet)
	store.Delete(oldKey)

	// Update the packet
	rollappPacket.Status = newStatus

	// Create a new rollapp packet with the updated status
	err := k.SetRollappPacket(ctx, rollappPacket)
	if err != nil {
		return commontypes.RollappPacket{}, err
	}

	// Call hook subscribers
	newKey := commontypes.GetRollappPacketKey(rollappPacket.RollappId, newStatus, rollappPacket.ProofHeight, *rollappPacket.Packet)
	keeperHooks := k.GetHooks()
	err = keeperHooks.AfterPacketStatusUpdated(ctx, &rollappPacket, string(oldKey), string(newKey))
	if err != nil {
		return commontypes.RollappPacket{}, err
	}
	return rollappPacket, nil
}

// ListRollappPendingPackets retrieves a list of pending rollapp packets from the KVStore.
// It builds a prefix using the rollappID and the pending status, and iterates over the range from lastProofHeight to proofHeight.
// If the packet's proofHeight is less than or equal to the maxProofHeight, it is added to the list.
// The function returns the list of pending packets.
func (k Keeper) ListRollappPendingPackets(
	ctx sdk.Context,
	rollappId string,
	maxProofHeight uint64,
) (list []commontypes.RollappPacket) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(commontypes.RollappPacketKeyPrefix))

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
		var val commontypes.RollappPacket
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.ProofHeight <= maxProofHeight {
			list = append(list, val)
		} else {
			break
		}
	}

	return list
}

func (k Keeper) GetAllRollappPackets(ctx sdk.Context) (list []commontypes.RollappPacket) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(commontypes.RollappPacketKeyPrefix))

	// Iterate over the range from lastProofHeight to proofHeight
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val commontypes.RollappPacket
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return list
}

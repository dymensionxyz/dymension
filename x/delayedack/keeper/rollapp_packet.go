package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// SetRollappPacket stores a rollapp packet in the KVStore.
// It logs the saving of the packet and marshals the packet into bytes before storing.
// The key for the packet is generated using the rollappID, proofHeight and the packet itself.
func (k Keeper) SetRollappPacket(ctx sdk.Context, rollappPacket commontypes.RollappPacket) {
	store := ctx.KVStore(k.storeKey)
	rollappPacketKey := rollappPacket.RollappPacketKey()
	b := k.cdc.MustMarshal(&rollappPacket)
	store.Set(rollappPacketKey, b)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelayedAck,
			rollappPacket.GetEvents()...,
		),
	)
}

// SetPendingPacketByAddress stores a rollapp packet in the KVStore by its receiver.
// Helper index to query all packets by receiver.
func (k Keeper) SetPendingPacketByAddress(ctx sdk.Context, receiver string, rollappPacketKey []byte) error {
	return k.pendingPacketsByAddress.Set(ctx, collections.Join(receiver, rollappPacketKey))
}

// MustSetPendingPacketByAddress stores a rollapp packet in the KVStore by its receiver.
// Helper index to query all packets by receiver. Panics on encoding errors.
func (k Keeper) MustSetPendingPacketByAddress(ctx sdk.Context, receiver string, rollappPacketKey []byte) {
	err := k.SetPendingPacketByAddress(ctx, receiver, rollappPacketKey)
	if err != nil {
		panic(err)
	}
}

// DeletePendingPacketByAddress deletes a rollapp packet from the KVStore by its receiver.
func (k Keeper) DeletePendingPacketByAddress(ctx sdk.Context, receiver string, rollappPacketKey []byte) error {
	return k.pendingPacketsByAddress.Remove(ctx, collections.Join(receiver, rollappPacketKey))
}

// MustDeletePendingPacketByAddress deletes a rollapp packet from the KVStore by its receiver.
// Panics on encoding error. Do not panic if the key is not found.
func (k Keeper) MustDeletePendingPacketByAddress(ctx sdk.Context, receiver string, rollappPacketKey []byte) {
	err := k.DeletePendingPacketByAddress(ctx, receiver, rollappPacketKey)
	if err != nil {
		panic(err)
	}
}

// GetPendingPacketsByAddress retrieves rollapp packets from the KVStore by their receiver.
func (k Keeper) GetPendingPacketsByAddress(ctx sdk.Context, receiver string) ([]commontypes.RollappPacket, error) {
	var packets []commontypes.RollappPacket
	rng := collections.NewPrefixedPairRange[string, []byte](receiver)
	err := k.pendingPacketsByAddress.Walk(ctx, rng, func(key collections.Pair[string, []byte]) (stop bool, err error) {
		packet, err := k.GetRollappPacket(ctx, string(key.K2()))
		if err != nil {
			return true, err
		}
		packets = append(packets, *packet)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return packets, nil
}

// GetRollappPacket retrieves a rollapp packet from the KVStore.
func (k Keeper) GetRollappPacket(ctx sdk.Context, rollappPacketKey string) (*commontypes.RollappPacket, error) {
	store := ctx.KVStore(k.storeKey)
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

// UpdateRollappPacketTransferAddress updates the recipient of the underlying packet.
// Only pending packets can be updated.
func (k Keeper) UpdateRollappPacketTransferAddress(
	ctx sdk.Context,
	rollappPacketKey string,
	address string,
) error {
	rollappPacket, err := k.GetRollappPacket(ctx, rollappPacketKey)
	if err != nil {
		return err
	}
	if rollappPacket.Status != commontypes.Status_PENDING {
		return types.ErrCanOnlyUpdatePendingPacket
	}

	transferPacketData, err := rollappPacket.GetTransferPacketData()
	if err != nil {
		return err
	}

	// Set the recipient and sender based on the rollapp packet type
	var (
		recipient              = transferPacketData.Receiver
		sender                 = transferPacketData.Sender
		originalTransferTarget string
	)
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_RECV:
		originalTransferTarget = recipient
		recipient = address
	case commontypes.RollappPacket_ON_ACK, commontypes.RollappPacket_ON_TIMEOUT:
		originalTransferTarget = sender
		sender = address
	}

	// Create a new packet data with the updated recipient and sender
	newPacketData := transfertypes.NewFungibleTokenPacketData(
		transferPacketData.Denom,
		transferPacketData.Amount,
		sender,
		recipient,
		transferPacketData.Memo,
	)

	// Marshall to binary and update the packet with this data
	packet := rollappPacket.Packet
	packet.Data = newPacketData.GetBytes()
	// Update rollapp packet with the new updated packet and save in the store
	rollappPacket.Packet = packet
	rollappPacket.OriginalTransferTarget = originalTransferTarget

	// Update index: delete the old packet and save the new one
	k.MustDeletePendingPacketByAddress(ctx, originalTransferTarget, []byte(rollappPacketKey))
	k.MustSetPendingPacketByAddress(ctx, address, rollappPacket.RollappPacketKey())

	// Save updated rollapp packet
	k.SetRollappPacket(ctx, *rollappPacket)

	return nil
}

// UpdateRollappPacketAfterFinalization deletes the current rollapp packet and creates a new one with and updated status under a new key.
// Updating the status should be called only with this method as it effects the key of the packet.
// The assumption is that the passed rollapp packet status field is not updated directly.
func (k *Keeper) UpdateRollappPacketAfterFinalization(ctx sdk.Context, rollappPacket commontypes.RollappPacket) (commontypes.RollappPacket, error) {
	if rollappPacket.Status != commontypes.Status_PENDING {
		return commontypes.RollappPacket{}, types.ErrCanOnlyUpdatePendingPacket
	}

	transferPacketData, err := rollappPacket.GetTransferPacketData()
	if err != nil {
		return commontypes.RollappPacket{}, err
	}

	oldKey := rollappPacket.RollappPacketKey()

	// Delete the old packet from the pending packets index depending on the packet type
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_RECV:
		k.MustDeletePendingPacketByAddress(ctx, transferPacketData.Receiver, oldKey)
	case commontypes.RollappPacket_ON_ACK, commontypes.RollappPacket_ON_TIMEOUT:
		k.MustDeletePendingPacketByAddress(ctx, transferPacketData.Sender, oldKey)
	}

	// Delete the old rollapp packet
	store := ctx.KVStore(k.storeKey)
	store.Delete(oldKey)

	// Update the packet
	rollappPacket.Status = commontypes.Status_FINALIZED
	// Create a new rollapp packet with the updated status
	k.SetRollappPacket(ctx, rollappPacket)

	newKey := rollappPacket.RollappPacketKey()

	// Call hook subscribers
	keeperHooks := k.GetHooks()
	err = keeperHooks.AfterPacketStatusUpdated(ctx, &rollappPacket, string(oldKey), string(newKey))
	if err != nil {
		return rollappPacket, err
	}
	return rollappPacket, nil
}

// ListRollappPackets retrieves a list rollapp packets from the KVStore by applying the given filter
func (k Keeper) ListRollappPackets(ctx sdk.Context, listFilter types.RollappPacketListFilter) (list []commontypes.RollappPacket) {
	store := ctx.KVStore(k.storeKey)
	withLimit := listFilter.Limit > 0
	// Iterate over the range of filters and get all the rollapp packets
	// that meet the filter criteria
outer:
	for _, pref := range listFilter.Prefixes {
		if len(pref.Start) == 0 {
			pref.Start = commontypes.AllRollappPacketKeyPrefix
		}
		if len(pref.End) == 0 {
			pref.End = sdk.PrefixEndBytes(pref.Start)
		}
		iterator := store.Iterator(pref.Start, pref.End)
		for ; iterator.Valid(); iterator.Next() {
			var val commontypes.RollappPacket
			k.cdc.MustUnmarshal(iterator.Value(), &val)
			// Apply the filter function
			if !listFilter.FilterFunc(val) {
				continue
			}
			list = append(list, val)

			if withLimit && len(list) == listFilter.Limit {
				_ = iterator.Close()
				break outer
			}
		}
		_ = iterator.Close()
	}

	return list
}

func (k Keeper) GetAllRollappPackets(ctx sdk.Context) (list []commontypes.RollappPacket) {
	store := ctx.KVStore(k.storeKey)

	// Iterate over the range from lastProofHeight to proofHeight
	iterator := sdk.KVStorePrefixIterator(store, commontypes.AllRollappPacketKeyPrefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val commontypes.RollappPacket
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return list
}

func (k Keeper) DeleteRollappPacket(ctx sdk.Context, rollappPacket *commontypes.RollappPacket) {
	store := ctx.KVStore(k.storeKey)
	rollappPacketKey := rollappPacket.RollappPacketKey()
	store.Delete(rollappPacketKey)

	// delete the PacketByAddress index
	pendingAddr := ""
	transfer := rollappPacket.MustGetTransferPacketData()
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_RECV:
		pendingAddr = transfer.Receiver
	case commontypes.RollappPacket_ON_ACK, commontypes.RollappPacket_ON_TIMEOUT:
		pendingAddr = transfer.Sender
	}
	k.MustDeletePendingPacketByAddress(ctx, pendingAddr, rollappPacket.RollappPacketKey())

	keeperHooks := k.GetHooks()
	// TODO: can call eIBC directly
	keeperHooks.AfterPacketDeleted(ctx, rollappPacket)
}

// GetPendingPacketsUntilFinalizedHeight returns all pending rollapp packets until the latest finalized height.
func (k Keeper) GetPendingPacketsUntilFinalizedHeight(ctx sdk.Context, rollappID string) ([]commontypes.RollappPacket, uint64, error) {
	// Get rollapp's latest finalized height
	latestFinalizedHeight, err := k.getRollappLatestFinalizedHeight(ctx, rollappID)
	if err != nil {
		return nil, 0, fmt.Errorf("get latest finalized height: rollapp '%s': %w", rollappID, err)
	}

	// Get all pending rollapp packets until the latest finalized height
	return k.ListRollappPackets(ctx, types.PendingByRollappIDByMaxHeight(rollappID, latestFinalizedHeight)), latestFinalizedHeight, nil
}

func (k Keeper) getRollappLatestFinalizedHeight(ctx sdk.Context, rollappID string) (uint64, error) {
	latestIndex, found := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, rollappID)
	if !found {
		return 0, gerrc.ErrNotFound.Wrapf("latest finalized state index is not found")
	}

	stateInfo := k.rollappKeeper.MustGetStateInfo(ctx, rollappID, latestIndex.Index)
	return stateInfo.GetLatestHeight(), nil
}

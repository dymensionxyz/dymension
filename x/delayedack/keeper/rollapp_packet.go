package keeper

import (
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
	store := ctx.KVStore(k.storeKey)
	rollappPacketKey := commontypes.RollappPacketKey(&rollappPacket)
	b, err := k.cdc.Marshal(&rollappPacket)
	if err != nil {
		return err
	}
	store.Set(rollappPacketKey, b)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelayedAck,
			rollappPacket.GetEvents()...,
		),
	)
	return nil
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
	recipient, sender := transferPacketData.Receiver, transferPacketData.Sender
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_RECV:
		recipient = address
	case commontypes.RollappPacket_ON_TIMEOUT:
		fallthrough
	case commontypes.RollappPacket_ON_ACK:
		sender = address
	}
	newPacketData := transfertypes.NewFungibleTokenPacketData(
		transferPacketData.Denom,
		transferPacketData.Amount,
		sender,
		recipient,
		transferPacketData.Memo,
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
func (k *Keeper) UpdateRollappPacketWithStatus(
	ctx sdk.Context,
	rollappPacket commontypes.RollappPacket,
	newStatus commontypes.Status,
) (commontypes.RollappPacket, error) {
	store := ctx.KVStore(k.storeKey)

	// Delete the old rollapp packet
	oldKey := commontypes.RollappPacketKey(&rollappPacket)
	store.Delete(oldKey)
	// Update the packet
	rollappPacket.Status = newStatus
	// Create a new rollapp packet with the updated status
	err := k.SetRollappPacket(ctx, rollappPacket)
	if err != nil {
		return commontypes.RollappPacket{}, err
	}

	// Call hook subscribers
	newKey := commontypes.RollappPacketKey(&rollappPacket)
	keeperHooks := k.GetHooks()
	err = keeperHooks.AfterPacketStatusUpdated(ctx, &rollappPacket, string(oldKey), string(newKey))
	if err != nil {
		return commontypes.RollappPacket{}, err
	}
	return rollappPacket, nil
}

// ListRollappPackets retrieves a list of rollapp packets from the KVStore by applying the provided filter
func (k Keeper) ListRollappPackets(ctx sdk.Context, listFilter rollappPacketListFilter) (list []commontypes.RollappPacket) {
	store := ctx.KVStore(k.storeKey)

	for _, prefix := range listFilter.prefixes {
		iterator := sdk.KVStorePrefixIterator(store, prefix)
		packetsForStatus := k.iterateOverRollappPacketsPerStatus(iterator, listFilter.filter, listFilter.breakOnMismatch)
		list = append(list, packetsForStatus...)
	}

	return
}

func (k Keeper) iterateOverRollappPacketsPerStatus(
	iterator sdk.Iterator,
	filter filterFunc,
	breakOnMismatch bool,
) (list []commontypes.RollappPacket) {
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val commontypes.RollappPacket
		k.cdc.MustUnmarshal(iterator.Value(), &val)

		if filter == nil || filter(val) {
			list = append(list, val)
		} else if breakOnMismatch {
			break
		}
	}

	return list
}

func (k Keeper) deleteRollappPacket(ctx sdk.Context, rollappPacket *commontypes.RollappPacket) error {
	rollappPacketKey := commontypes.RollappPacketKey(rollappPacket)

	ctx.KVStore(k.storeKey).Delete(rollappPacketKey)

	err := k.GetHooks().AfterPacketDeleted(ctx, rollappPacket)
	if err != nil {
		return err
	}

	return nil
}

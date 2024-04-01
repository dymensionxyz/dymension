package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ctypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// SetRollappPacket stores a rollapp packet in the KVStore.
// It logs the saving of the packet and marshals the packet into bytes before storing.
// The key for the packet is generated using the rollappID, proofHeight and the packet itself.
func (k Keeper) SetRollappPacket(ctx sdk.Context, rollappPacket ctypes.RollappPacket) error {
	logger := ctx.Logger()
	logger.Debug("Saving rollapp packet", "rollappID", rollappPacket.RollappId, "channel", rollappPacket.Packet.DestinationChannel,
		"sequence", rollappPacket.Packet.Sequence, "proofHeight", rollappPacket.ProofHeight, "type", rollappPacket.Type)

	b, err := k.cdc.Marshal(&rollappPacket)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	// set the rollapp packet
	rollappPacketKey := ctypes.RollappPacketKey(&rollappPacket) // rollappPacketKey -> rollappPacket
	store.Set(rollappPacketKey, b)
	// set the index for rollapp packet
	indexKey := ctypes.RollappPacketIndexKey(&rollappPacket)
	store.Set(indexKey, rollappPacketKey) // indexKey -> rollappPacketKey

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelayedAck,
			rollappPacket.GetEvents()...,
		),
	)
	return nil
}

// GetRollappPacket retrieves a rollapp packet from the KVStore.
func (k Keeper) GetRollappPacket(ctx sdk.Context, rollappPacketKey string) (*ctypes.RollappPacket, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get([]byte(rollappPacketKey))
	if b == nil {
		return nil, types.ErrRollappPacketDoesNotExist
	}

	var rollappPacket ctypes.RollappPacket
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
	if rollappPacket.Status != ctypes.Status_PENDING {
		return types.ErrCanOnlyUpdatePendingPacket
	}
	transferPacketData, err := rollappPacket.GetTransferPacketData()
	if err != nil {
		return err
	}
	// Set the recipient and sender based on the rollapp packet type
	recipient, sender := transferPacketData.Receiver, transferPacketData.Sender
	switch rollappPacket.Type {
	case ctypes.RollappPacket_ON_RECV:
		recipient = address
	case ctypes.RollappPacket_ON_TIMEOUT:
		fallthrough
	case ctypes.RollappPacket_ON_ACK:
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
	rollappPacket ctypes.RollappPacket,
	newStatus ctypes.Status,
) (ctypes.RollappPacket, error) {
	store := ctx.KVStore(k.storeKey)

	// Delete the old rollapp packet
	oldKey := ctypes.RollappPacketKey(&rollappPacket)
	store.Delete(oldKey)
	// Update the packet
	rollappPacket.Status = newStatus
	// Create a new rollapp packet with the updated status
	err := k.SetRollappPacket(ctx, rollappPacket)
	if err != nil {
		return ctypes.RollappPacket{}, err
	}

	// Call hook subscribers
	newKey := ctypes.RollappPacketKey(&rollappPacket)
	keeperHooks := k.GetHooks()
	err = keeperHooks.AfterPacketStatusUpdated(ctx, &rollappPacket, string(oldKey), string(newKey))
	if err != nil {
		return ctypes.RollappPacket{}, err
	}
	return rollappPacket, nil
}

// ListRollappPackets retrieves a list of rollapp packets from the KVStore by applying the provided filter
func (k Keeper) ListRollappPackets(ctx sdk.Context) (list []ctypes.RollappPacket) {
	return k.listRollappPacketsByPrefix(ctx, ctypes.RollappPacketKeyPrefix, nil, false)
}

// ListRollappPacketsByRollappID retrieves a list of rollapp packets from the KVStore by rollappID
func (k Keeper) ListRollappPacketsByRollappID(
	ctx sdk.Context,
	rollappID string,
) (list []ctypes.RollappPacket) {
	return k.listRollappPacketsByPrefix(ctx, ctypes.RollappPacketByRollappIDPrefix(rollappID), nil, false)
}

// ListRollappPacketsByRollappIDByStatus retrieves a list of rollapp packets from the KVStore by rollappID and status
func (k Keeper) ListRollappPacketsByRollappIDByStatus(
	ctx sdk.Context,
	rollappID string,
	status ctypes.Status,
) (list []ctypes.RollappPacket) {
	return k.listRollappPacketsByPrefix(ctx, ctypes.RollappPacketByRollappIDByStatusPrefix(rollappID, status), nil, false)
}

// ListRollappPacketsByStatus retrieves a list of rollapp packets from the KVStore by status
func (k Keeper) ListRollappPacketsByStatus(
	ctx sdk.Context,
	status ctypes.Status,
) (list []ctypes.RollappPacket) {
	return k.listRollappPacketsByPrefix(ctx, ctypes.RollappPacketByStatusIndexPrefix(status), nil, true)
}

// ListPendingRollappPacketsByRollappIDByMaxHeight retrieves a list of pending rollapp packets from the KVStore by rollappID and max proof height
func (k Keeper) ListPendingRollappPacketsByRollappIDByMaxHeight(
	ctx sdk.Context,
	rollappID string,
	maxProofHeight uint64,
) (list []ctypes.RollappPacket) {
	start, end := ctypes.RollappPacketByRollappIDByStatusByMaxProofHeightPrefixes(rollappID, ctypes.Status_PENDING, maxProofHeight)
	return k.listRollappPacketsByPrefix(ctx, start, end, false)
}

func (k Keeper) listRollappPacketsByPrefix(
	ctx sdk.Context,
	prefix,
	suffix []byte,
	indexed bool,
) (list []ctypes.RollappPacket) {
	store := ctx.KVStore(k.storeKey)
	if suffix == nil {
		suffix = sdk.PrefixEndBytes(prefix)
	}
	iterator := store.Iterator(prefix, suffix)
	defer iterator.Close() //nolint:errcheck

	for ; iterator.Valid(); iterator.Next() {
		var value []byte
		if indexed {
			value = store.Get(iterator.Value())
		} else {
			value = iterator.Value()
		}
		var val ctypes.RollappPacket
		k.cdc.MustUnmarshal(value, &val)
		list = append(list, val)
	}
	return
}

func (k Keeper) deleteRollappPacket(ctx sdk.Context, rollappPacket *ctypes.RollappPacket) error {
	rollappPacketRollappIDKey := ctypes.RollappPacketKey(rollappPacket)
	rollappPacketStatusIndexKey := ctypes.RollappPacketIndexKey(rollappPacket)

	ctx.KVStore(k.storeKey).Delete(rollappPacketRollappIDKey)
	ctx.KVStore(k.storeKey).Delete(rollappPacketStatusIndexKey)

	if err := k.GetHooks().AfterPacketDeleted(ctx, rollappPacket); err != nil {
		return err
	}

	return nil
}

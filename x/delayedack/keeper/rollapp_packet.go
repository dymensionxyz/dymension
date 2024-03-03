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
	rollappPacketKey, err := commontypes.RollappPacketKey(&rollappPacket)
	if err != nil {
		return err
	}
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

// UpdateRollappPacketTransferData updates the recipient of the underlying packet.
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
	if rollappPacket.Type == commontypes.RollappPacket_ON_RECV {
		recipient = address
	} else if rollappPacket.Type == commontypes.RollappPacket_ON_TIMEOUT {
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
func (k *Keeper) UpdateRollappPacketWithStatus(ctx sdk.Context, rollappPacket commontypes.RollappPacket, newStatus commontypes.Status) (commontypes.RollappPacket, error) {
	store := ctx.KVStore(k.storeKey)

	// Delete the old rollapp packet
	oldKey, err := commontypes.RollappPacketKey(&rollappPacket)
	if err != nil {
		return commontypes.RollappPacket{}, err
	}
	store.Delete(oldKey)
	// Update the packet
	rollappPacket.Status = newStatus
	// Create a new rollapp packet with the updated status
	err = k.SetRollappPacket(ctx, rollappPacket)
	if err != nil {
		return commontypes.RollappPacket{}, err
	}

	// Call hook subscribers
	newKey, err := commontypes.RollappPacketKey(&rollappPacket)
	if err != nil {
		return commontypes.RollappPacket{}, err
	}
	keeperHooks := k.GetHooks()
	err = keeperHooks.AfterPacketStatusUpdated(ctx, &rollappPacket, string(oldKey), string(newKey))
	if err != nil {
		return commontypes.RollappPacket{}, err
	}
	return rollappPacket, nil
}

// ListRollappPacketsByStatus retrieves a list of pending rollapp packets from the KVStore.
// It builds a prefix using the rollappID and the pending status, and iterates over the range from lastProofHeight to proofHeight.
// If the packet's proofHeight is less than or equal to the maxProofHeight, it is added to the list.
// if maxProofHeight is 0, all packets are returned.
// The function returns the list of pending packets.
func (k Keeper) ListRollappPacketsByStatus(
	ctx sdk.Context,
	status commontypes.Status,
	maxProofHeight uint64,
) (list []commontypes.RollappPacket) {
	logger := ctx.Logger()
	store := ctx.KVStore(k.storeKey)
	// switch prefix based on status
	statusPrefix, err := commontypes.GetStatusBytes(status)
	if err != nil {
		logger.Error("Failed to get status bytes", "error", err)
		return nil
	}
	// Iterate over the range from lastProofHeight to proofHeight.
	// We are guaranteed order by the proof height so can break early if we
	// find a packet with a proof height greater than maxProofHeight
	iterator := sdk.KVStorePrefixIterator(store, statusPrefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val commontypes.RollappPacket
		err := k.cdc.Unmarshal(iterator.Value(), &val)
		if err != nil {
			logger.Error("Failed to unmarshal rollapp packet", "error", err)
			continue
		}
		if maxProofHeight == 0 || val.ProofHeight <= maxProofHeight {
			list = append(list, val)
		} else {
			break
		}
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
		err := k.cdc.Unmarshal(iterator.Value(), &val)
		if err != nil {
			ctx.Logger().Error("Failed to unmarshal rollapp packet", "error", err)
			continue
		}
		list = append(list, val)
	}

	return list
}

func (k Keeper) deleteRollappPacket(ctx sdk.Context, rollappPacket *commontypes.RollappPacket) error {
	store := ctx.KVStore(k.storeKey)
	rollappPacketKey, err := commontypes.RollappPacketKey(rollappPacket)
	if err != nil {
		return err
	}
	store.Delete(rollappPacketKey)

	keeperHooks := k.GetHooks()
	err = keeperHooks.AfterPacketDeleted(ctx, rollappPacket)
	if err != nil {
		return err
	}

	return nil
}

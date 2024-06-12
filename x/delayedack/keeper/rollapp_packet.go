package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/dymension/v3/utils/gerr"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// SetRollappPacket stores a rollapp packet in the KVStore.
// It logs the saving of the packet and marshals the packet into bytes before storing.
// The key for the packet is generated using the rollappID, proofHeight and the packet itself.
func (k Keeper) SetRollappPacket(ctx sdk.Context, rollappPacket commontypes.RollappPacket) {
	store := ctx.KVStore(k.storeKey)
	rollappPacketKey := commontypes.RollappPacketKey(&rollappPacket)
	b := k.cdc.MustMarshal(&rollappPacket)
	store.Set(rollappPacketKey, b)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelayedAck,
			rollappPacket.GetEvents()...,
		),
	)
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
	var originalTransferTarget string
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_RECV:
		originalTransferTarget = recipient
		recipient = address
	case commontypes.RollappPacket_ON_TIMEOUT:
		fallthrough
	case commontypes.RollappPacket_ON_ACK:
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
	packetBytes := newPacketData.GetBytes()
	packet := rollappPacket.Packet
	packet.Data = packetBytes
	// Update rollapp packet with the new updated packet and save in the store
	rollappPacket.Packet = packet
	rollappPacket.OriginalTransferTarget = originalTransferTarget
	k.SetRollappPacket(ctx, *rollappPacket)
	return nil
}

// UpdateRollappPacketWithStatus deletes the current rollapp packet and creates a new one with and updated status under a new key.
// Updating the status should be called only with this method as it effects the key of the packet.
// The assumption is that the passed rollapp packet status field is not updated directly.
func (k *Keeper) UpdateRollappPacketWithStatus(ctx sdk.Context, rollappPacket commontypes.RollappPacket, newStatus commontypes.Status) (commontypes.RollappPacket, error) {
	store := ctx.KVStore(k.storeKey)

	// Delete the old rollapp packet
	oldKey := commontypes.RollappPacketKey(&rollappPacket)
	store.Delete(oldKey)
	// Update the packet
	rollappPacket.Status = newStatus
	// Create a new rollapp packet with the updated status
	k.SetRollappPacket(ctx, rollappPacket)

	// Call hook subscribers
	newKey := commontypes.RollappPacketKey(&rollappPacket)
	keeperHooks := k.GetHooks()
	err := keeperHooks.AfterPacketStatusUpdated(ctx, &rollappPacket, string(oldKey), string(newKey))
	if err != nil {
		return rollappPacket, err
	}
	return rollappPacket, nil
}

// ListRollappPackets retrieves a list rollapp packets from the KVStore by applying the given filter
func (k Keeper) ListRollappPackets(ctx sdk.Context, listFilter types.RollappPacketListFilter) (list []commontypes.RollappPacket) {
	store := ctx.KVStore(k.storeKey)
	// Iterate over the range of filters and get all the rollapp packets
	// that meet the filter criteria
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

func (k Keeper) deleteRollappPacket(ctx sdk.Context, rollappPacket *commontypes.RollappPacket) error {
	store := ctx.KVStore(k.storeKey)
	rollappPacketKey := commontypes.RollappPacketKey(rollappPacket)
	store.Delete(rollappPacketKey)

	keeperHooks := k.GetHooks()
	err := keeperHooks.AfterPacketDeleted(ctx, rollappPacket)
	if err != nil {
		return err
	}

	return nil
}

// GetValidTransferWithFinalizationInfo does GetValidTransferFromReceivedPacket, but additionally it gets the finalization status and proof height
// of the packet.
func (k Keeper) GetValidTransferWithFinalizationInfo(
	ctx sdk.Context,
	packet channeltypes.Packet,
	packetType commontypes.RollappPacket_Type,
) (data types.TransferDataWithFinalization, err error) {
	switch packetType {
	case commontypes.RollappPacket_ON_RECV:
		data.TransferData, err = k.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	case commontypes.RollappPacket_ON_TIMEOUT, commontypes.RollappPacket_ON_ACK:
		data.TransferData, err = k.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetSourcePort(), packet.GetSourceChannel())
	}
	if err != nil {
		err = errorsmod.Wrap(err, "get valid transfer data")
	}

	packetId := commontypes.NewPacketUID(packetType, packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	height, ok := types.PacketProofHeightFromCtx(ctx, packetId)
	if !ok {
		// TODO: should probably be a panic
		err = errorsmod.Wrapf(gerr.ErrNotFound, "get proof height from context: packetID: %s", packetId)
		return
	}
	data.ProofHeight = height.RevisionHeight

	if !data.IsRollapp() {
		return
	}

	finalizedHeight, err := k.getRollappFinalizedHeight(ctx, data.Rollapp.RollappId)
	if errorsmod.IsOf(err, rollapptypes.ErrNoFinalizedStateYetForRollapp) {
		err = nil
	} else if err != nil {
		err = errorsmod.Wrap(err, "get rollapp finalized height")
		return
	} else {
		data.Finalized = finalizedHeight >= data.ProofHeight
	}

	return
}

package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayeacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

/* -------------------------------------------------------------------------- */
/*                              delayed ack hooks                             */
/* -------------------------------------------------------------------------- */
var _ delayeacktypes.DelayedAckHooks = delayedAckHooks{}

type delayedAckHooks struct {
	delayeacktypes.BaseDelayedAckHook
	Keeper
}

func (k Keeper) GetDelayedAckHooks() delayeacktypes.DelayedAckHooks {
	return delayedAckHooks{
		BaseDelayedAckHook: delayeacktypes.BaseDelayedAckHook{},
		Keeper:             k,
	}
}

// AfterPacketStatusUpdated is called every time the underlying IBC packet is updated.
// There are 2 assumptions here:
// 1. The packet status can change only once hence the oldPacketKey should always represent the order ID as it was created from it.
// 2. The packet status can only change from PENDING
func (d delayedAckHooks) AfterPacketStatusUpdated(ctx sdk.Context, packet *commontypes.RollappPacket,
	oldPacketKey string, newPacketKey string,
) error {
	demandOrderID := types.BuildDemandIDFromPacketKey(oldPacketKey)
	demandOrder, err := d.GetDemandOrder(ctx, commontypes.Status_PENDING, demandOrderID)
	if err != nil {
		// If demand order does not exist, then we don't need to do anything // TODO: why
		if errors.Is(err, types.ErrDemandOrderDoesNotExist) {
			return nil
		}
		return err
	}
	demandOrder.TrackingPacketKey = newPacketKey
	_, err = d.UpdateDemandOrderWithStatus(ctx, demandOrder, packet.Status)
	if err != nil {
		return err
	}

	return nil
}

// AfterPacketDeleted is called every time the underlying IBC packet is deleted.
// We only want to delete the demand order when the underlying packet is deleted to not
// break the invariant that the demand order is always in sync with the underlying packet.
func (d delayedAckHooks) AfterPacketDeleted(ctx sdk.Context, rollappPacket *commontypes.RollappPacket) {
	// Get the demand order from the packet key. The initial demand order was built when
	// the packet was created, hence with PENDING status.
	rollappPacket.Status = commontypes.Status_PENDING
	packetKey := rollappPacket.RollappPacketKey()
	demandOrderID := types.BuildDemandIDFromPacketKey(string(packetKey))

	statuses := []commontypes.Status{commontypes.Status_PENDING, commontypes.Status_FINALIZED}
	for _, status := range statuses {
		d.deleteDemandOrder(ctx, status, demandOrderID)

		if err := uevent.EmitTypedEvent(ctx, &types.EventDemandOrderDeleted{
			OrderId:      demandOrderID,
			PacketKey:    string(packetKey),
			PacketStatus: status.String(),
			RollappId:    rollappPacket.RollappId,
			PacketType:   rollappPacket.Type.String(),
		}); err != nil {
			d.Logger(ctx).Error("emit event", "error", err)
		}
	}
}

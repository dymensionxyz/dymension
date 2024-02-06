package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayeacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	types "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
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
func (d delayedAckHooks) AfterPacketStatusUpdated(ctx sdk.Context, packet *delayeacktypes.RollappPacket,
	oldPacketKey string, newPacketKey string) error {
	// Get the demand order from the old packet keyxx
	demandOrderID := types.BuildDemandIDFromPacketKey(oldPacketKey)
	demandOrder := d.GetDemandOrder(ctx, demandOrderID)
	// If no demand order was found, return
	if demandOrder == nil {
		return nil
	}
	// Update the demand order tracking packet key
	demandOrder.TrackingPacketKey = newPacketKey
	// Update the demand order status according to the underlying packet status
	d.UpdateDemandOrderWithStatus(ctx, demandOrder, packet.Status)

	return nil
}

/* -------------------------------------------------------------------------- */
/*                                 epoch hooks                                */
/* -------------------------------------------------------------------------- */
var _ epochstypes.EpochHooks = epochHooks{}

type epochHooks struct {
	Keeper
}

func (k Keeper) GetEpochHooks() epochstypes.EpochHooks {
	return epochHooks{
		Keeper: k,
	}
}

// BeforeEpochStart is the epoch start hook.
func (e epochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
// We want to clean up the demand orders that are with underlying packet status which are finalized.
func (e epochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier != e.GetParams(ctx).EpochIdentifier {
		return nil
	}
	// Get all demand orders with status FINALIZED
	demandOrders := e.ListDemandOrdersByStatus(ctx, commontypes.Status_FINALIZED)
	// Iterate over all demand orders
	for _, demandOrder := range demandOrders {
		e.deleteDemandOrder(ctx, demandOrder)
	}
	return nil
}

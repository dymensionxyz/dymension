package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayeacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	types "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
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
func (d delayedAckHooks) AfterPacketStatusUpdated(ctx sdk.Context, packet *commontypes.RollappPacket,
	oldPacketKey string, newPacketKey string) error {
	// Get the demand order from the old packet key
	demandOrderID := types.BuildDemandIDFromPacketKey(oldPacketKey)
	demandOrder, err := d.GetDemandOrder(ctx, demandOrderID)
	if err != nil {
		// If demand order does not exist, then we don't need to do anything
		if errors.Is(err, types.ErrDemandOrderDoesNotExist) {
			return nil
		}
		return err
	}
	// Update the demand order tracking packet key
	demandOrder.TrackingPacketKey = newPacketKey
	// Update the demand order status according to the underlying packet status
	_, err = d.UpdateDemandOrderWithStatus(ctx, demandOrder, packet.Status)
	if err != nil {
		return err
	}

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
	demandOrders, err := e.ListDemandOrdersByStatus(ctx, commontypes.Status_FINALIZED)
	if err != nil {
		return err
	}
	// Iterate over all demand orders
	for _, demandOrder := range demandOrders {
		wrapFunc := func(ctx sdk.Context) error {
			return e.deleteDemandOrder(ctx, demandOrder)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFunc)
		if err != nil {
			e.Keeper.Logger(ctx).Error("Error deleting demand order", "orderID", demandOrder.Id, "error", err.Error())
		}
	}
	return nil
}

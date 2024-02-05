package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
)

/* -------------------------------------------------------------------------- */
/*                                 eIBC Hooks                                 */
/* -------------------------------------------------------------------------- */

var _ eibctypes.EIBCHooks = eibcHooks{}

type eibcHooks struct {
	eibctypes.BaseEIBCHook
	Keeper
}

func (k Keeper) GetEIBCHooks() eibctypes.EIBCHooks {
	return eibcHooks{
		BaseEIBCHook: eibctypes.BaseEIBCHook{},
		Keeper:       k,
	}
}

// AfterDemandOrderFulfilled is called every time a demand order is fulfilled.
// Once it is fulfilled the underlying packet recipient should be updated to the fullfiller.
func (k eibcHooks) AfterDemandOrderFulfilled(ctx sdk.Context, demandOrder *eibctypes.DemandOrder, fulfillerAddress string) error {
	err := k.UpdateRollappPacketRecipient(ctx, demandOrder.TrackingPacketKey, fulfillerAddress)
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
	// Get all rollapp packets with status FINALIZED and delete them
	finalizedRollappPackets := e.ListRollappPacketsByStatus(ctx, commontypes.Status_FINALIZED, 0)
	for _, finalizedPacket := range finalizedRollappPackets {
		finalizedPacketCopy := finalizedPacket
		e.deleteRollappPacket(ctx, &finalizedPacketCopy)
	}
	return nil
}

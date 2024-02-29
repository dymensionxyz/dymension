package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		hooks      types.EIBCHooks
		paramstore paramtypes.Subspace
		types.AccountKeeper
		types.BankKeeper
		types.DelayedAckKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	delayedAckKeeper types.DelayedAckKeeper,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		paramstore:       ps,
		AccountKeeper:    accountKeeper,
		BankKeeper:       bankKeeper,
		DelayedAckKeeper: delayedAckKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetDemandOrder(ctx sdk.Context, order *types.DemandOrder) error {
	store := ctx.KVStore(k.storeKey)
	demandOrderKey, err := types.GetDemandOrderKey(order.TrackingPacketStatus, order.Id)
	if err != nil {
		return err
	}
	data, err := k.cdc.Marshal(order)
	if err != nil {
		return err
	}
	store.Set(demandOrderKey, data)

	// Emit events
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeyId, order.Id),
		sdk.NewAttribute(types.AttributeKeyPrice, order.Price.String()),
		sdk.NewAttribute(types.AttributeKeyFee, order.Fee.String()),
		sdk.NewAttribute(types.AttributeKeyIsFullfilled, strconv.FormatBool(order.IsFullfilled)),
		sdk.NewAttribute(types.AttributeKeyPacketStatus, order.TrackingPacketStatus.String()),
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEIBC,
			eventAttributes...,
		),
	)

	return nil

}

func (k Keeper) deleteDemandOrder(ctx sdk.Context, order *types.DemandOrder) error {
	store := ctx.KVStore(k.storeKey)
	demandOrderKey, err := types.GetDemandOrderKey(order.TrackingPacketStatus, order.Id)
	if err != nil {
		return err
	}
	store.Delete(demandOrderKey)
	return nil
}

// UpdateDemandOrderWithStatus deletes the current demand order and creates a new one with and updated packet status under a new key.
// Updating the status should be called only with this method as it effects the key of the packet.
// The assumption is that the passed demand order packet status field is not updated directly.
func (k *Keeper) UpdateDemandOrderWithStatus(ctx sdk.Context, demandOrder *types.DemandOrder, newStatus commontypes.Status) (*types.DemandOrder, error) {
	err := k.deleteDemandOrder(ctx, demandOrder)
	if err != nil {
		return nil, err
	}
	demandOrder.TrackingPacketStatus = newStatus
	err = k.SetDemandOrder(ctx, demandOrder)
	if err != nil {
		return nil, err
	}

	return demandOrder, nil
}

// This should be called only once per order.
func (k Keeper) FullfillOrder(ctx sdk.Context, order *types.DemandOrder, fulfillerAddress sdk.AccAddress) error {
	order.IsFullfilled = true
	err := k.SetDemandOrder(ctx, order)
	if err != nil {
		return err
	}
	// Call hooks if fulfilled. This hook should be called only once per fulfilment.
	err = k.hooks.AfterDemandOrderFulfilled(ctx, order, fulfillerAddress.String())
	if err != nil {
		return err
	}
	return nil
}

// GetDemandOrder returns the demand order with the given id and status.
func (k Keeper) GetDemandOrder(ctx sdk.Context, status commontypes.Status, id string) (*types.DemandOrder, error) {
	store := ctx.KVStore(k.storeKey)
	demandOrderKey, err := types.GetDemandOrderKey(status, id)
	if err != nil {
		return nil, err
	}
	bz := store.Get(demandOrderKey)
	if bz == nil {
		return nil, types.ErrDemandOrderDoesNotExist
	}
	var order types.DemandOrder
	err = k.cdc.Unmarshal(bz, &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ListAllDemandOrders returns all demand orders.
func (k Keeper) ListAllDemandOrders(
	ctx sdk.Context,
) (list []*types.DemandOrder, err error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.AllDemandOrdersKeyPrefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.DemandOrder
		wrapFn := func(ctx sdk.Context) error {
			return k.cdc.Unmarshal(iterator.Value(), &val)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("error unmarshalling demand order", "error", err.Error())
			continue
		}
		list = append(list, &val)
	}

	return list, nil
}

func (k Keeper) ListDemandOrdersByStatus(ctx sdk.Context, status commontypes.Status) (list []*types.DemandOrder, err error) {
	store := ctx.KVStore(k.storeKey)
	var statusPrefix []byte

	switch status {
	case commontypes.Status_PENDING:
		statusPrefix = types.PendingDemandOrderKeyPrefix
	case commontypes.Status_FINALIZED:
		statusPrefix = types.FinalizedDemandOrderKeyPrefix
	case commontypes.Status_REVERTED:
		statusPrefix = types.RevertedDemandOrderKeyPrefix
	default:
		return nil, fmt.Errorf("invalid packet status: %s", status)
	}

	iterator := sdk.KVStorePrefixIterator(store, statusPrefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.DemandOrder
		wrapFn := func(ctx sdk.Context) error {
			return k.cdc.Unmarshal(iterator.Value(), &val)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("error unmarshalling demand order", "error", err.Error())
			continue
		}
		list = append(list, &val)
	}

	return list, nil
}

/* -------------------------------------------------------------------------- */
/*                               Hooks handling                               */
/* -------------------------------------------------------------------------- */

func (k *Keeper) SetHooks(hooks types.EIBCHooks) {
	if k.hooks != nil {
		panic("EIBCHooks already set")
	}
	k.hooks = hooks
}

func (k *Keeper) GetHooks() types.EIBCHooks {
	return k.hooks
}

/* -------------------------------------------------------------------------- */
/*                                 Set Keepers                                */
/* -------------------------------------------------------------------------- */

// SetDelayedAckKeeper sets the delayedack keeper.
// must be called when initializing the keeper.
func (k *Keeper) SetDelayedAckKeeper(delayedAckKeeper types.DelayedAckKeeper) {
	k.DelayedAckKeeper = delayedAckKeeper
}

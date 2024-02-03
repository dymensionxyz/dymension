package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
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
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetDemandOrder(ctx sdk.Context, order *types.DemandOrder) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DemandOrderKeyPrefix))
	//TODO: FIXME -  Need to write a method of update demand order with status
	store.Set(types.GetDemandOrderKey(order.TrackingPacketStatus.String(), order.Id), k.cdc.MustMarshal(order))

	// Emit events
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeyPacketKey, order.Id),
		sdk.NewAttribute(types.AttributeKeyPrice, order.Price),
		sdk.NewAttribute(types.AttributeKeyFee, order.Fee),
		sdk.NewAttribute(types.AttributeKeyDenom, order.Denom),
		sdk.NewAttribute(types.AttributeKeyIsFullfilled, strconv.FormatBool(order.IsFullfilled)),
		sdk.NewAttribute(types.AttributeKeyPacketStatus, order.TrackingPacketStatus.String()),
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEIBC,
			eventAttributes...,
		),
	)

}

// UpdateDemandOrderWithStatus deletes the current demand order and creates a new one with and updated packet status under a new key.
// Updating the status should be called only with this method as it effects the key of the packet.
// The assumption is that the passed demand order packet status field is not updated directly.
func (k *Keeper) UpdateDemandOrderWithStatus(ctx sdk.Context, demandOrder *types.DemandOrder, newStatus commontypes.Status) *types.DemandOrder {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DemandOrderKeyPrefix))

	// Delete the old demand order
	oldKey := types.GetDemandOrderKey(demandOrder.TrackingPacketStatus.String(), demandOrder.Id)
	store.Delete(oldKey)

	// Update the demand order
	demandOrder.TrackingPacketStatus = newStatus

	// Create a new demand with the updated status
	k.SetDemandOrder(ctx, demandOrder)

	return demandOrder
}

// This should be called only once per order.
func (k Keeper) FullfillOrder(ctx sdk.Context, order *types.DemandOrder, fulfillerAddress sdk.AccAddress) {
	order.IsFullfilled = true
	k.SetDemandOrder(ctx, order)
	// Call hooks if fulfilled. This hook should be called only once per fulfilment.
	err := k.hooks.AfterDemandOrderFulfilled(ctx, order, fulfillerAddress.String())
	if err != nil {
		panic("Error calling AfterDemandOrderFulfilled hook: " + err.Error())
	}
}

func (k Keeper) GetDemandOrder(ctx sdk.Context, id string) *types.DemandOrder {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DemandOrderKeyPrefix))
	// Im only interested in the pending orders
	bz := store.Get(types.GetDemandOrderKey(commontypes.Status_PENDING.String(), id))
	if bz == nil {
		return nil
	}
	var order types.DemandOrder
	k.cdc.MustUnmarshal(bz, &order)
	return &order
}

// ListAllDemandOrders returns all demand orders. Shouldn't be exposed to the client.
func (k Keeper) ListAllDemandOrders(
	ctx sdk.Context,
) (list []types.DemandOrder) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DemandOrderKeyPrefix))

	// Build the prefix which is composed of the rollappID and the status
	var prefix []byte

	// Iterate over the range from lastProofHeight to proofHeight
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.DemandOrder
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return list
}

func (k Keeper) ListDemandOrdersByStatus(ctx sdk.Context, status commontypes.Status) (list []*types.DemandOrder) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DemandOrderKeyPrefix))

	// Build the prefix which is composed of  the status
	var prefix []byte

	prefix = append(prefix, []byte(status.String())...)
	prefix = append(prefix, []byte("/")...)

	// Iterate over the range from lastProofHeight to proofHeight
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.DemandOrder
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, &val)
	}

	return list
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

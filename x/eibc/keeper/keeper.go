package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

type (
	Keeper struct {
		cdc       codec.BinaryCodec
		storeKey  storetypes.StoreKey
		memKey    storetypes.StoreKey
		hooks     types.EIBCHooks
		ak        types.AccountKeeper
		bk        types.BankKeeper
		dack      types.DelayedAckKeeper
		rk        types.RollappKeeper
		Schema    collections.Schema
		LPs       LPs
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	delayedAckKeeper types.DelayedAckKeeper,
	rk types.RollappKeeper,
	authority string,
) *Keeper {
	service := collcompat.NewKVStoreService(storeKey)
	sb := collections.NewSchemaBuilder(service)
	lps := makeLPsStore(sb, cdc)

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		memKey:    memKey,
		ak:        accountKeeper,
		bk:        bankKeeper,
		dack:      delayedAckKeeper,
		rk:        rk,
		Schema:    schema,
		LPs:       lps,
		authority: authority,
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

	return nil
}

func (k Keeper) deleteDemandOrder(ctx sdk.Context, status commontypes.Status, orderID string) {
	store := ctx.KVStore(k.storeKey)
	// we can skip error check, the status is known, if key is not valid, order will not be deleted anyway
	demandOrderKey, _ := types.GetDemandOrderKey(status, orderID)
	store.Delete(demandOrderKey)
}

// UpdateDemandOrderWithStatus deletes the current demand order and creates a new one with and updated packet status under a new key.
// Updating the status should be called only with this method as it effects the key of the packet.
// The assumption is that the passed demand order packet status field is not updated directly.
func (k *Keeper) UpdateDemandOrderWithStatus(ctx sdk.Context, demandOrder *types.DemandOrder, newStatus commontypes.Status) (*types.DemandOrder, error) {
	k.deleteDemandOrder(ctx, demandOrder.TrackingPacketStatus, demandOrder.Id)

	demandOrder.TrackingPacketStatus = newStatus
	err := k.SetDemandOrder(ctx, demandOrder)
	if err != nil {
		return nil, err
	}

	if err = uevent.EmitTypedEvent(ctx, types.GetPacketStatusUpdatedEvent(demandOrder)); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return demandOrder, nil
}

func (k Keeper) PendingOrderByPacket(ctx sdk.Context, p *commontypes.RollappPacket) (*types.DemandOrder, error) {
	key := p.RollappPacketKey()
	id := types.BuildDemandIDFromPacketKey(string(key))
	return k.GetDemandOrder(ctx, commontypes.Status_PENDING, id)
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

func (k Keeper) GetOutstandingOrder(ctx sdk.Context, orderId string) (*types.DemandOrder, error) {
	// Check that the order exists in status PENDING
	demandOrder, err := k.GetDemandOrder(ctx, commontypes.Status_PENDING, orderId)
	if err != nil {
		return nil, err
	}
	if err := demandOrder.ValidateOrderIsOutstanding(); err != nil {
		return nil, err
	}

	// TODO: would be nice if the demand order already has the proofHeight, so we don't have to fetch the packet
	packet, err := k.dack.GetRollappPacket(ctx, demandOrder.TrackingPacketKey)
	if err != nil {
		return nil, err
	}

	// No error means the order is due to be finalized,
	// in which case the order is not outstanding anymore
	if err = k.dack.VerifyHeightFinalized(ctx, demandOrder.RollappId, packet.ProofHeight); err == nil {
		return nil, types.ErrDemandOrderInactive
	}

	return demandOrder, nil
}

// ListAllDemandOrders returns all demand orders.
func (k Keeper) ListAllDemandOrders(
	ctx sdk.Context,
) (list []*types.DemandOrder, err error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.AllDemandOrdersKeyPrefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.DemandOrder
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, &val)
	}

	return list, nil
}

func (k Keeper) ListDemandOrdersByStatus(ctx sdk.Context, status commontypes.Status, limit int, opts ...filterOption) (list []*types.DemandOrder, err error) {
	store := ctx.KVStore(k.storeKey)

	var statusPrefix []byte
	switch status {
	case commontypes.Status_PENDING:
		statusPrefix = types.PendingDemandOrderKeyPrefix
	case commontypes.Status_FINALIZED:
		statusPrefix = types.FinalizedDemandOrderKeyPrefix
	default:
		return nil, fmt.Errorf("invalid packet status: %s", status)
	}

	iterator := storetypes.KVStorePrefixIterator(store, statusPrefix)
	defer iterator.Close() // nolint: errcheck

outer:
	for ; iterator.Valid(); iterator.Next() {
		if limit > 0 && len(list) >= limit {
			break
		}
		var val types.DemandOrder
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		for _, opt := range opts {
			if !opt(val) {
				continue outer
			}
		}
		list = append(list, &val)
	}

	return list, nil
}

func (k Keeper) ListDemandOrdersByStatusPaginated(
	ctx sdk.Context,
	status commontypes.Status,
	pageReq *query.PageRequest,
	opts ...filterOption,
) (list []*types.DemandOrder, pageResp *query.PageResponse, err error) {
	store := ctx.KVStore(k.storeKey)

	var statusPrefix []byte
	switch status {
	case commontypes.Status_PENDING:
		statusPrefix = types.PendingDemandOrderKeyPrefix
	case commontypes.Status_FINALIZED:
		statusPrefix = types.FinalizedDemandOrderKeyPrefix
	default:
		err = fmt.Errorf("invalid demand order status: %s", status)
		return
	}

	prefixStore := prefix.NewStore(store, statusPrefix)

	if pageReq == nil {
		pageReq = &query.PageRequest{}
	}

	pageResp, err = query.Paginate(prefixStore, pageReq, func(key []byte, value []byte) error {
		var val types.DemandOrder
		if err := k.cdc.Unmarshal(value, &val); err != nil {
			return err
		}
		for _, opt := range opts {
			if !opt(val) {
				return nil
			}
		}
		list = append(list, &val)
		return nil
	})

	return
}

func (k Keeper) ensureAccount(ctx sdk.Context, address sdk.AccAddress) error {
	account := k.ak.GetAccount(ctx, address)
	if account == nil {
		return errorsmod.Wrapf(types.ErrAccountDoesNotExist, "address: %s", address)
	}
	return nil
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
	k.dack = delayedAckKeeper
}

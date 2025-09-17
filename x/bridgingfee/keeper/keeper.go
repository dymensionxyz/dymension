package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	postdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

type Keeper struct {
	feeHooks         collections.Map[uint64, types.HLFeeHook]
	aggregationHooks collections.Map[uint64, types.AggregationHook]

	coreKeeper types.CoreKeeper
	bankKeeper types.BankKeeper
	warpQuery  types.WarpQuery

	schema collections.Schema
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	coreKeeper types.CoreKeeper,
	warpQuery types.WarpQuery,
	bankKeeper types.BankKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		feeHooks: collections.NewMap(
			sb,
			types.KeyFeeHooks,
			"fee_hooks",
			collections.Uint64Key,
			collcompat.ProtoValue[types.HLFeeHook](cdc),
		),
		aggregationHooks: collections.NewMap(
			sb,
			types.KeyAggregationHooks,
			"aggregation_hooks",
			collections.Uint64Key,
			collcompat.ProtoValue[types.AggregationHook](cdc),
		),
		bankKeeper: bankKeeper,
		warpQuery:  warpQuery,
	}

	k.SetCoreKeeper(coreKeeper)

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.schema = schema

	return k
}

func (k *Keeper) SetCoreKeeper(coreKeeper types.CoreKeeper) {
	if k.coreKeeper != nil {
		panic("core keeper already set")
	}

	k.coreKeeper = coreKeeper

	router := coreKeeper.PostDispatchRouter()
	router.RegisterModule(postdispatchtypes.POST_DISPATCH_HOOK_TYPE_PROTOCOL_FEE, FeeHookHandler{*k})
	router.RegisterModule(postdispatchtypes.POST_DISPATCH_HOOK_TYPE_AGGREGATION, AggregationHookHandler{*k})
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// CreateFeeHook creates a new fee hook (business logic)
func (k Keeper) CreateFeeHook(ctx context.Context, msg *types.MsgCreateBridgingFeeHook) (hyputil.HexAddress, error) {
	// Get next hook ID
	hexAddr, err := k.coreKeeper.PostDispatchRouter().GetNextSequence(ctx, postdispatchtypes.POST_DISPATCH_HOOK_TYPE_PROTOCOL_FEE)
	if err != nil {
		return hyputil.HexAddress{}, fmt.Errorf("get next hook id: %w", err)
	}

	// Create and save the fee hook
	hook := types.HLFeeHook{
		Id:    hexAddr,
		Owner: msg.Owner,
		Fees:  msg.Fees,
	}

	if err := k.feeHooks.Set(ctx, hexAddr.GetInternalId(), hook); err != nil {
		return hyputil.HexAddress{}, fmt.Errorf("save fee hook: %w", err)
	}

	// Emit event
	err = uevent.EmitTypedEvent(sdk.UnwrapSDKContext(ctx), &types.EventFeeHookCreated{
		HookId: hexAddr,
		Owner:  msg.Owner,
	})
	if err != nil {
		return hyputil.HexAddress{}, fmt.Errorf("emit event: %w", err)
	}

	return hexAddr, nil
}

// UpdateFeeHook updates an existing fee hook (business logic)
func (k Keeper) UpdateFeeHook(ctx context.Context, msg *types.MsgSetBridgingFeeHook) error {
	// Get existing hook
	hookId := msg.Id.GetInternalId()
	hook, err := k.feeHooks.Get(ctx, hookId)
	if err != nil {
		return fmt.Errorf("get fee hook: %w", err)
	}

	// Check ownership
	if hook.Owner != msg.Owner {
		return gerrc.ErrPermissionDenied.Wrap("not the owner of the hook")
	}

	// Update hook
	hook.Fees = msg.Fees

	// Handle ownership transfer
	if msg.RenounceOwnership {
		hook.Owner = ""
	} else if msg.NewOwner != "" {
		hook.Owner = msg.NewOwner
	}

	if err := k.feeHooks.Set(ctx, hookId, hook); err != nil {
		return fmt.Errorf("save fee hook: %w", err)
	}

	// Emit event
	err = uevent.EmitTypedEvent(sdk.UnwrapSDKContext(ctx), &types.EventFeeHookUpdated{
		HookId: msg.Id,
		Owner:  hook.Owner,
	})
	if err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

// CreateAggregationHook creates a new aggregation hook (business logic)
func (k Keeper) CreateAggregationHook(ctx context.Context, msg *types.MsgCreateAggregationHook) (hyputil.HexAddress, error) {
	// Get next hook ID
	hexAddr, err := k.coreKeeper.PostDispatchRouter().GetNextSequence(ctx, postdispatchtypes.POST_DISPATCH_HOOK_TYPE_AGGREGATION)
	if err != nil {
		return hyputil.HexAddress{}, fmt.Errorf("get next hook id: %w", err)
	}

	// Create and save the aggregation hook
	hook := types.AggregationHook{
		Id:      hexAddr,
		Owner:   msg.Owner,
		HookIds: msg.HookIds,
	}

	if err := k.aggregationHooks.Set(ctx, hexAddr.GetInternalId(), hook); err != nil {
		return hyputil.HexAddress{}, fmt.Errorf("save aggregation hook: %w", err)
	}

	// Emit event
	err = uevent.EmitTypedEvent(sdk.UnwrapSDKContext(ctx), &types.EventAggregationHookCreated{
		HookId: hexAddr,
		Owner:  msg.Owner,
	})
	if err != nil {
		return hyputil.HexAddress{}, fmt.Errorf("emit event: %w", err)
	}

	return hexAddr, nil
}

// UpdateAggregationHook updates an existing aggregation hook (business logic)
func (k Keeper) UpdateAggregationHook(ctx context.Context, msg *types.MsgSetAggregationHook) error {
	// Get existing hook
	hookId := msg.Id.GetInternalId()
	hook, err := k.aggregationHooks.Get(ctx, hookId)
	if err != nil {
		return fmt.Errorf("get aggregation hook: %w", err)
	}

	// Check ownership
	if hook.Owner != msg.Owner {
		return gerrc.ErrPermissionDenied.Wrap("not the owner of the hook")
	}

	// Update hook
	hook.HookIds = msg.HookIds

	// Handle ownership transfer
	if msg.RenounceOwnership {
		hook.Owner = ""
	} else if msg.NewOwner != "" {
		hook.Owner = msg.NewOwner
	}

	if err := k.aggregationHooks.Set(ctx, hookId, hook); err != nil {
		return fmt.Errorf("save aggregation hook: %w", err)
	}

	// Emit event
	err = uevent.EmitTypedEvent(sdk.UnwrapSDKContext(ctx), &types.EventAggregationHookUpdated{
		HookId: msg.Id,
		Owner:  hook.Owner,
	})
	if err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

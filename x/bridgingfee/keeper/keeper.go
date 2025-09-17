package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	postdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
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

	coreKeeper   types.CoreKeeper
	bankKeeper   types.BankKeeper
	txFeesKeeper types.TxFeesKeeper
	warpQuery    types.WarpQuery

	schema collections.Schema
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	coreKeeper types.CoreKeeper,
	bankKeeper types.BankKeeper,
	txFeesKeeper types.TxFeesKeeper,
	warpQuery types.WarpQuery,
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
		coreKeeper:   coreKeeper,
		bankKeeper:   bankKeeper,
		txFeesKeeper: txFeesKeeper,
		warpQuery:    warpQuery,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.schema = schema

	// Register HL post-dispatch hooks in x/core
	router := coreKeeper.PostDispatchRouter()
	router.RegisterModule(postdispatchtypes.POST_DISPATCH_HOOK_TYPE_PROTOCOL_FEE, FeeHookHandler{k})
	router.RegisterModule(postdispatchtypes.POST_DISPATCH_HOOK_TYPE_AGGREGATION, AggregationHookHandler{k})

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// CreateFeeHook creates a new fee hook (business logic)
func (k Keeper) CreateFeeHook(ctx context.Context, msg *types.MsgCreateBridgingFeeHook) (hyputil.HexAddress, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return hyputil.HexAddress{}, fmt.Errorf("invalid msg: %w", err)
	}

	// Verify all tokenIDs exist
	for _, fee := range msg.Fees {
		_, err := k.warpQuery.Token(ctx, &warptypes.QueryTokenRequest{Id: fee.TokenID})
		if err != nil {
			return hyputil.HexAddress{}, fmt.Errorf("token %s does not exist: %w", fee.TokenID, err)
		}
	}

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
	err := msg.ValidateBasic()
	if err != nil {
		return fmt.Errorf("invalid msg: %w", err)
	}

	// Verify all tokenIDs exist
	for _, fee := range msg.Fees {
		_, err := k.warpQuery.Token(ctx, &warptypes.QueryTokenRequest{Id: fee.TokenID})
		if err != nil {
			return fmt.Errorf("token %s does not exist: %w", fee.TokenID, err)
		}
	}

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
	err := msg.ValidateBasic()
	if err != nil {
		return hyputil.HexAddress{}, fmt.Errorf("invalid msg: %w", err)
	}

	// Verify all referenced hooks exist
	for _, hookId := range msg.HookIds {
		_, err := k.coreKeeper.PostDispatchRouter().GetModule(hookId)
		if err != nil {
			return hyputil.HexAddress{}, fmt.Errorf("get hook with id: %s: %w", hookId, err)
		}
	}

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
	err := msg.ValidateBasic()
	if err != nil {
		return fmt.Errorf("invalid msg: %w", err)
	}

	// Verify all referenced hooks exist
	for _, hookId := range msg.HookIds {
		_, err := k.coreKeeper.PostDispatchRouter().GetModule(hookId)
		if err != nil {
			return fmt.Errorf("get hook with id: %s: %w", hookId, err)
		}
	}

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

	err = uevent.EmitTypedEvent(sdk.UnwrapSDKContext(ctx), &types.EventAggregationHookUpdated{
		HookId: msg.Id,
		Owner:  hook.Owner,
	})
	if err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

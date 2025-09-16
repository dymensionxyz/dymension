package keeper

import (
	"context"
	"fmt"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	postdispathkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/keeper"
	postdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type HLFeeHook struct {
	Keeper
}

func (k HLFeeHook) Exists(ctx context.Context, hookId util.HexAddress) (bool, error) {
	ok, err := k.feeHooks.Has(ctx, hookId.GetInternalId())
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (k HLFeeHook) PostDispatch(ctx context.Context, mailboxId, hookId util.HexAddress, metadata util.StandardHookMetadata, message util.HyperlaneMessage, maxFee sdk.Coins) (sdk.Coins, error) {
	//TODO implement me
	panic("implement me")
}

func (k HLFeeHook) QuoteDispatch(ctx context.Context, mailboxId, hookId util.HexAddress, metadata util.StandardHookMetadata, message util.HyperlaneMessage) (sdk.Coins, error) {
	//TODO implement me
	panic("implement me")
}

func (k HLFeeHook) HookType() uint8 {
	//TODO implement me
	panic("implement me")
}

type AggregatedHook struct {
	Keeper
	hookIDs []util.HexAddress
}

func (h AggregatedHook) Exists(ctx context.Context, hookId util.HexAddress) (bool, error) {
	return k.aggregationHooks.Has(ctx, hookId.GetInternalId())
}

func (h AggregatedHook) HookType() uint8 {
	return postdispatchtypes.POST_DISPATCH_HOOK_TYPE_AGGREGATION
}

func (h AggregatedHook) PostDispatch(ctx context.Context, mailboxId, hookId util.HexAddress, metadata util.StandardHookMetadata, message util.HyperlaneMessage, maxFee sdk.Coins) (sdk.Coins, error) {
	hook := h.aggregationHooks.Get(ctx, hookId)
	totalCharged := sdk.NewCoins()
	remaining := maxFee

	pdRouter := h.coreKeeper.PostDispatchRouter()
	for _, childHookId := range hook.hookIds {
		pdModule := pdRouter.GetModule(childHookId)

		chargedFee := pdModule.PostDispatch(ctx, mailboxId, childHookId, metadata, message, remaining)

		totalCharged = totalCharged.Add(chargedFee...)
		remaining = remaining.Sub(chargedFee...)

		if remaining.IsAnyNegative() {
			return err(exceeded max fee)
		}
	}

	return totalCharged, nil
}

func (h AggregatedHook) QuoteDispatch(ctx context.Context, mailboxId, hookId util.HexAddress, metadata util.StandardHookMetadata, message util.HyperlaneMessage) (sdk.Coins, error) {
	hook := h.aggregationHooks.Get(ctx, hookId)
	totalQuote := sdk.NewCoins()

	pdRouter := h.coreKeeper.PostDispatchRouter()
	for _, childHookId := range hook.hookIds {
		pdModule := pdRouter.GetModule(childHookId)

		quote := pdModule.QuoteDispatch(ctx, mailboxId, childHookId, metadata, message)

		totalQuote = totalQuote.Add(quote...)
	}

	return totalQuote, nil
}

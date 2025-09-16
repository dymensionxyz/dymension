package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	postdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
)

// AggregationHookHandler implements the aggregation post-dispatch hook
type AggregationHookHandler struct {
	k Keeper
}

var _ hyputil.PostDispatchModule = AggregationHookHandler{}

func (a AggregationHookHandler) Exists(ctx context.Context, hookId hyputil.HexAddress) (bool, error) {
	has, err := a.k.aggregationHooks.Has(ctx, hookId.GetInternalId())
	if err != nil {
		return false, err
	}
	return has, nil
}

func (a AggregationHookHandler) HookType() uint8 {
	return postdispatchtypes.POST_DISPATCH_HOOK_TYPE_AGGREGATION
}

// PostDispatch executes multiple hooks in sequence
func (a AggregationHookHandler) PostDispatch(ctx context.Context, mailboxId, hookId hyputil.HexAddress, metadata hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage, maxFee sdk.Coins) (sdk.Coins, error) {
	hook, err := a.k.aggregationHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return nil, errorsmod.Wrap(err, "get aggregation hook")
	}

	totalCharged := sdk.NewCoins()
	remaining := maxFee

	pdRouter := a.k.coreKeeper.PostDispatchRouter()
	for _, subHookId := range hook.HookIds {
		pdModule, err := pdRouter.GetModule(subHookId)
		if err != nil {
			return nil, fmt.Errorf("get post-dispatch module for %s: %w", subHookId.String(), err)
		}
		chargedFee, err := (*pdModule).PostDispatch(ctx, mailboxId, subHookId, metadata, message, remaining)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "execute sub-hook %s", subHookId.String())
		}

		totalCharged = totalCharged.Add(chargedFee...)
		remaining = remaining.Sub(chargedFee...)

		if remaining.IsAnyNegative() {
			return nil, fmt.Errorf("fee collection exceeded max fee")
		}
	}

	return totalCharged, nil
}

// QuoteDispatch returns the total required fees for all sub-hooks
func (a AggregationHookHandler) QuoteDispatch(ctx context.Context, mailboxId, hookId hyputil.HexAddress, metadata hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage) (sdk.Coins, error) {
	hook, err := a.k.aggregationHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return nil, errorsmod.Wrap(err, "get aggregation hook")
	}

	totalQuote := sdk.NewCoins()

	pdRouter := a.k.coreKeeper.PostDispatchRouter()
	for _, subHookId := range hook.HookIds {
		pdModule, err := pdRouter.GetModule(subHookId)
		if err != nil {
			return nil, fmt.Errorf("get post-dispatch module for %s: %w", subHookId.String(), err)
		}

		quote, err := (*pdModule).QuoteDispatch(ctx, mailboxId, subHookId, metadata, message)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "quote sub-hook %s", subHookId.String())
		}

		totalQuote = totalQuote.Add(quote...)
	}

	return totalQuote, nil
}

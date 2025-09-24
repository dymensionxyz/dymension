package keeper

import (
	"context"
	"fmt"

	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

// AggregationHookHandler is a Hyperlane post-dispatch hook that combines multiple sub-hooks into a single hook.
// This allows chaining logics (fee collection, gas payment, merkle tree) to be executed sequentially.
// The hook will execute all sub-hooks in the order they are provided and will fail if any sub-hook fails.
type AggregationHookHandler struct {
	k Keeper
}

func NewAggregationHookHandler(k Keeper) AggregationHookHandler {
	return AggregationHookHandler{k: k}
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
	return types.PostDispatchHookDymAggregation
}

// PostDispatch executes multiple hooks in sequence. The method is atomic – all sub-hooks must be executed,
// otherwise reverted.
func (a AggregationHookHandler) PostDispatch(goCtx context.Context, mailboxId, hookId hyputil.HexAddress, metadata hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage, maxFee sdk.Coins) (sdk.Coins, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hook, err := a.k.aggregationHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return nil, fmt.Errorf("get aggregation hook: %w", err)
	}

	var (
		totalCharged = sdk.NewCoins()
		remaining    = maxFee
		pdRouter     = a.k.coreKeeper.PostDispatchRouter()
	)

	err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		for _, subHookId := range hook.HookIds {
			pdModule, err := pdRouter.GetModule(subHookId)
			if err != nil {
				return fmt.Errorf("get post-dispatch module for %s: %w", subHookId.String(), err)
			}
			chargedFee, err := (*pdModule).PostDispatch(ctx, mailboxId, subHookId, metadata, message, remaining)
			if err != nil {
				return fmt.Errorf("execute sub-hook %s: %w", subHookId.String(), err)
			}

			var negative bool
			remaining, negative = remaining.SafeSub(chargedFee...)
			totalCharged = totalCharged.Add(chargedFee...)

			if negative {
				return fmt.Errorf("fee collection exceeded max fee")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return totalCharged, nil
}

// QuoteDispatch returns the total quote after applying all sub-hooks
func (a AggregationHookHandler) QuoteDispatch(ctx context.Context, mailboxId, hookId hyputil.HexAddress, metadata hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage) (sdk.Coins, error) {
	hook, err := a.k.aggregationHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return nil, fmt.Errorf("get aggregation hook: %w", err)
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
			return nil, fmt.Errorf("quote sub-hook %s: %w", subHookId.String(), err)
		}

		totalQuote = totalQuote.Add(quote...)
	}

	return totalQuote, nil
}

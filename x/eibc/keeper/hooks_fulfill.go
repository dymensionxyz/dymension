package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	HookNameForward = "forward"
)

type FulfillHook interface {
	ValidateData(hookData []byte) error
	Run(ctx sdk.Context, order *types.DemandOrder, args fulfillArgs, hookData []byte) error
}

type FulfillHooks struct {
	hooks map[string]FulfillHook
}

func NewHooks() FulfillHooks {
	return FulfillHooks{
		hooks: make(map[string]FulfillHook),
	}
}

// assumed already passed validate basic
func (h FulfillHooks) validateFulfillHook(info types.FulfillHook) error {
	f, ok := h.hooks[info.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrap("hook")
	}
	return f.ValidateData(info.HookData)
}

func (h FulfillHooks) doFulfillHook(ctx sdk.Context, order *types.DemandOrder, args fulfillArgs) error {
	f, ok := h.hooks[order.FulfillHook.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrap("hook")
	}
	return f.Run(ctx, order, args, order.FulfillHook.HookData)
}

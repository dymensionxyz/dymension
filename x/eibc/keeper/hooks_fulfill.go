package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type FulfillHook interface {
	ValidateData(hookData []byte) error
	Run(ctx sdk.Context, order *types.DemandOrder,
		fundsSource sdk.AccAddress,
		newTransferRecipient sdk.AccAddress,
		fulfiller sdk.AccAddress,

		hookData []byte) error
}

type FulfillHooks map[string]FulfillHook

// assumed already passed validate basic
func (h FulfillHooks) validate(info types.FulfillHook) error {
	f, ok := h[info.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrapf("hook: name: %s", info.HookName)
	}
	return f.ValidateData(info.HookData)
}

func (h FulfillHooks) exec(ctx sdk.Context, order *types.DemandOrder, args fulfillArgs) error {
	f, ok := h[order.FulfillHook.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrap("hook")
	}
	return f.Run(ctx, order, args.FundsSource, args.NewTransferRecipient, args.Fulfiller, order.FulfillHook.HookData)
}

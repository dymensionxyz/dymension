package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/gogo/protobuf/proto"
)

const (
	HookNameForward = "forward"
)

// assumed already passed validate basic
func validateFulfillHook(info types.FulfillHook) error {
	switch info.HookName {
	case HookNameForward:
		return validForward(info.HookData)
	default:
		return gerrc.ErrInvalidArgument.Wrap("hook name")
	}
}

func validForward(data []byte) error {
	var d types.ForwardHook
	err := proto.Unmarshal(data, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	if err := d.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "validate forward hook")
	}
	return nil
}

func (k Keeper) doFulfillHook(ctx sdk.Context, order *types.DemandOrder) error {
	switch order.FulfillHook.HookName {
	case HookNameForward:
		return k.doForwardHook(ctx, order)
	default:
		return gerrc.ErrInvalidArgument.Wrap("hook name")
	}
}

func (k Keeper) doForwardHook(ctx sdk.Context, order *types.DemandOrder) error {
	var d types.ForwardHook
	err := proto.Unmarshal(order.FulfillHook.HookData, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	return nil
}

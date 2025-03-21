package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
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
		return fmt.Errorf("invalid hook name: %s", info.HookName)
	}
}

func validForward(data []byte) error {
	var d types.ForwardHook
	err := proto.Unmarshal(data, &d)
	if err != nil {
		return fmt.Errorf("unmarshal forward hook metadata: %w", err)
	}
	if err := d.ValidateBasic(); err != nil {
		return fmt.Errorf("validate forward hook metadata: %w", err)
	}
	return nil
}

func (k Keeper) doFulfillHook(ctx sdk.Context, order *types.DemandOrder) error {
	switch order.FulfillHook.HookName {
	case HookNameForward:
		return k.doForwardHook(ctx, order)
	default:
		return fmt.Errorf("invalid hook name: %s", order.FulfillHook.HookName)
	}
}

func (k Keeper) doForwardHook(ctx sdk.Context, order *types.DemandOrder) error {
	var d types.ForwardHook
	err := proto.Unmarshal(order.FulfillHook.HookData, &d)
	if err != nil {
		return fmt.Errorf("unmarshal forward hook metadata: %w", err)
	}
}

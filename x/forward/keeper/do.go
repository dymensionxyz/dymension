package keeper

import (
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

)

const (
	HookNameForward = "forward"
)

var _ eibckeeper.FulfillHook = Hook{}

func (k Keeper) Hook() Hook {
	return Hook{
		Keeper: &k,
	}
}

type Hook struct {
	*Keeper
}

func (h Hook) ValidateData(data []byte) error {
	return validForward(data)
}

func (h Hook) Run(ctx sdk.Context, order *eibctypes.DemandOrder,		fundsSource sdk.AccAddress,
		newTransferRecipient sdk.AccAddress,
		fulfiller sdk.AccAddress, hookData []byte) error {
	return h.doForwardHook(ctx, order, args.FundsSource, hookData)
}

func (h Hook) Name() string {
	return HookNameForward
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


func (k Keeper) doForwardHook(ctx sdk.Context, order *types.DemandOrder, fundsSource sdk.AccAddress, data ) error {
	var d types.ForwardHook
	err := proto.Unmarshal(order.FulfillHook.HookData, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
}
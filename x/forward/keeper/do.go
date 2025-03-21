package keeper

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
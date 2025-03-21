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

/*
What was I doing?
I need to make a hook type here that can be passed to the eibc keeper
The hook type needs to embed the keeper and call whatever methods we need to do here
*/

func (k Keeper) doForwardHook(ctx sdk.Context, order *types.DemandOrder, fundsSource sdk.AccAddress, data ) error {
	var d types.ForwardHook
	err := proto.Unmarshal(order.FulfillHook.HookData, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
}
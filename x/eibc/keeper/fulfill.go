package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	transfer "github.com/dymensionxyz/dymension/v3/x/transfer"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// basic i.e. not authorized
func (k Keeper) fulfillBasic(ctx sdk.Context,
	o *types.DemandOrder,
	fulfiller sdk.AccAddress,
) error {

	err := k.fulfill(ctx, o, transfer.EIBCFulfillArgs{
		FundsSource:          fulfiller,
		NewTransferRecipient: fulfiller,
		Fulfiller:            fulfiller,
	})

	if err != nil {
		return err
	}

	if err = uevent.EmitTypedEvent(ctx, o.GetFulfilledEvent()); err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

func (k Keeper) fulfill(ctx sdk.Context,
	o *types.DemandOrder,
	args transfer.EIBCFulfillArgs,
) error {
	if err := k.ensureAccount(ctx, args.FundsSource); err != nil {
		return errorsmod.Wrap(err, "ensure fulfiller account")
	}

	if o.CompletionHookCall != nil {
		err := k.transferHooks.OnFulfill(ctx, o, args)
		if err != nil {
			return errorsmod.Wrap(err, "do fulfill hook")
		}
	} else {
		// TODO: could make this an instance of a hook / type of hook, then instead of branching it would be one flow
		err := k.bk.SendCoins(ctx, args.FundsSource, o.GetRecipientBech32Address(), o.Price)
		if err != nil {
			return errorsmod.Wrap(err, "send coins")
		}
	}

	o.FulfillerAddress = args.Fulfiller.String()
	err := k.SetDemandOrder(ctx, o)
	if err != nil {
		return err
	}

	err = k.hooks.AfterDemandOrderFulfilled(ctx, o, args.NewTransferRecipient.String())
	if err != nil {
		return err
	}

	return nil
}

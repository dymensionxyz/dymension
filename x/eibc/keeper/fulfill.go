package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// basic i.e. not authorized
func (k Keeper) fulfillBasic(ctx sdk.Context,
	o *types.DemandOrder,
	fulfiller sdk.AccAddress,
) error {
	err := k.fulfill(ctx, o, fulfillArgs{
		FundsSource: fulfiller,
		Fulfiller:   fulfiller,
	})
	if err != nil {
		return err
	}

	if err = uevent.EmitTypedEvent(ctx, types.GetFulfilledEvent(o)); err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

type fulfillArgs struct {
	FundsSource sdk.AccAddress
	Fulfiller   sdk.AccAddress
}

func (k Keeper) fulfill(ctx sdk.Context,
	o *types.DemandOrder,
	args fulfillArgs,
) error {
	if err := k.ensureAccount(ctx, args.FundsSource); err != nil {
		return errorsmod.Wrap(err, "ensure fulfiller account")
	}

	err := k.bk.SendCoins(ctx, args.FundsSource, o.GetRecipientBech32Address(), o.Price)
	if err != nil {
		return errorsmod.Wrap(err, "send coins")
	}

	o.FulfillerAddress = args.Fulfiller.String()
	err = k.SetDemandOrder(ctx, o)
	if err != nil {
		return err
	}

	err = k.hooks.AfterDemandOrderFulfilled(ctx, o, args.FundsSource.String())
	if err != nil {
		return err
	}

	return nil
}

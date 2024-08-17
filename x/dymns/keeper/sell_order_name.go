package keeper

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// CompleteDymNameSellOrder completes the active sell order of the Dym-Name,
// give value to the previous owner, and transfer ownership to new owner.
//
// Sell-Order is considered completed when:
//   - There is a bid placed, and the Sell-Order has expired.
//   - There is a bid placed, and it matches the sell price.
func (k Keeper) CompleteDymNameSellOrder(ctx sdk.Context, name string) error {
	dymName := k.GetDymName(ctx, name)
	if dymName == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", name)
	}

	// here we don't check Dym-Name expiration, because it can not happen,
	// and there is a grace period for the owner to renew the Dym-Name in case bad things happen

	so := k.GetSellOrder(ctx, name, dymnstypes.TypeName)
	if so == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s", name)
	}

	if !so.HasFinishedAtCtx(ctx) {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "Sell-Order has not finished yet")
	}

	// the SO can be expired at this point,
	// in case the highest bid is lower than sell price or no sell price is set,
	// so the order is expired, but no logic to complete the SO, then will be completed via hooks

	if so.HighestBid == nil {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no bid placed")
	}

	newOwner := so.HighestBid.Bidder

	// complete the Sell-Order

	previousOwner := dymName.Owner

	// bid placed by the bidder will be transferred to the previous owner

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		dymnstypes.ModuleName,
		sdk.MustAccAddressFromBech32(previousOwner),
		sdk.Coins{so.HighestBid.Price},
	); err != nil {
		return err
	}

	// remove SO record
	k.DeleteSellOrder(ctx, so.AssetId, so.AssetType)

	// transfer ownership

	// remove the existing reverse mapping

	if err := k.BeforeDymNameOwnerChanged(ctx, dymName.Name); err != nil {
		return err
	}

	if err := k.BeforeDymNameConfigChanged(ctx, dymName.Name); err != nil {
		return err
	}

	// update Dym records to prevent any potential mistake
	dymName.Owner = newOwner      // ownership transfer
	dymName.Controller = newOwner // new owner becomes the controller
	dymName.Configs = nil         // clear all configs
	dymName.Contact = ""          // clear contact

	// persist updated DymName
	if err := k.SetDymName(ctx, *dymName); err != nil {
		return err
	}

	// update reverse mapping

	if err := k.AfterDymNameOwnerChanged(ctx, dymName.Name); err != nil {
		return err
	}

	if err := k.AfterDymNameConfigChanged(ctx, dymName.Name); err != nil {
		return err
	}

	return nil
}

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// CompleteAliasSellOrder completes the active sell order of the Alias,
// give value to the previous owner, and assign alias usage to destination RollApp.
func (k Keeper) CompleteAliasSellOrder(ctx sdk.Context, name string) error {
	so := k.GetSellOrder(ctx, name, dymnstypes.TypeAlias)
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

	existingRollAppIdUsingAlias, found := k.GetRollAppIdByAlias(ctx, so.AssetId)
	if !found {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias not owned by any RollApp: %s", so.AssetId)
	}

	existingRollAppUsingAlias, found := k.rollappKeeper.GetRollapp(ctx, existingRollAppIdUsingAlias)
	if !found {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "RollApp: %s", existingRollAppIdUsingAlias)
	}

	destinationRollAppId := so.HighestBid.Params[0]
	if !k.IsRollAppId(ctx, destinationRollAppId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "destination Roll-App does not exists: %s", destinationRollAppId)
	}

	// complete the Sell

	// give value to the previous owner
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		dymnstypes.ModuleName,
		sdk.MustAccAddressFromBech32(existingRollAppUsingAlias.Owner),
		sdk.Coins{so.HighestBid.Price},
	); err != nil {
		return err
	}

	// remove SO record
	k.DeleteSellOrder(ctx, so.AssetId, so.AssetType)

	// unlink from source RollApp
	if err := k.RemoveAliasFromRollAppId(ctx, existingRollAppIdUsingAlias, so.AssetId); err != nil {
		return err
	}

	// link to destination RollApp
	if err := k.SetAliasForRollAppId(ctx, destinationRollAppId, so.AssetId); err != nil {
		return err
	}

	return nil
}

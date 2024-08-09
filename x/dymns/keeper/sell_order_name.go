package keeper

import (
	"sort"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// CompleteDymNameSellOrder completes the active sell order of the Dym-Name,
// give value to the previous owner, and transfer ownership to new owner.
func (k Keeper) CompleteDymNameSellOrder(ctx sdk.Context, name string) error {
	dymName := k.GetDymName(ctx, name)
	if dymName == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", name)
	}

	// here we don't check Dym-Name expiration, because it can not happen,
	// and there is a grace period for the owner to renew the Dym-Name in case bad things happen

	so := k.GetSellOrder(ctx, name, dymnstypes.NameOrder)
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

	// complete the Sell

	previousOwner := dymName.Owner

	// give value to the previous owner
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		dymnstypes.ModuleName,
		sdk.MustAccAddressFromBech32(previousOwner),
		sdk.Coins{so.HighestBid.Price},
	); err != nil {
		return err
	}

	// move the SO to history
	if err := k.MoveSellOrderToHistorical(ctx, dymName.Name, so.Type); err != nil {
		return err
	}

	// transfer ownership

	if err := k.BeforeDymNameOwnerChanged(ctx, dymName.Name); err != nil {
		return err
	}

	if err := k.BeforeDymNameConfigChanged(ctx, dymName.Name); err != nil {
		return err
	}

	// update Dym records to prevent any potential mistake
	dymName.Owner = newOwner
	dymName.Controller = newOwner
	dymName.Configs = nil
	dymName.Contact = ""

	// persist updated DymName
	if err := k.SetDymName(ctx, *dymName); err != nil {
		return err
	}

	if err := k.AfterDymNameOwnerChanged(ctx, dymName.Name); err != nil {
		return err
	}

	if err := k.AfterDymNameConfigChanged(ctx, dymName.Name); err != nil {
		return err
	}

	return nil
}

// GetMinExpiryOfAllHistoricalDymNameSellOrders returns the minimum expiry
// of all historical Sell-Orders by each Dym-Name.
func (k Keeper) GetMinExpiryOfAllHistoricalDymNameSellOrders(
	ctx sdk.Context,
) (minExpiryPerDymNameRecords []dymnstypes.HistoricalSellOrderMinExpiry) {
	store := ctx.KVStore(k.storeKey)

	nameToMinExpiry := make(map[string]int64)
	defer func() {
		if len(nameToMinExpiry) < 1 {
			return
		}

		minExpiryPerDymNameRecords = make([]dymnstypes.HistoricalSellOrderMinExpiry, 0, len(nameToMinExpiry))
		for name, minExpiry := range nameToMinExpiry {
			minExpiryPerDymNameRecords = append(minExpiryPerDymNameRecords, dymnstypes.HistoricalSellOrderMinExpiry{
				DymName:   name,
				MinExpiry: minExpiry,
			})
		}

		sort.Slice(minExpiryPerDymNameRecords, func(i, j int) bool {
			return minExpiryPerDymNameRecords[i].DymName < minExpiryPerDymNameRecords[j].DymName
		})
	}()

	iterator := sdk.KVStorePrefixIterator(store, dymnstypes.KeyPrefixMinExpiryDymNameHistoricalSellOrders)
	defer func() {
		_ = iterator.Close() // nolint: errcheck
	}()

	for ; iterator.Valid(); iterator.Next() {
		dymName := string(iterator.Key()[len(dymnstypes.KeyPrefixMinExpiryDymNameHistoricalSellOrders):])
		minExpiry := int64(sdk.BigEndianToUint64(iterator.Value()))

		nameToMinExpiry[dymName] = minExpiry
	}

	return
}

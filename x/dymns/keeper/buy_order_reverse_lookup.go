package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// AddReverseMappingBuyerToBuyOrderRecord stores a reverse mapping from buyer to IDs of Buy-Order into the KVStore.
func (k Keeper) AddReverseMappingBuyerToBuyOrderRecord(ctx sdk.Context, buyer, orderId string) error {
	accAddr, err := sdk.AccAddressFromBech32(buyer)
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid buyer address: %s", buyer)
	}

	if !dymnstypes.IsValidBuyOrderId(orderId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Buy-Order ID: %s", orderId)
	}

	key := dymnstypes.BuyerToOrderIdsRvlKey(accAddr)

	return k.GenericAddReverseLookupBuyOrderIdsRecord(ctx, key, orderId)
}

// GetBuyOrdersByBuyer returns all Buy-Orders placed by the account address.
func (k Keeper) GetBuyOrdersByBuyer(
	ctx sdk.Context, buyer string,
) ([]dymnstypes.BuyOrder, error) {
	accAddr, err := sdk.AccAddressFromBech32(buyer)
	if err != nil {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid buyer address: %s", buyer)
	}

	key := dymnstypes.BuyerToOrderIdsRvlKey(accAddr)

	existingOrderIds := k.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)

	var buyOrders []dymnstypes.BuyOrder
	for _, orderId := range existingOrderIds.OrderIds {
		offer := k.GetBuyOrder(ctx, orderId)
		if offer == nil {
			// offer not found, skip
			continue
		}
		if offer.Buyer != buyer {
			// offer buyer mismatch, skip
			continue
		}
		buyOrders = append(buyOrders, *offer)
	}

	return buyOrders, nil
}

// RemoveReverseMappingBuyerToBuyOrder removes a reverse mapping from buyer to a Buy-Order ID from the KVStore.
func (k Keeper) RemoveReverseMappingBuyerToBuyOrder(ctx sdk.Context, buyer, orderId string) error {
	accAddr, err := sdk.AccAddressFromBech32(buyer)
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid buyer address: %s", buyer)
	}

	if !dymnstypes.IsValidBuyOrderId(orderId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Buy-Order ID: %s", orderId)
	}

	key := dymnstypes.BuyerToOrderIdsRvlKey(accAddr)

	return k.GenericRemoveReverseLookupBuyOrderIdRecord(ctx, key, orderId)
}

// AddReverseMappingAssetIdToBuyOrder add a reverse mapping from Dym-Name/Alias to the Buy-Order ID which placed for it, into the KVStore.
func (k Keeper) AddReverseMappingAssetIdToBuyOrder(ctx sdk.Context, assetId string, assetType dymnstypes.AssetType, orderId string) error {
	var key []byte

	switch assetType {
	case dymnstypes.TypeName:
		if !dymnsutils.IsValidDymName(assetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Dym-Name: %s", assetId)
		}
		key = dymnstypes.DymNameToBuyOrderIdsRvlKey(assetId)
	case dymnstypes.TypeAlias:
		if !dymnsutils.IsValidAlias(assetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Alias: %s", assetId)
		}
		key = dymnstypes.AliasToBuyOrderIdsRvlKey(assetId)
	default:
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", assetType)
	}

	if !dymnstypes.IsValidBuyOrderId(orderId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Buy-Order ID: %s", orderId)
	}

	return k.GenericAddReverseLookupBuyOrderIdsRecord(ctx, key, orderId)
}

// GetBuyOrdersOfDymName returns all Buy-Orders that placed for the Dym-Name.
func (k Keeper) GetBuyOrdersOfDymName(
	ctx sdk.Context, name string,
) ([]dymnstypes.BuyOrder, error) {
	if !dymnsutils.IsValidDymName(name) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Dym-Name: %s", name)
	}

	key := dymnstypes.DymNameToBuyOrderIdsRvlKey(name)

	orderIds := k.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)

	var buyOrders []dymnstypes.BuyOrder
	for _, orderId := range orderIds.OrderIds {
		bo := k.GetBuyOrder(ctx, orderId)
		if bo == nil {
			// not found, skip
			continue
		}
		buyOrders = append(buyOrders, *bo)
	}

	return buyOrders, nil
}

// GetBuyOrdersOfAlias returns all Buy-Orders that placed for the Alias.
func (k Keeper) GetBuyOrdersOfAlias(
	ctx sdk.Context, alias string,
) ([]dymnstypes.BuyOrder, error) {
	if !dymnsutils.IsValidAlias(alias) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Alias: %s", alias)
	}

	key := dymnstypes.AliasToBuyOrderIdsRvlKey(alias)

	orderIds := k.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)

	var buyOrders []dymnstypes.BuyOrder
	for _, orderId := range orderIds.OrderIds {
		buyOrder := k.GetBuyOrder(ctx, orderId)
		if buyOrder == nil {
			// not found, skip
			continue
		}
		buyOrders = append(buyOrders, *buyOrder)
	}

	return buyOrders, nil
}

// RemoveReverseMappingAssetIdToBuyOrder removes reverse mapping from Dym-Name/Alias to Buy-Order which placed for it, from the KVStore.
func (k Keeper) RemoveReverseMappingAssetIdToBuyOrder(ctx sdk.Context, assetId string, assetType dymnstypes.AssetType, orderId string) error {
	var key []byte

	switch assetType {
	case dymnstypes.TypeName:
		if !dymnsutils.IsValidDymName(assetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Dym-Name: %s", assetId)
		}
		key = dymnstypes.DymNameToBuyOrderIdsRvlKey(assetId)
	case dymnstypes.TypeAlias:
		if !dymnsutils.IsValidAlias(assetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Alias: %s", assetId)
		}
		key = dymnstypes.AliasToBuyOrderIdsRvlKey(assetId)
	default:
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", assetType)
	}

	if !dymnstypes.IsValidBuyOrderId(orderId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Buy-Order ID: %s", orderId)
	}

	return k.GenericRemoveReverseLookupBuyOrderIdRecord(ctx, key, orderId)
}

// GenericAddReverseLookupBuyOrderIdsRecord is a utility method that help to add a reverse lookup record for Buy-Order ID.
func (k Keeper) GenericAddReverseLookupBuyOrderIdsRecord(ctx sdk.Context, key []byte, orderId string) error {
	return k.GenericAddReverseLookupRecord(
		ctx,
		key, orderId,
		func(list []string) []byte {
			record := dymnstypes.ReverseLookupBuyOrderIds{
				OrderIds: list,
			}
			return k.cdc.MustMarshal(&record)
		},
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupBuyOrderIds
			k.cdc.MustUnmarshal(bz, &record)
			return record.OrderIds
		},
	)
}

// GenericGetReverseLookupBuyOrderIdsRecord is a utility method that help to get a reverse lookup record for Buy-Order IDs.
func (k Keeper) GenericGetReverseLookupBuyOrderIdsRecord(
	ctx sdk.Context, key []byte,
) dymnstypes.ReverseLookupBuyOrderIds {
	dymNames := k.GenericGetReverseLookupRecord(
		ctx,
		key,
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupBuyOrderIds
			k.cdc.MustUnmarshal(bz, &record)
			return record.OrderIds
		},
	)

	return dymnstypes.ReverseLookupBuyOrderIds{
		OrderIds: dymNames,
	}
}

// GenericRemoveReverseLookupBuyOrderIdRecord is a utility method that help to remove a reverse lookup record for Buy-Order ID.
func (k Keeper) GenericRemoveReverseLookupBuyOrderIdRecord(ctx sdk.Context, key []byte, orderId string) error {
	return k.GenericRemoveReverseLookupRecord(
		ctx,
		key, orderId,
		func(list []string) []byte {
			record := dymnstypes.ReverseLookupBuyOrderIds{
				OrderIds: list,
			}
			return k.cdc.MustMarshal(&record)
		},
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupBuyOrderIds
			k.cdc.MustUnmarshal(bz, &record)
			return record.OrderIds
		},
	)
}

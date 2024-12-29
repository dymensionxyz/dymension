package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
)

/*
// TODO: rename denom to 'ibcDenom' https://github.com/dymensionxyz/dymension/issues/1650
*/

func (k Keeper) SetRegisteredDenom(ctx sdk.Context, rollappID, denom string) error {
	return k.registeredRollappDenoms.Set(ctx, collections.Join(rollappID, denom))
}

func (k Keeper) HasRegisteredDenom(ctx sdk.Context, rollappID, denom string) (bool, error) {
	return k.registeredRollappDenoms.Has(ctx, collections.Join(rollappID, denom))
}

func (k Keeper) GetAllRegisteredDenomsPaginated(ctx sdk.Context, rollappID string, pageReq *query.PageRequest) ([]string, *query.PageResponse, error) {
	return collcompat.CollectionPaginate(ctx, k.registeredRollappDenoms, pageReq,
		func(key collections.Pair[string, string], _ collections.NoValue) (string, error) {
			return key.K1(), nil
		}, collcompat.WithCollectionPaginationPairPrefix[string, string](rollappID),
	)
}

func (k Keeper) GetAllRegisteredDenoms(ctx sdk.Context, rollappID string) ([]string, error) {
	var denoms []string
	if err := k.IterateRegisteredDenoms(ctx, rollappID, func(denom string) (bool, error) {
		denoms = append(denoms, denom)
		return false, nil
	}); err != nil {
		return nil, fmt.Errorf("get all registered denoms: %w", err)
	}
	return denoms, nil
}

func (k Keeper) IterateRegisteredDenoms(ctx sdk.Context, rollappID string, cb func(denom string) (bool, error)) error {
	rng := collections.NewPrefixedPairRange[string, string](rollappID)
	return k.registeredRollappDenoms.Walk(ctx, rng, func(item collections.Pair[string, string]) (bool, error) {
		return cb(item.K2())
	})
}

func (k Keeper) ClearRegisteredDenoms(ctx sdk.Context, rollappID string) error {
	rng := collections.NewPrefixedPairRange[string, string](rollappID)
	return k.registeredRollappDenoms.Clear(ctx, rng)
}

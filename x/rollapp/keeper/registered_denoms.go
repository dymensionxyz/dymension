package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetRegisteredDenom(ctx sdk.Context, rollappID, denom string) error {
	key := collections.Join(rollappID, denom)
	if err := k.registeredRollappDenoms.Set(ctx, key); err != nil {
		return fmt.Errorf("set registered denom: %w", err)
	}
	return nil
}

func (k Keeper) HasRegisteredDenom(ctx sdk.Context, rollappID, denom string) (bool, error) {
	key := collections.Join(rollappID, denom)
	ok, err := k.registeredRollappDenoms.Has(ctx, key)
	if err != nil {
		return false, fmt.Errorf("has registered denom: %w", err)
	}
	return ok, nil
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
	pref := collections.PairPrefix[string, string](rollappID)
	iter, err := k.registeredRollappDenoms.Iterate(ctx, new(collections.Range[collections.Pair[string, string]]).Prefix(pref))
	if err != nil {
		return fmt.Errorf("iterate registered denoms: %w", err)
	}
	defer iter.Close()

	for iter.Valid() {
		key, err := iter.Key()
		if err != nil {
			return fmt.Errorf("get key: %w", err)
		}
		denom := key.K2()
		stop, err := cb(denom)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		iter.Next()
	}

	return nil
}

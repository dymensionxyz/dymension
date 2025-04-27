package v3

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/dymensionxyz/dymension/v3/x/rollapp/migrations/v2"
	types "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// MigrateStore performs in-place store migrations from v2 to v3.
// It migrates:
// - Module parameters from legacy Params subspace to the appropriate store
func Migrate2to3(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, legacySubspace types.Subspace) error {
	// migrate params
	if err := migrateParams(ctx, store, cdc, legacySubspace); err != nil {
		return err
	}

	return nil
}

// migrateParams will set the params to store from legacySubspace
func migrateParams(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, legacySubspace types.Subspace) error {
	var legacyParams v2.Params
	legacySubspace.GetParamSet(ctx, &legacyParams)

	if err := legacyParams.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&legacyParams)
	store.Set(types.KeyParams, bz)
	return nil
}

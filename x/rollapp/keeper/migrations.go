package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3 "github.com/dymensionxyz/dymension/v3/x/rollapp/migrations/v3"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper         Keeper
	legacySubspace types.Subspace
}

// NewMigrator returns a new Migrator instance.
func NewMigrator(keeper Keeper, legacySubspace types.Subspace) Migrator {
	return Migrator{
		keeper:         keeper,
		legacySubspace: legacySubspace,
	}
}

// Migrate2to3 migrates x/staking state from consensus version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.Migrate2to3(ctx, ctx.KVStore(m.keeper.storeKey), m.keeper.cdc, m.legacySubspace)
}

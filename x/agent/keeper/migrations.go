package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// Migrator handles in-place x/agent store migrations.
type Migrator struct {
	k Keeper
}

// NewMigrator returns an x/agent store migrator.
func NewMigrator(k Keeper) Migrator {
	return Migrator{k: k}
}

// Migrate1to2 initializes the recipient allowlist cap for params written
// before the field existed.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	params, err := m.k.GetParams(ctx)
	if err != nil {
		return err
	}
	params.SpendRecipientAllowlistMax = types.DefaultSpendRecipientAllowlistMax
	return m.k.SetParams(ctx, params)
}

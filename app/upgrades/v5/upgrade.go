package v5

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	lockupkeeper "github.com/dymensionxyz/dymension/v3/x/lockup/keeper"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		// (This is how osmosis do it)
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		migrateLockupParams(ctx, keepers.LockupKeeper)

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return migrations, nil
	}
}

func migrateLockupParams(ctx sdk.Context, lockupKeeper *lockupkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	params := lockuptypes.DefaultParams()

	// ForceUnlockAllowedAddresses is the only one that hasn't changed
	params.ForceUnlockAllowedAddresses = lockupKeeper.GetForceUnlockAllowedAddresses(ctx)

	lockupKeeper.SetParams(ctx, params)
}

package v5

import (
	"context"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	irokeeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lockupkeeper "github.com/dymensionxyz/dymension/v3/x/lockup/keeper"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"

	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
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
		// IRO store upgraded through module migrations
		// TXFEES store upgraded through module migrations
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// lockup module params migrations
		migrateLockupParams(ctx, keepers.LockupKeeper)

		// IRO module params migration
		migrateIROParams(ctx, keepers.IROKeeper)

		// GAMM module params migration
		migrateGAMMParams(ctx, keepers.GAMMKeeper)

		// TODO: V50 migrations

		// TODO: IBCKEEPER override

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return migrations, nil
	}
}

func migrateLockupParams(ctx sdk.Context, k *lockupkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	params := k.GetParams(ctx)

	params.LockCreationFee = lockuptypes.DefaultLockFee
	params.MinLockDuration = 24 * time.Hour

	k.SetParams(ctx, params)
}

func migrateGAMMParams(ctx sdk.Context, k *gammkeeper.Keeper) {
	params := k.GetParams(ctx)

	for _, coin := range params.PoolCreationFee {
		params.AllowedPoolCreationDenoms = append(params.AllowedPoolCreationDenoms, coin.Denom)
	}
	k.SetParams(ctx, params)
}

func migrateIROParams(ctx sdk.Context, k *irokeeper.Keeper) {
	params := k.GetParams(ctx)
	defParams := irotypes.DefaultParams()

	params.MinLiquidityPart = defParams.MinLiquidityPart
	params.MinVestingDuration = defParams.MinVestingDuration
	params.MinVestingStartTimeAfterSettlement = defParams.MinVestingStartTimeAfterSettlement

	k.SetParams(ctx, params)
}

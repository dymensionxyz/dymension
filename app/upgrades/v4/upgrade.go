package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v4
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// overwrite params for delayedack module due to added parameters
		daParams := delayedacktypes.DefaultParams()
		keepers.DelayedAckKeeper.SetParams(ctx, daParams)

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

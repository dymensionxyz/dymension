package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
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

		LoadDeprecatedParamsSubspaces(keepers)

		migrateModuleParams(ctx, keepers)
		migrateDelayedAckParams(ctx, keepers.DelayedAckKeeper)
		migrateRollappParams(ctx, keepers.RollappKeeper)
		if err := migrateRollapps(ctx, keepers.RollappKeeper); err != nil {
			return nil, err
		}

		migrateSequencerParams(ctx, keepers.SequencerKeeper)
		migrateSequencers(ctx, keepers.SequencerKeeper)
		migrateRollappLightClients(ctx, keepers.RollappKeeper, keepers.LightClientKeeper, keepers.IBCKeeper.ChannelKeeper)
		if err := migrateStreamer(ctx, keepers.StreamerKeeper, keepers.EpochsKeeper); err != nil {
			return nil, err
		}
		migrateIncentivesParams(ctx, keepers.IncentivesKeeper)

		if err := migrateRollappGauges(ctx, keepers.RollappKeeper, keepers.IncentivesKeeper); err != nil {
			return nil, err
		}

		if err := migrateDelayedAckPacketIndex(ctx, keepers.DelayedAckKeeper); err != nil {
			return nil, err
		}

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

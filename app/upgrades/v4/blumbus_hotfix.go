package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
)

const (
	UpgradeNameBlumbusHotfix = "v4_blumbus_hotfix"
)

var UpgradeBlumbusHotfix = upgrades.Upgrade{
	Name:          UpgradeNameBlumbusHotfix,
	CreateHandler: CreateBlumbusHotfixUpgradeHandler,
}

// CreateUpgradeHandler creates an SDK upgrade handler for v4
func CreateBlumbusHotfixUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		setKeyTables(keepers)

		if err := migrateRollappLightClients(ctx, keepers.RollappKeeper, keepers.LightClientKeeper, keepers.IBCKeeper.ChannelKeeper); err != nil {
			return nil, err
		}

		// terminate active streams
		for _, stream := range keepers.StreamerKeeper.GetActiveStreams(ctx) {
			err := keepers.StreamerKeeper.TerminateStream(ctx, stream.Id)
			if err != nil {
				return nil, err
			}
		}

		return fromVM, nil
	}
}

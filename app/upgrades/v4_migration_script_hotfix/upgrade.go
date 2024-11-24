package v4hotfix

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
)

const (
	UpgradeName = "v4_migration_script_hotfix"
)

var Upgrade = upgrades.Upgrade{
	Name:          UpgradeName,
	CreateHandler: CreateUpgradeHandler,
}

// CreateUpgradeHandler creates an SDK upgrade handler for v4 migration script hotfix
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		k := keepers.DistrKeeper
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// update validator_current_rewards period to 2 instead of 1
		keepers.DistrKeeper.IterateValidatorCurrentRewards(ctx, func(val sdk.ValAddress, rewards disttypes.ValidatorCurrentRewards) bool {
			if rewards.Period != 0 {
				logger.Error("fixing validator current rewards period. expected 0", "val", val, "period", rewards.Period)
				return true
			}
			rewards.Period = 2
			keepers.DistrKeeper.SetValidatorCurrentRewards(ctx, val, rewards)

			// validator_historical_rewards set period 1 instead of 0 with reference count 2
			old := k.GetValidatorHistoricalRewards(ctx, val, 0)
			if old.ReferenceCount != 0 {
				logger.Error("fixing validator historical rewards period. expected reference count 0", "val", val, "period", old.ReferenceCount)
				return true
			}
			k.DeleteValidatorHistoricalReward(ctx, val, 0)

			old.ReferenceCount = 2
			k.SetValidatorHistoricalRewards(ctx, val, 1, old)

			return false
		})

		// fix delegation starting info from 0 to 1
		k.IterateDelegatorStartingInfos(ctx, func(val sdk.ValAddress, del sdk.AccAddress, info disttypes.DelegatorStartingInfo) bool {
			info.PreviousPeriod = 1
			keepers.DistrKeeper.SetDelegatorStartingInfo(ctx, val, del, info)
			return false
		})

		// remove slash events
		k.DeleteAllValidatorSlashEvents(ctx)

		return fromVM, nil
	}
}

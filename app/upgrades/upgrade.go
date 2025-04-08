package upgrades

import (
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	incentiveskeeper "github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	irokeeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	lockupkeeper "github.com/dymensionxyz/dymension/v3/x/lockup/keeper"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
)

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.go.
// The app.go will then define the handler.
type Upgrade struct {
	// Upgrade version name, for the upgrade handler, e.g. `v4`
	Name string

	// CreateHandler defines the function that creates an upgrade handler
	CreateHandler func(
		mm *module.Manager,
		configurator module.Configurator,
		appKeepers *UpgradeKeepers,
	) upgradetypes.UpgradeHandler

	// Store upgrades, should be used for any new modules introduced, new modules deleted, or store names renamed.
	StoreUpgrades storetypes.StoreUpgrades
}

type UpgradeKeepers struct {
	LockupKeeper     *lockupkeeper.Keeper
	IROKeeper        *irokeeper.Keeper
	GAMMKeeper       *gammkeeper.Keeper
	GovKeeper        *govkeeper.Keeper
	IncentivesKeeper *incentiveskeeper.Keeper
}

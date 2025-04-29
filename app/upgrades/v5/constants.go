package v5

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	hypercoretypes "github.com/dymensionxyz/hyperlane-cosmos/x/core/types"
	hyperwarptypes "github.com/dymensionxyz/hyperlane-cosmos/x/warp/types"
)

const (
	UpgradeName = "v5"
)

var Upgrade = upgrades.Upgrade{
	Name:          UpgradeName,
	CreateHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			hypercoretypes.ModuleName,
			hyperwarptypes.ModuleName,
		},
	},
}

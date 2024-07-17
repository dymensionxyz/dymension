package v3

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

const (
	UpgradeName = "v3"
)

var Upgrade = upgrades.Upgrade{
	Name:          UpgradeName,
	CreateHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{eibctypes.ModuleName},
	},
}

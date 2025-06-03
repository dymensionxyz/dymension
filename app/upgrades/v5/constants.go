package v5

import (
	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"

	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	hyperwarptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	kastypes "github.com/dymensionxyz/dymension/v3/x/kas/types"
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
			kastypes.ModuleName,
			circuittypes.ModuleName,
		},
	},
}

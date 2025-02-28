package v4

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"

	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lightclienttypes "github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

const (
	UpgradeName = "v4"
)

var Upgrade = upgrades.Upgrade{
	Name:          UpgradeName,
	CreateHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			consensustypes.ModuleName,
			crisistypes.ModuleName,
			lightclienttypes.ModuleName,
			sponsorshiptypes.ModuleName,
			dymnstypes.ModuleName,
			irotypes.ModuleName,
			grouptypes.ModuleName,
		},
	},
}

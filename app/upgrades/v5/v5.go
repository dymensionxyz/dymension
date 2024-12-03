package v5

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lightclienttypes "github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// playground only

const (
	UpgradeName = "v5"
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

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		log := ctx.Logger().With("upgrade", UpgradeName)

		p := keepers.RollappKeeper.GetParams(ctx)
		p.MinSequencerBondGlobal = rollapptypes.DefaultMinSequencerBondGlobalCoin
		keepers.RollappKeeper.SetParams(ctx, p)
		rollapps := keepers.RollappKeeper.GetAllRollapps(ctx)
		for _, ra := range rollapps {
			ra.MinSequencerBond = sdk.NewCoins(rollapptypes.DefaultMinSequencerBondGlobalCoin)
			if ra.GenesisState.GetDeprecatedBridgeOpen() {
				h, ok := keepers.RollappKeeper.GetLatestHeight(ctx, ra.RollappId)
				if !ok {
					log.Error("latest height for transfer enabled not found")
				} else {
					ra.GenesisState.TransferProofHeight = h
				}
			}
			keepers.RollappKeeper.SetRollapp(ctx, ra)
		}
		return fromVM, nil
	}
}

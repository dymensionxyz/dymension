package v3

import (
	"math/big"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"

	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	seqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func GetStoreUpgrades() *storetypes.StoreUpgrades {
	storeUpgrades := storetypes.StoreUpgrades{
		// Set migrations for all new modules
		Added: []string{"eibc"},
	}
	return &storeUpgrades
}

// CreateUpgradeHandler creates an SDK upgrade handler for v3
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	rollappkeeper RollappKeeper,
	seqkeeper SequencerKeeper,
	dakeeper DelayedAckKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		newVM, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return newVM, err
		}

		// overwrite params for delayedack module due to proto change
		daParams := delayedacktypes.DefaultParams()
		dakeeper.SetParams(ctx, daParams)

		// overwrite params for rollapp module due to proto change
		rollappParams := rollapptypes.DefaultParams()
		rollappParams.RollappsEnabled = false
		rollappParams.DisputePeriodInBlocks = 120960 // 1 week
		rollappkeeper.SetParams(ctx, rollappParams)

		// overwrite params for sequencer module due to proto change
		DYM := sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		seqParams := seqtypes.DefaultParams()
		seqParams.MinBond = sdk.NewCoin(appparams.BaseDenom, DYM.Mul(sdk.NewInt(1000))) // 1000DYM
		seqkeeper.SetParams(ctx, seqParams)

		return newVM, nil
	}
}

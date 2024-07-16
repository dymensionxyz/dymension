package v3

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	seqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v3
func CreateUpgradeHandler(
	mm *module.Manager,
	_ codec.Codec,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// overwrite params for delayedack module due to proto change
		daParams := delayedacktypes.DefaultParams()
		keepers.DelayedAckKeeper.SetParams(ctx, daParams)

		// overwrite params for rollapp module due to proto change
		rollappParams := rollapptypes.DefaultParams()
		rollappParams.DisputePeriodInBlocks = 120960 // 1 week
		keepers.RollappKeeper.SetParams(ctx, rollappParams)

		// overwrite params for sequencer module due to proto change
		seqParams := seqtypes.DefaultParams()
		DYM := sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		seqParams.MinBond = sdk.NewCoin(appparams.BaseDenom, DYM.Mul(sdk.NewInt(1000))) // 1000DYM
		keepers.SequencerKeeper.SetParams(ctx, seqParams)

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

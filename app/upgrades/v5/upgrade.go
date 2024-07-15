package v5

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	appCodec codec.Codec,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
	getStoreKey func(string) *storetypes.KVStoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		migrateRollappParams(ctx, keepers.RollappKeeper)
		migrateRollapps(ctx, getStoreKey(rollapptypes.ModuleName), appCodec, keepers.RollappKeeper)

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func migrateRollappParams(ctx sdk.Context, rollappkeeper rollappkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	params := rollapptypes.DefaultParams()
	params.DisputePeriodInBlocks = rollappkeeper.DisputePeriodInBlocks(ctx)
	rollappkeeper.SetParams(ctx, params)
}

func migrateRollapps(ctx sdk.Context, rollappStoreKey *storetypes.KVStoreKey, appCodec codec.Codec, rollappkeeper rollappkeeper.Keeper) {
	for _, oldRollapp := range getAllOldRollapps(ctx, rollappStoreKey, appCodec) {
		newRollapp := ConvertOldRollappToNew(oldRollapp)
		rollappkeeper.SetRollapp(ctx, newRollapp)
	}
}

func ConvertOldRollappToNew(oldRollapp types.Rollapp) rollapptypes.Rollapp {
	return rollapptypes.Rollapp{
		RollappId: oldRollapp.RollappId,
		Creator:   oldRollapp.Creator,
		GenesisState: rollapptypes.RollappGenesisState{
			TransfersEnabled: oldRollapp.GenesisState.TransfersEnabled,
		},
		ChannelId:               oldRollapp.ChannelId,
		Frozen:                  oldRollapp.Frozen,
		RegisteredDenoms:        oldRollapp.RegisteredDenoms,
		InitialSequencerAddress: "", // whatever
		GenesisChecksum:         "", // TODO
		Bech32Prefix:            "", // TODO
		Website:                 "", // TODO
		Description:             "", // TODO
		LogoDataUri:             "", // TODO
		Alias:                   "", // TODO
	}
}

func getAllOldRollapps(ctx sdk.Context, storeKey *storetypes.KVStoreKey, appCodec codec.Codec) (list []types.Rollapp) {
	store := prefix.NewStore(ctx.KVStore(storeKey), rollapptypes.KeyPrefix(rollapptypes.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Rollapp
		bz := iterator.Value()
		appCodec.MustUnmarshalJSON(bz, &val)
		list = append(list, val)
	}

	return
}

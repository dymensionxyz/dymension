package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func GetStoreUpgrades() *storetypes.StoreUpgrades {
	storeUpgrades := storetypes.StoreUpgrades{
		// Set migrations for all new modules
	}
	return &storeUpgrades
}

// CreateHandler creates an SDK upgrade handler for v5
func CreateHandler(
	mm *module.Manager,
	appCodec codec.Codec,
	configurator module.Configurator,
	rollappkeeper RollappKeeper,
	storeKey *storetypes.KVStoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// overwrite params for rollapp module due to proto change
		rollappParams := rollappkeeper.GetParams(ctx)
		rollappParams.DisputePeriodInBlocks = rollapptypes.DefaultDisputePeriodInBlocks
		rollappkeeper.SetParams(ctx, rollappParams)

		oldRollapps := getAllOldRollapps(ctx, storeKey, appCodec)
		newRollapps := make([]rollapptypes.Rollapp, len(oldRollapps))

		for i, oldRollapp := range oldRollapps {
			newRollapp := rollapptypes.Rollapp{
				RollappId:               oldRollapp.RollappId,
				Creator:                 oldRollapp.Creator,
				InitialSequencerAddress: "",
				GenesisInfo:             nil,
				GenesisState: rollapptypes.RollappGenesisState{
					TransfersEnabled: oldRollapp.GenesisState.TransfersEnabled,
				},
				ChannelId: oldRollapp.ChannelId,
				Frozen:    oldRollapp.Frozen,
				// Bech32Prefix:            oldRollapp.Bech32Prefix,
				RegisteredDenoms: oldRollapp.RegisteredDenoms,
				Version:          oldRollapp.Version,
			}
			newRollapps[i] = newRollapp
		}

		// delete old rollapps
		for _, oldRollapp := range oldRollapps {
			rollappkeeper.RemoveRollapp(ctx, oldRollapp.RollappId)
		}

		// add new rollapps
		for _, newRollapp := range newRollapps {
			rollappkeeper.SetRollapp(ctx, newRollapp)
		}

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func getAllOldRollapps(ctx sdk.Context, storeKey *storetypes.KVStoreKey, appCodec codec.Codec) (list []types.Rollapp) {
	store := prefix.NewStore(ctx.KVStore(storeKey), rollapptypes.KeyPrefix(rollapptypes.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Rollapp
		appCodec.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

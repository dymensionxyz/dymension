package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v4/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v4
func CreateUpgradeHandler(
	mm *module.Manager,
	appCodec codec.Codec,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		migrateDelayedAckParams(ctx, keepers.DelayedAckKeeper)
		migrateRollappParams(ctx, keepers.RollappKeeper)
		migrateRollapps(ctx, keepers.GetKey(rollapptypes.ModuleName), appCodec, keepers.RollappKeeper)
		migrateSequencers(ctx, keepers.GetKey(sequencertypes.ModuleName), appCodec, keepers.SequencerKeeper)

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func migrateDelayedAckParams(ctx sdk.Context, delayedAckKeeper delayedackkeeper.Keeper) {
	// overwrite params for delayedack module due to added parameters
	params := delayedacktypes.DefaultParams()
	delayedAckKeeper.SetParams(ctx, params)
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

func migrateSequencers(ctx sdk.Context, sequencerStoreKey *storetypes.KVStoreKey, appCodec codec.Codec, sequencerkeeper sequencerkeeper.Keeper) {
	list := getAllOldSequencers(ctx, sequencerStoreKey, appCodec)
	for _, oldSequencer := range list {
		newSequencer := ConvertOldSequencerToNew(oldSequencer)
		sequencerkeeper.SetSequencer(ctx, newSequencer)
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
		Alias:                   "", // TODO IMPORTANT! This is needed for GetByAlias endpoint
		Metadata: &rollapptypes.RollappMetadata{
			Website:      "", // TODO
			Description:  "", // TODO
			LogoDataUri:  "", // TODO
			TokenLogoUri: "", // TODO
		},
	}
}

var defaultGasPrice, _ = sdk.NewIntFromString("10000000000")

func ConvertOldSequencerToNew(oldSequencer types.Sequencer) sequencertypes.Sequencer {
	return sequencertypes.Sequencer{
		Address:      oldSequencer.SequencerAddress,
		DymintPubKey: oldSequencer.DymintPubKey,
		RollappId:    oldSequencer.RollappId,
		Status:       sequencertypes.OperatingStatus(oldSequencer.Status),
		Proposer:     oldSequencer.Proposer,
		Tokens:       oldSequencer.Tokens,
		Metadata: sequencertypes.SequencerMetadata{
			Moniker:         oldSequencer.Description.Moniker,
			Identity:        oldSequencer.Description.Identity,
			SecurityContact: oldSequencer.Description.SecurityContact,
			Details:         oldSequencer.Description.Details,
			GasPrice:        &defaultGasPrice,
			// P2PSeed:         "", TODO
			// Rpcs:            []string{}, TODO
			// EvmRpcs:         []string{}, TODO
			// RestApiUrls:     []string{}, TODO
			// GenesisUrl:      "", TODO
			// ExplorerUrl:     "", TODO
			// ContactDetails: &sequencertypes.ContactDetails{ TODO
			// 	Website:  "", TODO
			// 	Telegram: "", TODO
			// 	X:        "", TODO
			// }, TODO
			// ExtraData: []byte{}, TODO
			// Snapshots: []*sequencertypes.SnapshotInfo{}, TODO
		},
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

func getAllOldSequencers(ctx sdk.Context, storeKey *storetypes.KVStoreKey, appCodec codec.Codec) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(storeKey), sequencertypes.SequencersKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		bz := iterator.Value()
		appCodec.MustUnmarshalJSON(bz, &val)
		list = append(list, val)
	}

	return
}

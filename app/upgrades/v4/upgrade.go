package v4

import (
	"strings"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/baseapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	// Ethermint modules
	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
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
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		migrateModuleParams(ctx, keepers)
		migrateDelayedAckParams(ctx, keepers.DelayedAckKeeper)
		migrateRollappParams(ctx, keepers.RollappKeeper)
		if err := migrateRollapps(ctx, keepers.RollappKeeper); err != nil {
			return nil, err
		}
		migrateSequencers(ctx, keepers.SequencerKeeper)

		// TODO: create rollapp gauges for each existing rollapp

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

//nolint:staticcheck
func migrateModuleParams(ctx sdk.Context, keepers *keepers.AppKeepers) {
	// Set param key table for params module migration
	for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		case authtypes.ModuleName:
			keyTable = authtypes.ParamKeyTable()
		case banktypes.ModuleName:
			keyTable = banktypes.ParamKeyTable()
		case stakingtypes.ModuleName:
			keyTable = stakingtypes.ParamKeyTable()
		case minttypes.ModuleName:
			keyTable = minttypes.ParamKeyTable()
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable()
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable()
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable()
		case crisistypes.ModuleName:
			keyTable = crisistypes.ParamKeyTable()

		// Ethermint  modules
		case evmtypes.ModuleName:
			keyTable = evmtypes.ParamKeyTable()
		case feemarkettypes.ModuleName:
			keyTable = feemarkettypes.ParamKeyTable()
		default:
			continue
		}

		if !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}
	}
	// Migrate Tendermint consensus parameters from x/params module to a dedicated x/consensus module.
	baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
	baseapp.MigrateParams(ctx, baseAppLegacySS, &keepers.ConsensusParamsKeeper)
}

func migrateDelayedAckParams(ctx sdk.Context, delayedAckKeeper delayedackkeeper.Keeper) {
	// overwrite params for delayedack module due to added parameters
	params := delayedacktypes.DefaultParams()
	delayedAckKeeper.SetParams(ctx, params)
}

func migrateRollappParams(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	params := rollapptypes.DefaultParams()
	params.DisputePeriodInBlocks = rollappkeeper.DisputePeriodInBlocks(ctx)
	rollappkeeper.SetParams(ctx, params)
}

func migrateRollapps(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper) error {
	list := rollappkeeper.GetAllRollapps(ctx)
	for _, oldRollapp := range list {
		newRollapp := ConvertOldRollappToNew(oldRollapp)
		if err := newRollapp.ValidateBasic(); err != nil {
			return err
		}
		rollappkeeper.SetRollapp(ctx, newRollapp)
	}
	return nil
}

func migrateSequencers(ctx sdk.Context, sequencerkeeper sequencerkeeper.Keeper) {
	list := sequencerkeeper.GetAllSequencers(ctx)
	for _, oldSequencer := range list {
		newSequencer := ConvertOldSequencerToNew(oldSequencer)
		sequencerkeeper.SetSequencer(ctx, newSequencer)
	}
}

func ConvertOldRollappToNew(oldRollapp rollapptypes.Rollapp) rollapptypes.Rollapp {
	bech32Prefix := oldRollapp.RollappId[:5]
	alias := strings.Split(oldRollapp.RollappId, "_")[0]
	return rollapptypes.Rollapp{
		RollappId:        oldRollapp.RollappId,
		Owner:            oldRollapp.Owner,
		GenesisState:     oldRollapp.GenesisState,
		ChannelId:        oldRollapp.ChannelId,
		Frozen:           oldRollapp.Frozen,
		RegisteredDenoms: oldRollapp.RegisteredDenoms,
		// TODO: regarding missing data - https://github.com/dymensionxyz/dymension/issues/986
		Bech32Prefix:    bech32Prefix,                                        // placeholder data
		GenesisChecksum: string(crypto.Sha256([]byte(oldRollapp.RollappId))), // placeholder data
		Alias:           alias,                                               // placeholder data
		Metadata: &rollapptypes.RollappMetadata{
			Website:          "",
			Description:      "",
			LogoDataUri:      "",
			TokenLogoDataUri: "",
			Telegram:         "",
			X:                "",
		},
	}
}

var defaultGasPrice, _ = sdk.NewIntFromString("10000000000")

func ConvertOldSequencerToNew(old sequencertypes.Sequencer) sequencertypes.Sequencer {
	return sequencertypes.Sequencer{
		Address:      old.Address,
		DymintPubKey: old.DymintPubKey,
		RollappId:    old.RollappId,
		Status:       old.Status,
		Proposer:     old.Proposer,
		Tokens:       old.Tokens,
		Metadata: sequencertypes.SequencerMetadata{
			Moniker: old.Metadata.Moniker,
			Details: old.Metadata.Details,
			// TODO: regarding missing data - https://github.com/dymensionxyz/dymension/issues/987
			P2PSeeds:    nil,
			Rpcs:        nil,
			EvmRpcs:     nil,
			RestApiUrls: []string{},
			ExplorerUrl: "",
			GenesisUrls: []string{},
			ContactDetails: &sequencertypes.ContactDetails{
				Website:  "",
				Telegram: "",
				X:        "",
			},
			ExtraData: nil,
			Snapshots: []*sequencertypes.SnapshotInfo{},
			GasPrice:  &defaultGasPrice,
		},
	}
}

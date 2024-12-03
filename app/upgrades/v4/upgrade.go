package v4

import (
	"slices"
	"strings"

	errorsmod "cosmossdk.io/errors"
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
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcchannelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	epochskeeper "github.com/osmosis-labs/osmosis/v15/x/epochs/keeper"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	incentiveskeeper "github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	streamerkeeper "github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
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

		setKeyTables(keepers)

		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		// (This is how osmosis do it)
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}
		migrateModuleParams(ctx, keepers)
		keepers.MigrateModuleAccountPerms(ctx)

		if err := deprecateCrisisModule(ctx, keepers.CrisisKeeper); err != nil {
			return nil, err
		}

		migrateDelayedAckParams(ctx, keepers.DelayedAckKeeper)
		migrateRollappParams(ctx, keepers.RollappKeeper)
		if err := migrateRollapps(ctx, keepers.RollappKeeper, keepers.DymNSKeeper); err != nil {
			return nil, err
		}

		migrateSequencerParams(ctx, keepers.SequencerKeeper)
		if err := migrateSequencerIndices(ctx, keepers.SequencerKeeper); err != nil {
			return nil, errorsmod.Wrap(err, "migrate sequencer indices")
		}
		migrateSequencers(ctx, keepers.SequencerKeeper)

		migrateRollappLightClients(ctx, keepers.RollappKeeper, keepers.LightClientKeeper, keepers.IBCKeeper.ChannelKeeper)
		if err := migrateStreamer(ctx, keepers.StreamerKeeper, keepers.EpochsKeeper); err != nil {
			return nil, err
		}
		migrateIncentivesParams(ctx, keepers.IncentivesKeeper)

		if err := migrateRollappGauges(ctx, keepers.RollappKeeper, keepers.IncentivesKeeper); err != nil {
			return nil, err
		}

		if err := migrateDelayedAckPacketIndex(ctx, keepers.DelayedAckKeeper); err != nil {
			return nil, err
		}

		if err := migrateRollappRegisteredDenoms(ctx, keepers.RollappKeeper); err != nil {
			return nil, err
		}

		if err := migrateRollappStateInfoNextProposer(ctx, keepers.RollappKeeper, keepers.SequencerKeeper); err != nil {
			return nil, err
		}

		if err := migrateRollappFinalizationQueue(ctx, keepers.RollappKeeper); err != nil {
			return nil, err
		}

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return migrations, nil
	}
}

func setKeyTables(keepers *keepers.AppKeepers) {
	for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		// Cosmos SDK modules
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

		// Dymension modules
		case rollapptypes.ModuleName:
			keyTable = rollapptypes.ParamKeyTable()
		case sequencertypes.ModuleName:
			continue

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
}

//nolint:staticcheck - note this is a cosmos SDK supplied function specifically for upgrading consensus params
func migrateModuleParams(ctx sdk.Context, keepers *keepers.AppKeepers) {
	// Migrate Tendermint consensus parameters from x/params module to a dedicated x/consensus module.
	baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
	baseapp.MigrateParams(ctx, baseAppLegacySS, &keepers.ConsensusParamsKeeper)
}

// migrateRollappGauges creates a gauge for each rollapp in the store
func migrateRollappGauges(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper, incentivizeKeeper *incentiveskeeper.Keeper) error {
	rollapps := rollappkeeper.GetAllRollapps(ctx)
	for _, rollapp := range rollapps {
		_, err := incentivizeKeeper.CreateRollappGauge(ctx, rollapp.RollappId)
		if err != nil {
			return err
		}
	}
	return nil
}

func migrateRollapps(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper, dymnsKeeper dymnskeeper.Keeper) error {
	// in theory, there should be only two rollapps in the store, but we iterate over all of them just in case
	list := rollappkeeper.GetAllRollapps(ctx)
	for _, oldRollapp := range list {
		newRollapp := ConvertOldRollappToNew(oldRollapp)
		if err := newRollapp.ValidateBasic(); err != nil {
			return err
		}
		rollappkeeper.SetRollapp(ctx, newRollapp)

		switch oldRollapp.RollappId {
		case nimRollappID:
			if err := dymnsKeeper.SetAliasForRollAppId(ctx, oldRollapp.RollappId, nimAlias); err != nil {
				return err
			}
		case mandeRollappID:
			if err := dymnsKeeper.SetAliasForRollAppId(ctx, oldRollapp.RollappId, mandeAlias); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateRollappLightClients(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper, lightClientKeeper lightclientkeeper.Keeper, ibcChannelKeeper ibcchannelkeeper.Keeper) {
	list := rollappkeeper.GetAllRollapps(ctx)
	for _, rollapp := range list {
		// check if the rollapp has a canonical channel already
		if rollapp.ChannelId == "" {
			return
		}
		// get the client ID the channel belongs to
		_, connection, err := ibcChannelKeeper.GetChannelConnection(ctx, ibctransfertypes.PortID, rollapp.ChannelId)
		if err != nil {
			// if could not find a connection, skip the canonical client assignment
			return
		}
		clientID := connection.GetClientID()
		// store the rollapp to canonical light client ID mapping
		lightClientKeeper.SetCanonicalClient(ctx, rollapp.RollappId, clientID)
	}
}

// migrateStreamer creates epoch pointers for all epoch infos and updates module params
func migrateStreamer(ctx sdk.Context, sk streamerkeeper.Keeper, ek *epochskeeper.Keeper) error {
	// update module params
	params := streamertypes.DefaultParams()
	sk.SetParams(ctx, params)

	// create epoch pointers for all epoch infos
	for _, epoch := range ek.AllEpochInfos(ctx) {
		err := sk.SaveEpochPointer(ctx, streamertypes.NewEpochPointer(epoch.Identifier, epoch.Duration))
		if err != nil {
			return err
		}
	}
	return nil
}

func migrateRollappFinalizationQueue(ctx sdk.Context, rk *rollappkeeper.Keeper) error {
	q := rk.GetAllBlockHeightToFinalizationQueue(ctx)

	// iterate over queues on different heights
	for _, queue := range q {
		// convert the queue to the new format
		newQueues := ReformatFinalizationQueue(queue)

		// save the new queues
		for _, newQueue := range newQueues {
			err := rk.SetFinalizationQueue(ctx, newQueue)
			if err != nil {
				return err
			}
		}

		// remove the old queue
		rk.RemoveBlockHeightToFinalizationQueue(ctx, queue.CreationHeight)
	}
	return nil
}

// ReformatFinalizationQueue groups the finalization queue by rollapp
func ReformatFinalizationQueue(queue rollapptypes.BlockHeightToFinalizationQueue) []rollapptypes.BlockHeightToFinalizationQueue {
	// Map is used for convenient data aggregation.
	// Later it is converted to a slice and sorted by rollappID, so the output is always deterministic.
	grouped := make(map[string][]rollapptypes.StateInfoIndex)

	// group indexes by rollapp
	for _, index := range queue.FinalizationQueue {
		grouped[index.RollappId] = append(grouped[index.RollappId], index)
	}

	// cast map to slice
	queues := make([]rollapptypes.BlockHeightToFinalizationQueue, 0, len(grouped))
	for rollappID, indexes := range grouped {
		queues = append(queues, rollapptypes.BlockHeightToFinalizationQueue{
			CreationHeight:    queue.CreationHeight,
			FinalizationQueue: indexes,
			RollappId:         rollappID,
		})
	}

	// sort by rollappID
	slices.SortFunc(queues, func(a, b rollapptypes.BlockHeightToFinalizationQueue) int {
		return strings.Compare(a.RollappId, b.RollappId)
	})

	return queues
}

func migrateIncentivesParams(ctx sdk.Context, ik *incentiveskeeper.Keeper) {
	params := incentivestypes.DefaultParams()
	params.DistrEpochIdentifier = ik.DistrEpochIdentifier(ctx)
	ik.SetParams(ctx, params)
}

func migrateDelayedAckPacketIndex(ctx sdk.Context, dk delayedackkeeper.Keeper) error {
	pendingPackets := dk.ListRollappPackets(ctx, delayedacktypes.ByStatus(commontypes.Status_PENDING))
	for _, packet := range pendingPackets {
		pd, err := packet.GetTransferPacketData()
		if err != nil {
			return err
		}

		switch packet.Type {
		case commontypes.RollappPacket_ON_RECV:
			dk.MustSetPendingPacketByAddress(ctx, pd.Receiver, packet.RollappPacketKey())
		case commontypes.RollappPacket_ON_ACK, commontypes.RollappPacket_ON_TIMEOUT:
			dk.MustSetPendingPacketByAddress(ctx, pd.Sender, packet.RollappPacketKey())
		}
	}
	return nil
}

func ConvertOldRollappToNew(oldRollapp rollapptypes.Rollapp) rollapptypes.Rollapp {
	genesisInfo := rollapptypes.GenesisInfo{
		Bech32Prefix:    oldRollapp.RollappId[:5],                            // placeholder data
		GenesisChecksum: string(crypto.Sha256([]byte(oldRollapp.RollappId))), // placeholder data
		NativeDenom: rollapptypes.DenomMetadata{
			Display:  "DEN",  // placeholder data
			Base:     "aden", // placeholder data
			Exponent: 18,     // placeholder data
		},
		InitialSupply: sdk.NewInt(100000), // placeholder data
		Sealed:        true,
	}

	// migrate existing rollapps

	// mainnet
	if oldRollapp.RollappId == nimRollappID {
		genesisInfo = nimGenesisInfo
	}
	if oldRollapp.RollappId == mandeRollappID {
		genesisInfo = mandeGenesisInfo
	}

	// testnet
	if oldRollapp.RollappId == rollappXRollappID {
		genesisInfo = rollappXGenesisInfo
	}
	if oldRollapp.RollappId == crynuxRollappID {
		genesisInfo = crynuxGenesisInfo
	}

	genesisState := oldRollapp.GenesisState
	// min(nim=813701, mande=1332619) on Dec 2nd : a sufficient and safe number
	genesisState.TransferProofHeight = 813701

	return rollapptypes.Rollapp{
		RollappId:    oldRollapp.RollappId,
		Owner:        oldRollapp.Owner,
		GenesisState: genesisState,
		ChannelId:    oldRollapp.ChannelId,
		Metadata: &rollapptypes.RollappMetadata{ // Can be updated in runtime
			Website:     "",
			Description: "",
			LogoUrl:     "",
			Telegram:    "",
			X:           "",
			GenesisUrl:  "",
			DisplayName: "",
			Tagline:     "",
			FeeDenom:    nil,
		},
		GenesisInfo:                  genesisInfo,
		InitialSequencer:             "*",
		MinSequencerBond:             sdk.NewCoins(rollapptypes.DefaultMinSequencerBondGlobalCoin),
		VmType:                       rollapptypes.Rollapp_EVM, // EVM for existing rollapps
		Launched:                     true,                     // Existing rollapps are already launched
		PreLaunchTime:                nil,                      // We can just let it be zero. Existing rollapps are already launched.
		LivenessEventHeight:          0,                        // Filled lazily in runtime
		LivenessCountdownStartHeight: 0,                        // Filled lazily in runtime
		Revisions: []rollapptypes.Revision{{
			Number:      0,
			StartHeight: 0,
		}},
	}
}

var defaultGasPrice, _ = sdk.NewIntFromString("10000000000")

func ConvertOldSequencerToNew(old sequencertypes.Sequencer) sequencertypes.Sequencer {
	return sequencertypes.Sequencer{
		Address:      old.Address,
		DymintPubKey: old.DymintPubKey,
		RollappId:    old.RollappId,
		Status:       old.Status,
		Tokens:       old.Tokens,
		OptedIn:      true,
		Proposer:     old.Proposer,
		Metadata: sequencertypes.SequencerMetadata{
			Moniker:     old.Metadata.Moniker,
			Details:     old.Metadata.Details,
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

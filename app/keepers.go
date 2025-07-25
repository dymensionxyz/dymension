package app

import (
	"context"

	"cosmossdk.io/log"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	ethflags "github.com/evmos/ethermint/server/flags"

	storetypes "cosmossdk.io/store/types"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcporttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"
	"github.com/evmos/ethermint/x/evm"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/ethermint/x/evm/vm/geth"
	feemarketkeeper "github.com/evmos/ethermint/x/feemarket/keeper"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	epochskeeper "github.com/osmosis-labs/osmosis/v15/x/epochs/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagerkeeper "github.com/osmosis-labs/osmosis/v15/x/poolmanager/keeper"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"
	"github.com/spf13/cast"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"

	circuittypes "cosmossdk.io/x/circuit/types"
	delayedackmodule "github.com/dymensionxyz/dymension/v3/x/delayedack"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	denommetadatamodule "github.com/dymensionxyz/dymension/v3/x/denommetadata"
	denommetadatamodulekeeper "github.com/dymensionxyz/dymension/v3/x/denommetadata/keeper"
	denommetadatamoduletypes "github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	eibcmoduletypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	incentiveskeeper "github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	irokeeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lightclientmodulekeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	lightclientmoduletypes "github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	lockupkeeper "github.com/dymensionxyz/dymension/v3/x/lockup/keeper"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/genesisbridge"
	rollappmodulekeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollappmoduletypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencermodulekeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencermoduletypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	sponsorshipkeeper "github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	streamermodulekeeper "github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	streamermoduletypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	vfchooks "github.com/dymensionxyz/dymension/v3/x/vfc/hooks"

	hypercorekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	hyperwarpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	hyperwarptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	kaskeeper "github.com/dymensionxyz/dymension/v3/x/kas/keeper"
	kastypes "github.com/dymensionxyz/dymension/v3/x/kas/types"

	forward "github.com/dymensionxyz/dymension/v3/x/forward"
	forwardtypes "github.com/dymensionxyz/dymension/v3/x/forward/types"
)

type AppKeepers struct {
	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CircuitBreakerKeeper  circuitkeeper.Keeper

	// IBC keepers
	RateLimitingKeeper            ratelimitkeeper.Keeper
	IBCKeeper                     *ibckeeper.Keeper
	TransferKeeper                ibctransferkeeper.Keeper
	TransferStack                 ibcporttypes.IBCModule
	PacketForwardMiddlewareKeeper *packetforwardkeeper.Keeper
	CapabilityKeeper              *capabilitykeeper.Keeper
	DelayedAckMiddleware          *delayedackmodule.IBCMiddleware
	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper

	// Ethermint keepers
	EvmKeeper       *evmkeeper.Keeper
	FeeMarketKeeper feemarketkeeper.Keeper

	// Osmosis keepers
	GAMMKeeper        *gammkeeper.Keeper
	PoolManagerKeeper *poolmanagerkeeper.Keeper
	LockupKeeper      *lockupkeeper.Keeper
	EpochsKeeper      *epochskeeper.Keeper
	IncentivesKeeper  *incentiveskeeper.Keeper
	TxFeesKeeper      *txfeeskeeper.Keeper

	RollappKeeper     *rollappmodulekeeper.Keeper
	IROKeeper         *irokeeper.Keeper
	SequencerKeeper   *sequencermodulekeeper.Keeper
	SponsorshipKeeper sponsorshipkeeper.Keeper
	StreamerKeeper    streamermodulekeeper.Keeper
	EIBCKeeper        eibckeeper.Keeper
	LightClientKeeper lightclientmodulekeeper.Keeper

	DelayedAckKeeper    delayedackkeeper.Keeper
	DenomMetadataKeeper *denommetadatamodulekeeper.Keeper

	DymNSKeeper dymnskeeper.Keeper

	HyperCoreKeeper hypercorekeeper.Keeper
	HyperWarpKeeper hyperwarpkeeper.Keeper
	KasKeeper       *kaskeeper.Keeper

	Forward *forward.Forward

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey
}

// InitKeepers initializes all keepers for the app
func (a *AppKeepers) InitKeepers(
	appCodec codec.Codec,
	legacyAmino *codec.LegacyAmino,
	bApp *baseapp.BaseApp,
	logger log.Logger,
	moduleAccountAddrs map[string]bool,
	appOpts servertypes.AppOptions,
) {
	govModuleAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	tracer := cast.ToString(appOpts.Get(ethflags.EVMTracer))
	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))

	// init keepers
	a.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, a.keys[paramstypes.StoreKey], a.tkeys[paramstypes.TStoreKey])
	// set the BaseApp's parameter store
	a.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(a.keys[consensusparamtypes.StoreKey]), govModuleAddress, runtime.EventService{})
	bApp.SetParamStore(a.ConsensusParamsKeeper.ParamsStore)

	// grant capabilities for the ibc and ibc-transfer modules
	a.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, a.keys[capabilitytypes.StoreKey], a.memKeys[capabilitytypes.MemStoreKey])
	a.ScopedIBCKeeper = a.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	a.ScopedTransferKeeper = a.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

	// seal capability keeper after scoping modules
	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	a.CapabilityKeeper.Seal()

	// set the governance module account as the authority for conducting upgrades
	a.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, runtime.NewKVStoreService(a.keys[upgradetypes.StoreKey]), appCodec, homePath, bApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	a.AccountKeeper = authkeeper.NewAccountKeeper(appCodec, runtime.NewKVStoreService(a.keys[authtypes.StoreKey]), authtypes.ProtoBaseAccount, maccPerms, authcodec.NewBech32Codec(appparams.AccountAddressPrefix), appparams.AccountAddressPrefix, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	a.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(a.keys[banktypes.StoreKey]),
		a.AccountKeeper,
		moduleAccountAddrs,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)
	a.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(a.keys[stakingtypes.StoreKey]), a.AccountKeeper, a.BankKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(), authcodec.NewBech32Codec(appparams.Bech32PrefixValAddr), authcodec.NewBech32Codec(appparams.Bech32PrefixConsAddr),
	)
	a.MintKeeper = mintkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(a.keys[minttypes.StoreKey]), a.StakingKeeper, a.AccountKeeper, a.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	a.DistrKeeper = distrkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(a.keys[distrtypes.StoreKey]), a.AccountKeeper, a.BankKeeper, a.StakingKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	a.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, legacyAmino, runtime.NewKVStoreService(a.keys[slashingtypes.StoreKey]), a.StakingKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	a.CrisisKeeper = crisiskeeper.NewKeeper(appCodec, runtime.NewKVStoreService(a.keys[crisistypes.StoreKey]), invCheckPeriod,
		a.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(), a.AccountKeeper.AddressCodec())

	a.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(a.keys[feegrant.StoreKey]), a.AccountKeeper)

	a.AuthzKeeper = authzkeeper.NewKeeper(runtime.NewKVStoreService(a.keys[authzkeeper.StoreKey]), appCodec, bApp.MsgServiceRouter(), a.AccountKeeper)

	groupConfig := group.DefaultConfig()
	groupConfig.MaxMetadataLen = 5500
	a.GroupKeeper = groupkeeper.NewKeeper(a.keys[group.StoreKey], appCodec, bApp.MsgServiceRouter(), a.AccountKeeper, groupConfig)

	// set the governance module account as the authority for conducting upgrades
	a.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, runtime.NewKVStoreService(a.keys[upgradetypes.StoreKey]), appCodec, homePath, bApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(a.keys[evidencetypes.StoreKey]), a.StakingKeeper, a.SlashingKeeper, a.AccountKeeper.AddressCodec(), runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	a.EvidenceKeeper = *evidenceKeeper

	// Create Ethermint keepers
	a.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec,
		sdk.MustAccAddressFromBech32(govModuleAddress),
		a.ConsensusParamsKeeper,
		a.keys[feemarkettypes.StoreKey],
		a.tkeys[feemarkettypes.TransientKey],
		a.GetSubspace(feemarkettypes.ModuleName),
	)

	// Create evmos keeper
	a.EvmKeeper = evmkeeper.NewKeeper(
		appCodec,
		a.keys[evmtypes.StoreKey],
		a.tkeys[evmtypes.TransientKey],
		authtypes.NewModuleAddress(govtypes.ModuleName),
		a.AccountKeeper,
		a.BankKeeper,
		a.StakingKeeper,
		a.FeeMarketKeeper,
		nil,
		geth.NewEVM,
		tracer,
		a.GetSubspace(evmtypes.ModuleName),
	)

	// Osmosis keepers

	a.EpochsKeeper = epochskeeper.NewKeeper(
		a.keys[epochstypes.StoreKey],
	)

	gammKeeper := gammkeeper.NewKeeper(
		appCodec, a.keys[gammtypes.StoreKey],
		a.GetSubspace(gammtypes.ModuleName),
		a.AccountKeeper,
		a.BankKeeper,
		a.DistrKeeper,
	)
	a.GAMMKeeper = &gammKeeper

	a.PoolManagerKeeper = poolmanagerkeeper.NewKeeper(
		a.keys[poolmanagertypes.StoreKey],
		a.GAMMKeeper,
		a.BankKeeper,
		a.AccountKeeper,
	)

	txFeesKeeper := txfeeskeeper.NewKeeper(
		a.keys[txfeestypes.StoreKey],
		a.GetSubspace(txfeestypes.ModuleName),
		a.AccountKeeper,
		a.EpochsKeeper,
		a.BankKeeper,
		a.PoolManagerKeeper,
		a.GAMMKeeper,
		a.DistrKeeper,
	)
	a.TxFeesKeeper = &txFeesKeeper

	a.GAMMKeeper.SetPoolManager(a.PoolManagerKeeper)
	a.GAMMKeeper.SetTxFees(a.TxFeesKeeper)

	a.LockupKeeper = lockupkeeper.NewKeeper(
		a.keys[lockuptypes.StoreKey],
		appCodec,
		a.AccountKeeper,
		a.BankKeeper,
		a.TxFeesKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create IBC Keeper
	a.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		a.keys[ibcexported.StoreKey],
		a.GetSubspace(ibcexported.ModuleName),
		a.StakingKeeper,
		a.UpgradeKeeper,
		a.ScopedIBCKeeper,
		govModuleAddress,
	)

	a.RollappKeeper = rollappmodulekeeper.NewKeeper(
		appCodec,
		a.keys[rollappmoduletypes.StoreKey],
		a.IBCKeeper.ChannelKeeper,
		nil,
		a.BankKeeper,
		a.TransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		nil,
	)

	a.GAMMKeeper.SetRollapp(a.RollappKeeper)

	a.SequencerKeeper = sequencermodulekeeper.NewKeeper(
		appCodec,
		a.keys[sequencermoduletypes.StoreKey],
		a.BankKeeper,
		a.AccountKeeper,
		a.RollappKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	a.LightClientKeeper = *lightclientmodulekeeper.NewKeeper(
		appCodec,
		a.keys[lightclientmoduletypes.StoreKey],
		a.IBCKeeper.ClientKeeper,
		a.IBCKeeper.ConnectionKeeper,
		a.IBCKeeper.ChannelKeeper,
		a.SequencerKeeper,
		a.RollappKeeper,
	)

	a.SequencerKeeper.SetUnbondBlockers(a.RollappKeeper, a.LightClientKeeper)
	a.SequencerKeeper.SetHooks(sequencermoduletypes.MultiHooks{rollappmodulekeeper.SequencerHooks{Keeper: a.RollappKeeper}})

	a.RollappKeeper.SetSequencerKeeper(a.SequencerKeeper)
	a.RollappKeeper.SetCanonicalClientKeeper(a.LightClientKeeper)

	a.DenomMetadataKeeper = denommetadatamodulekeeper.NewKeeper(
		a.BankKeeper,
		a.RollappKeeper,
		nil,
	)

	a.IncentivesKeeper = incentiveskeeper.NewKeeper(
		a.keys[incentivestypes.StoreKey],
		appCodec,
		a.BankKeeper,
		a.LockupKeeper,
		a.EpochsKeeper,
		a.TxFeesKeeper,
		a.RollappKeeper,
		a.SequencerKeeper,
		&a.SponsorshipKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	a.SponsorshipKeeper = sponsorshipkeeper.NewKeeper(
		appCodec,
		a.keys[sponsorshiptypes.StoreKey],
		a.AccountKeeper,
		a.StakingKeeper,
		a.IncentivesKeeper,
		a.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	a.IROKeeper = irokeeper.NewKeeper(
		appCodec,
		a.keys[irotypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		&a.AccountKeeper,
		a.BankKeeper,
		a.DenomMetadataKeeper,
		a.RollappKeeper,
		a.GAMMKeeper,
		a.IncentivesKeeper,
		a.PoolManagerKeeper,
		a.TxFeesKeeper,
	)

	a.StreamerKeeper = *streamermodulekeeper.NewKeeper(
		appCodec,
		a.keys[streamermoduletypes.StoreKey],
		a.BankKeeper,
		a.EpochsKeeper,
		a.AccountKeeper,
		a.IncentivesKeeper,
		a.SponsorshipKeeper,
		govModuleAddress,
	)

	a.EIBCKeeper = *eibckeeper.NewKeeper(
		appCodec,
		a.keys[eibcmoduletypes.StoreKey],
		a.keys[eibcmoduletypes.MemStoreKey],
		a.AccountKeeper,
		a.BankKeeper,
		a.DelayedAckKeeper,
		a.RollappKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	a.DymNSKeeper = dymnskeeper.NewKeeper(
		appCodec,
		a.keys[dymnstypes.StoreKey],
		a.BankKeeper,
		a.RollappKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	a.RateLimitingKeeper = *ratelimitkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(a.keys[ratelimittypes.StoreKey]),
		a.GetSubspace(ratelimittypes.ModuleName),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		a.BankKeeper,
		a.IBCKeeper.ChannelKeeper,
		genesisbridge.NewICS4Wrapper(
			denommetadatamodule.NewICS4Wrapper(a.IBCKeeper.ChannelKeeper, a.RollappKeeper, a.BankKeeper),
			a.RollappKeeper,
			a.IBCKeeper.ChannelKeeper,
		), // ICS4Wrapper
	)

	// Create Transfer Keepers
	a.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		a.keys[ibctransfertypes.StoreKey],
		a.GetSubspace(ibctransfertypes.ModuleName),
		a.RateLimitingKeeper, // ICS4Wrapper
		a.IBCKeeper.ChannelKeeper,
		a.IBCKeeper.PortKeeper,
		a.AccountKeeper,
		BankKeeperWithoutSetMetadata{a.BankKeeper},
		a.ScopedTransferKeeper,
		govModuleAddress,
	)
	a.RollappKeeper.SetTransferKeeper(a.TransferKeeper)

	a.DelayedAckKeeper = *delayedackkeeper.NewKeeper(
		appCodec,
		a.keys[delayedacktypes.StoreKey],
		a.keys[ibcexported.StoreKey],
		govModuleAddress,
		a.RollappKeeper,
		a.IBCKeeper.ChannelKeeper,
		a.IBCKeeper.ChannelKeeper,
		&a.EIBCKeeper,
	)

	a.EIBCKeeper.SetDelayedAckKeeper(a.DelayedAckKeeper)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/x/gov/spec/01_concepts.md#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(a.ParamsKeeper)).
		AddRoute(denommetadatamoduletypes.RouterKey, denommetadatamodule.NewDenomMetadataProposalHandler(a.DenomMetadataKeeper)).
		AddRoute(evmtypes.RouterKey, evm.NewEvmProposalHandler(a.EvmKeeper))

	govConfig := govtypes.DefaultConfig()
	a.GovKeeper = govkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(a.keys[govtypes.StoreKey]), a.AccountKeeper, a.BankKeeper,
		a.StakingKeeper, a.DistrKeeper, bApp.MsgServiceRouter(), govConfig, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	// Set legacy router for backwards compatibility with gov v1beta1
	a.GovKeeper.SetLegacyRouter(govRouter)

	a.PacketForwardMiddlewareKeeper = packetforwardkeeper.NewKeeper(
		appCodec, a.keys[packetforwardtypes.StoreKey],
		a.TransferKeeper,
		a.IBCKeeper.ChannelKeeper,
		a.BankKeeper,
		a.IBCKeeper.ChannelKeeper,
		govModuleAddress,
	)

	a.HyperCoreKeeper = hypercorekeeper.NewKeeper(
		appCodec,
		a.AccountKeeper.AddressCodec(),
		runtime.NewKVStoreService(a.keys[hypercoretypes.ModuleName]),
		govModuleAddress,
		a.BankKeeper,
	)

	a.HyperWarpKeeper = hyperwarpkeeper.NewKeeper(
		appCodec,
		a.AccountKeeper.AddressCodec(),
		runtime.NewKVStoreService(a.keys[hyperwarptypes.ModuleName]),
		govModuleAddress,
		a.BankKeeper,
		&a.HyperCoreKeeper,
		[]int32{int32(hyperwarptypes.HYP_TOKEN_TYPE_SYNTHETIC), int32(hyperwarptypes.HYP_TOKEN_TYPE_COLLATERAL)},
	)
	a.DenomMetadataKeeper.SetWarpKeeper(&a.HyperWarpKeeper)
	a.Forward = forward.New(
		a.TransferKeeper,

		hyperwarpkeeper.NewQueryServerImpl(a.HyperWarpKeeper),
		hyperwarpkeeper.NewMsgServerImpl(a.HyperWarpKeeper),
	)

	a.KasKeeper = kaskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(a.keys[kastypes.ModuleName]),
		govModuleAddress,
		&a.HyperCoreKeeper,
	)

	a.HyperWarpKeeper.SetHook(a.Forward)

	a.DelayedAckKeeper.SetCompletionHooks(map[string]delayedackkeeper.CompletionHookInstance{
		forwardtypes.HookNameRollToHL:  a.Forward.RollToHLHook(),
		forwardtypes.HookNameRollToIBC: a.Forward.RollToIBCHook(),
	})

	// Initialize circuit breaker keeper
	a.CircuitBreakerKeeper = circuitkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(a.keys[circuittypes.StoreKey]),
		govModuleAddress,
		a.AccountKeeper.AddressCodec(),
	)
	bApp.SetCircuitBreaker(&a.CircuitBreakerKeeper)
}

func (a *AppKeepers) SetupHooks() {
	a.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			a.DistrKeeper.Hooks(),
			a.SlashingKeeper.Hooks(),
			a.SponsorshipKeeper.StakingHooks(),
		),
	)

	// register the staking hooks
	a.LockupKeeper.SetHooks(
		lockuptypes.NewMultiLockupHooks(
		// insert lockup hooks receivers here
		),
	)

	a.DenomMetadataKeeper.SetHooks(
		denommetadatamoduletypes.NewMultiDenomMetadataHooks(
			vfchooks.NewVirtualFrontierBankContractRegistrationHook(*a.EvmKeeper),
		),
	)

	a.GAMMKeeper.SetHooks(
		gammtypes.NewMultiGammHooks(
			// insert gamm hooks receivers here
			a.StreamerKeeper.Hooks(),
			a.TxFeesKeeper.Hooks(),
		),
	)

	a.IncentivesKeeper.SetHooks(
		incentivestypes.NewMultiIncentiveHooks(
		// insert incentive hooks receivers here
		),
	)

	a.DelayedAckKeeper.SetHooks(delayedacktypes.NewMultiDelayedAckHooks(
		// insert delayedAck hooks receivers here
		a.EIBCKeeper.GetDelayedAckHooks(),
	))

	a.EpochsKeeper.SetHooks(
		epochstypes.NewMultiEpochHooks(
			// insert epochs hooks receivers here
			a.StreamerKeeper.Hooks(), // x/streamer must be before x/incentives
			a.IncentivesKeeper.EpochHooks(),
			a.TxFeesKeeper.Hooks(),
			a.DelayedAckKeeper.GetEpochHooks(),
		),
	)

	a.EIBCKeeper.SetHooks(eibcmoduletypes.NewMultiEIBCHooks(
		// insert eibc hooks receivers here
		a.DelayedAckKeeper.GetEIBCHooks(),
	))

	// dependencies injected in InitTransferStack()
	a.DelayedAckMiddleware = delayedackmodule.NewIBCMiddleware()
	// register the rollapp hooks
	a.RollappKeeper.SetHooks(rollappmoduletypes.NewMultiRollappHooks(
		// insert rollapp hooks receivers here
		a.SequencerKeeper.RollappHooks(),
		a.DelayedAckKeeper,
		a.StreamerKeeper.Hooks(),
		a.DymNSKeeper.GetRollAppHooks(),
		a.LightClientKeeper.RollappHooks(),
		a.IROKeeper,
		a.DenomMetadataKeeper.RollappHooks(),
	))
}

// GetIBCKeeper implements ibctesting.TestingApp
func (a *AppKeepers) GetIBCKeeper() *ibckeeper.Keeper {
	return a.IBCKeeper
}

// GetScopedIBCKeeper implements ibctesting.TestingApp
func (a *AppKeepers) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return a.ScopedIBCKeeper
}

// GetStakingKeeper implements ibctesting.TestingApp
func (a *AppKeepers) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return a.StakingKeeper
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// ibc-go subspaces
	// register the key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())

	// ethermint subspaces (keeper doesn't load key table so we do it manually)
	paramsKeeper.Subspace(evmtypes.ModuleName).WithKeyTable(evmtypes.ParamKeyTable())
	paramsKeeper.Subspace(feemarkettypes.ModuleName).WithKeyTable(feemarkettypes.ParamKeyTable())

	// osmosis subspaces
	paramsKeeper.Subspace(gammtypes.ModuleName)
	paramsKeeper.Subspace(txfeestypes.ModuleName)

	// rate limiting subspace
	paramsKeeper.Subspace(ratelimittypes.ModuleName)

	return paramsKeeper
}

// this is a workaround to get rid of the denommetadata set automatically by ibc-go v8.x
// it has 2 issues:
// - it's not valid metadata struct
// - it has no exponent
// we disable this feature by providing bank keeper that does nothing on SetDenomMetaData
type BankKeeperWithoutSetMetadata struct {
	ibctransfertypes.BankKeeper
}

func (bk BankKeeperWithoutSetMetadata) SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata) {
}

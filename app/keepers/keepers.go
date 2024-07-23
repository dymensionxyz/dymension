package keepers

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	packetforwardmiddleware "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"
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
	lockupkeeper "github.com/osmosis-labs/osmosis/v15/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	poolmanagerkeeper "github.com/osmosis-labs/osmosis/v15/x/poolmanager/keeper"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	"github.com/dymensionxyz/dymension/v3/x/bridgingfee"
	delayedackmodule "github.com/dymensionxyz/dymension/v3/x/delayedack"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	denommetadatamodule "github.com/dymensionxyz/dymension/v3/x/denommetadata"
	denommetadatamodulekeeper "github.com/dymensionxyz/dymension/v3/x/denommetadata/keeper"
	denommetadatamoduletypes "github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	eibcmoduletypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	incentiveskeeper "github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	rollappmodule "github.com/dymensionxyz/dymension/v3/x/rollapp"
	rollappmodulekeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/transfergenesis"
	rollappmoduletypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencermodulekeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencermoduletypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	streamermodule "github.com/dymensionxyz/dymension/v3/x/streamer"
	streamermodulekeeper "github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	streamermoduletypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	vfchooks "github.com/dymensionxyz/dymension/v3/x/vfc/hooks"
)

type AppKeepers struct {
	// keepers
	AccountKeeper                 authkeeper.AccountKeeper
	AuthzKeeper                   authzkeeper.Keeper
	BankKeeper                    bankkeeper.Keeper
	CapabilityKeeper              *capabilitykeeper.Keeper
	StakingKeeper                 *stakingkeeper.Keeper
	SlashingKeeper                slashingkeeper.Keeper
	MintKeeper                    mintkeeper.Keeper
	DistrKeeper                   distrkeeper.Keeper
	GovKeeper                     *govkeeper.Keeper
	CrisisKeeper                  *crisiskeeper.Keeper
	UpgradeKeeper                 *upgradekeeper.Keeper
	ParamsKeeper                  paramskeeper.Keeper
	IBCKeeper                     *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	TransferStack                 ibcporttypes.IBCModule
	ICS4Wrapper                   ibcporttypes.ICS4Wrapper
	delayedAckMiddleware          *delayedackmodule.IBCMiddleware
	EvidenceKeeper                evidencekeeper.Keeper
	TransferKeeper                ibctransferkeeper.Keeper
	FeeGrantKeeper                feegrantkeeper.Keeper
	PacketForwardMiddlewareKeeper *packetforwardkeeper.Keeper
	ConsensusParamsKeeper         consensusparamkeeper.Keeper

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

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper

	RollappKeeper   *rollappmodulekeeper.Keeper
	SequencerKeeper sequencermodulekeeper.Keeper
	StreamerKeeper  streamermodulekeeper.Keeper
	EIBCKeeper      eibckeeper.Keeper

	DelayedAckKeeper    delayedackkeeper.Keeper
	DenomMetadataKeeper *denommetadatamodulekeeper.Keeper

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey
}

// InitKeepers initializes all keepers for the app
func (a *AppKeepers) InitKeepers(
	appCodec codec.Codec,
	cdc *codec.LegacyAmino,
	bApp *baseapp.BaseApp,
	moduleAccountAddrs map[string]bool,
	skipUpgradeHeights map[int64]bool,
	invCheckPeriod uint,
	tracer, homePath string,
) {
	govModuleAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	// init keepers

	a.ParamsKeeper = initParamsKeeper(appCodec, cdc, a.keys[paramstypes.StoreKey], a.tkeys[paramstypes.TStoreKey])
	// set the BaseApp's parameter store
	a.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, a.keys[consensusparamtypes.StoreKey], govModuleAddress)
	bApp.SetParamStore(&a.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	a.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, a.keys[capabilitytypes.StoreKey], a.memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	a.ScopedIBCKeeper = a.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	a.ScopedTransferKeeper = a.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

	a.CapabilityKeeper.Seal()

	a.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec, a.keys[crisistypes.StoreKey], invCheckPeriod, a.BankKeeper, authtypes.FeeCollectorName, govModuleAddress,
	)

	a.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		a.keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		bApp,
		govModuleAddress,
	)

	a.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		a.keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		govModuleAddress,
	)

	a.AuthzKeeper = authzkeeper.NewKeeper(
		a.keys[authz.ModuleName],
		appCodec,
		bApp.MsgServiceRouter(),
		a.AccountKeeper,
	)

	a.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		a.keys[banktypes.StoreKey],
		a.AccountKeeper,
		moduleAccountAddrs,
		govModuleAddress,
	)

	a.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		a.keys[stakingtypes.StoreKey],
		a.AccountKeeper,
		a.BankKeeper,
		govModuleAddress,
	)

	a.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		a.keys[minttypes.StoreKey],
		a.StakingKeeper,
		a.AccountKeeper,
		a.BankKeeper,
		authtypes.FeeCollectorName,
		govModuleAddress,
	)

	a.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		a.keys[distrtypes.StoreKey],
		a.AccountKeeper,
		a.BankKeeper,
		a.StakingKeeper,
		authtypes.FeeCollectorName,
		govModuleAddress,
	)

	a.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		cdc,
		a.keys[slashingtypes.StoreKey],
		a.StakingKeeper,
		govModuleAddress,
	)

	a.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		a.keys[feegrant.StoreKey],
		a.AccountKeeper,
	)

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

	a.LockupKeeper = lockupkeeper.NewKeeper(
		a.keys[lockuptypes.StoreKey],
		a.GetSubspace(lockuptypes.ModuleName),
		a.AccountKeeper,
		a.BankKeeper,
	)

	a.EpochsKeeper = epochskeeper.NewKeeper(
		a.keys[epochstypes.StoreKey],
	)

	gammKeeper := gammkeeper.NewKeeper(
		appCodec, a.keys[gammtypes.StoreKey],
		a.GetSubspace(gammtypes.ModuleName),
		a.AccountKeeper,
		a.BankKeeper, a.DistrKeeper,
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
	)
	a.TxFeesKeeper = &txFeesKeeper

	a.GAMMKeeper.SetPoolManager(a.PoolManagerKeeper)
	a.GAMMKeeper.SetTxFees(a.TxFeesKeeper)

	// Create IBC Keeper
	a.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		a.keys[ibcexported.StoreKey],
		a.GetSubspace(ibcexported.ModuleName),
		a.StakingKeeper,
		a.UpgradeKeeper,
		a.ScopedIBCKeeper,
	)

	a.DenomMetadataKeeper = denommetadatamodulekeeper.NewKeeper(
		a.BankKeeper,
	)

	a.RollappKeeper = rollappmodulekeeper.NewKeeper(
		appCodec,
		a.keys[rollappmoduletypes.StoreKey],
		a.GetSubspace(rollappmoduletypes.ModuleName),
		a.IBCKeeper.ChannelKeeper,
		a.IBCKeeper.ClientKeeper,
	)

	a.SequencerKeeper = *sequencermodulekeeper.NewKeeper(
		appCodec,
		a.keys[sequencermoduletypes.StoreKey],
		a.GetSubspace(sequencermoduletypes.ModuleName),
		a.BankKeeper,
		a.RollappKeeper,
	)

	a.IncentivesKeeper = incentiveskeeper.NewKeeper(
		a.keys[incentivestypes.StoreKey],
		a.GetSubspace(incentivestypes.ModuleName),
		a.BankKeeper,
		a.LockupKeeper,
		a.EpochsKeeper,
		a.DistrKeeper,
		a.TxFeesKeeper,
		a.RollappKeeper,
		&a.SequencerKeeper,
	)

	a.StreamerKeeper = *streamermodulekeeper.NewKeeper(
		a.keys[streamermoduletypes.StoreKey],
		a.GetSubspace(streamermoduletypes.ModuleName),
		a.BankKeeper,
		a.EpochsKeeper,
		a.AccountKeeper,
		a.IncentivesKeeper,
	)

	a.EIBCKeeper = *eibckeeper.NewKeeper(
		appCodec,
		a.keys[eibcmoduletypes.StoreKey],
		a.keys[eibcmoduletypes.MemStoreKey],
		a.GetSubspace(eibcmoduletypes.ModuleName),
		a.AccountKeeper,
		a.BankKeeper,
		nil,
	)

	// Create Transfer Keepers
	a.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		a.keys[ibctransfertypes.StoreKey],
		a.GetSubspace(ibctransfertypes.ModuleName),
		denommetadatamodule.NewICS4Wrapper(a.IBCKeeper.ChannelKeeper, a.RollappKeeper, a.BankKeeper),
		a.IBCKeeper.ChannelKeeper,
		&a.IBCKeeper.PortKeeper,
		a.AccountKeeper,
		a.BankKeeper,
		a.ScopedTransferKeeper,
	)

	a.DelayedAckKeeper = *delayedackkeeper.NewKeeper(
		appCodec,
		a.keys[delayedacktypes.StoreKey],
		a.GetSubspace(delayedacktypes.ModuleName),
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
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(a.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(a.IBCKeeper.ClientKeeper)).
		AddRoute(streamermoduletypes.RouterKey, streamermodule.NewStreamerProposalHandler(a.StreamerKeeper)).
		AddRoute(rollappmoduletypes.RouterKey, rollappmodule.NewRollappProposalHandler(a.RollappKeeper)).
		AddRoute(denommetadatamoduletypes.RouterKey, denommetadatamodule.NewDenomMetadataProposalHandler(a.DenomMetadataKeeper)).
		AddRoute(evmtypes.RouterKey, evm.NewEvmProposalHandler(a.EvmKeeper))

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	// If evidence needs to be handled for the app, set routes in router here and seal
	a.EvidenceKeeper = *evidencekeeper.NewKeeper(
		appCodec, a.keys[evidencetypes.StoreKey], a.StakingKeeper, a.SlashingKeeper,
	)

	govConfig := govtypes.DefaultConfig()
	a.GovKeeper = govkeeper.NewKeeper(
		appCodec, a.keys[govtypes.StoreKey], a.AccountKeeper, a.BankKeeper,
		a.StakingKeeper, bApp.MsgServiceRouter(), govConfig, govModuleAddress,
	)
	a.GovKeeper.SetLegacyRouter(govRouter)

	a.PacketForwardMiddlewareKeeper = packetforwardkeeper.NewKeeper(
		appCodec, a.keys[packetforwardtypes.StoreKey],
		a.TransferKeeper,
		a.IBCKeeper.ChannelKeeper,
		a.DistrKeeper,
		a.BankKeeper,
		a.IBCKeeper.ChannelKeeper,
		govModuleAddress,
	)
}

func (a *AppKeepers) InitTransferStack() {
	a.TransferStack = ibctransfer.NewIBCModule(a.TransferKeeper)
	a.TransferStack = bridgingfee.NewIBCModule(
		a.TransferStack.(ibctransfer.IBCModule),
		a.DelayedAckKeeper,
		a.TransferKeeper,
		a.AccountKeeper.GetModuleAddress(txfeestypes.ModuleName),
		*a.RollappKeeper,
	)
	a.TransferStack = packetforwardmiddleware.NewIBCMiddleware(
		a.TransferStack,
		a.PacketForwardMiddlewareKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)

	a.TransferStack = denommetadatamodule.NewIBCModule(a.TransferStack, a.DenomMetadataKeeper, a.RollappKeeper)
	// already instantiated in SetupHooks()
	a.delayedAckMiddleware.Setup(
		delayedackmodule.WithIBCModule(a.TransferStack),
		delayedackmodule.WithKeeper(a.DelayedAckKeeper),
		delayedackmodule.WithRollappKeeper(a.RollappKeeper),
	)
	a.TransferStack = a.delayedAckMiddleware
	a.TransferStack = transfergenesis.NewIBCModule(a.TransferStack, a.DelayedAckKeeper, *a.RollappKeeper, a.TransferKeeper, a.DenomMetadataKeeper)
	a.TransferStack = transfergenesis.NewIBCModuleCanonicalChannelHack(a.TransferStack, *a.RollappKeeper, a.IBCKeeper.ChannelKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, a.TransferStack)
	a.IBCKeeper.SetRouter(ibcRouter)
}

func (a *AppKeepers) SetupHooks() {
	a.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(a.DistrKeeper.Hooks(), a.SlashingKeeper.Hooks()),
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
			a.IncentivesKeeper.Hooks(),
			a.StreamerKeeper.Hooks(),
			a.TxFeesKeeper.Hooks(),
			a.DelayedAckKeeper.GetEpochHooks(),
		),
	)

	a.EIBCKeeper.SetHooks(eibcmoduletypes.NewMultiEIBCHooks(
		// insert eibc hooks receivers here
		a.DelayedAckKeeper.GetEIBCHooks(),
	))

	// dependencies injected in InitTransferStack()
	a.delayedAckMiddleware = delayedackmodule.NewIBCMiddleware()
	// register the rollapp hooks
	a.RollappKeeper.SetHooks(rollappmoduletypes.NewMultiRollappHooks(
		// insert rollapp hooks receivers here
		a.SequencerKeeper.RollappHooks(),
		a.delayedAckMiddleware,
		a.StreamerKeeper.Hooks(),
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

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(rollappmoduletypes.ModuleName)
	paramsKeeper.Subspace(sequencermoduletypes.ModuleName)
	paramsKeeper.Subspace(streamermoduletypes.ModuleName)
	paramsKeeper.Subspace(denommetadatamoduletypes.ModuleName)
	paramsKeeper.Subspace(delayedacktypes.ModuleName)
	paramsKeeper.Subspace(eibcmoduletypes.ModuleName)

	// ethermint subspaces
	paramsKeeper.Subspace(evmtypes.ModuleName)
	paramsKeeper.Subspace(feemarkettypes.ModuleName)

	// osmosis subspaces
	paramsKeeper.Subspace(lockuptypes.ModuleName)
	paramsKeeper.Subspace(epochstypes.ModuleName)
	paramsKeeper.Subspace(gammtypes.ModuleName)
	paramsKeeper.Subspace(incentivestypes.ModuleName)
	paramsKeeper.Subspace(txfeestypes.ModuleName)

	return paramsKeeper
}

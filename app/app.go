package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simapp "github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/streaming"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"

	ibctransfer "github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v6/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v6/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v6/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	ibctestingtypes "github.com/cosmos/ibc-go/v6/testing/types"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"

	ante "github.com/dymensionxyz/dymension/app/ante"
	"github.com/dymensionxyz/dymension/app/fraudproof"
	appparams "github.com/dymensionxyz/dymension/app/params"
	rollappmodule "github.com/dymensionxyz/dymension/x/rollapp"
	rollappmodulekeeper "github.com/dymensionxyz/dymension/x/rollapp/keeper"
	rollappmoduletypes "github.com/dymensionxyz/dymension/x/rollapp/types"

	sequencermodule "github.com/dymensionxyz/dymension/x/sequencer"
	sequencermodulekeeper "github.com/dymensionxyz/dymension/x/sequencer/keeper"
	sequencermoduletypes "github.com/dymensionxyz/dymension/x/sequencer/types"

	streamermodule "github.com/dymensionxyz/dymension/x/streamer"
	streamermoduleclient "github.com/dymensionxyz/dymension/x/streamer/client"
	streamermodulekeeper "github.com/dymensionxyz/dymension/x/streamer/keeper"
	streamermoduletypes "github.com/dymensionxyz/dymension/x/streamer/types"

	denommetadatamodule "github.com/dymensionxyz/dymension/x/denommetadata"

	delayedackmodule "github.com/dymensionxyz/dymension/x/delayedack"
	delayedackkeeper "github.com/dymensionxyz/dymension/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/x/delayedack/types"

	packetforwardmiddleware "github.com/strangelove-ventures/packet-forward-middleware/v6/router"
	packetforwardkeeper "github.com/strangelove-ventures/packet-forward-middleware/v6/router/keeper"
	packetforwardtypes "github.com/strangelove-ventures/packet-forward-middleware/v6/router/types"

	/* ------------------------------ ethermint imports ----------------------------- */

	"github.com/evmos/evmos/v12/ethereum/eip712"

	evmante "github.com/evmos/evmos/v12/app/ante"

	"github.com/evmos/evmos/v12/server/flags"
	ethermint "github.com/evmos/evmos/v12/types"
	"github.com/evmos/evmos/v12/x/evm"
	evmkeeper "github.com/evmos/evmos/v12/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
	"github.com/evmos/evmos/v12/x/feemarket"
	feemarketkeeper "github.com/evmos/evmos/v12/x/feemarket/keeper"
	feemarkettypes "github.com/evmos/evmos/v12/x/feemarket/types"

	/* ----------------------------- osmosis imports ---------------------------- */

	"github.com/osmosis-labs/osmosis/v15/x/epochs"
	epochskeeper "github.com/osmosis-labs/osmosis/v15/x/epochs/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v15/x/lockup"
	lockupkeeper "github.com/osmosis-labs/osmosis/v15/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	incentives "github.com/osmosis-labs/osmosis/v15/x/incentives"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v15/x/incentives/keeper"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	poolmanagerkeeper "github.com/osmosis-labs/osmosis/v15/x/poolmanager/keeper"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	txfees "github.com/osmosis-labs/osmosis/v15/x/txfees"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

var (
	_ = packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp
	_ = packetforwardmiddleware.AppModule{}
	_ = packetforwardtypes.ErrIntOverflowGenesis
)

// this line is used by starport scaffolding # stargate/wasm/app/enabledProposals

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	// this line is used by starport scaffolding # stargate/app/govProposalHandlers

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		distrclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		streamermoduleclient.CreateStreamHandler,
		streamermoduleclient.TerminateStreamHandler,
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		rollappmodule.AppModuleBasic{},
		sequencermodule.AppModuleBasic{},
		streamermodule.AppModuleBasic{},
		packetforwardmiddleware.AppModuleBasic{},
		delayedackmodule.AppModuleBasic{},
		// this line is used by starport scaffolding # stargate/app/moduleBasic

		// Ethermint modules
		evm.AppModuleBasic{},
		feemarket.AppModuleBasic{},

		// Osmosis modules
		lockup.AppModuleBasic{},
		epochs.AppModuleBasic{},
		gamm.AppModuleBasic{},
		poolmanager.AppModuleBasic{},
		incentives.AppModuleBasic{},
		txfees.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:      nil,
		distrtypes.ModuleName:           nil,
		minttypes.ModuleName:            {authtypes.Minter},
		stakingtypes.BondedPoolName:     {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:  {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:             {authtypes.Burner},
		ibctransfertypes.ModuleName:     {authtypes.Minter, authtypes.Burner},
		sequencermoduletypes.ModuleName: {authtypes.Minter, authtypes.Burner, authtypes.Staking},
		streamermoduletypes.ModuleName:  nil,
		// this line is used by starport scaffolding # stargate/app/maccPerms

		evmtypes.ModuleName:        {authtypes.Minter, authtypes.Burner}, // used for secure addition and subtraction of balance using module account
		gammtypes.ModuleName:       {authtypes.Minter, authtypes.Burner},
		lockuptypes.ModuleName:     {authtypes.Minter, authtypes.Burner},
		incentivestypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		txfeestypes.ModuleName:     {authtypes.Burner},
	}
)

var (
	_ servertypes.Application = (*App)(nil)
	_ simapp.App              = (*App)(nil)
	_ ibctesting.TestingApp   = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+appparams.Name)

	sdk.DefaultPowerReduction = ethermint.PowerReduction
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper                 authkeeper.AccountKeeper
	AuthzKeeper                   authzkeeper.Keeper
	BankKeeper                    bankkeeper.Keeper
	CapabilityKeeper              *capabilitykeeper.Keeper
	StakingKeeper                 stakingkeeper.Keeper
	SlashingKeeper                slashingkeeper.Keeper
	MintKeeper                    mintkeeper.Keeper
	DistrKeeper                   distrkeeper.Keeper
	GovKeeper                     govkeeper.Keeper
	CrisisKeeper                  crisiskeeper.Keeper
	UpgradeKeeper                 upgradekeeper.Keeper
	ParamsKeeper                  paramskeeper.Keeper
	IBCKeeper                     *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper                evidencekeeper.Keeper
	TransferKeeper                ibctransferkeeper.Keeper
	FeeGrantKeeper                feegrantkeeper.Keeper
	PacketForwardMiddlewareKeeper *packetforwardkeeper.Keeper

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

	RollappKeeper   rollappmodulekeeper.Keeper
	SequencerKeeper sequencermodulekeeper.Keeper
	StreamerKeeper  streamermodulekeeper.Keeper

	// this line is used by starport scaffolding # stargate/app/keeperDeclaration
	DelayedAckKeeper delayedackkeeper.Keeper

	// the module manager
	mm *module.Manager

	// module configurator
	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Codec
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	eip712.SetEncodingConfig(simappparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             appCodec,
		TxConfig:          encodingConfig.TxConfig,
		Amino:             encodingConfig.Amino,
	})

	bApp := baseapp.NewBaseApp(appparams.Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, authzkeeper.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
		rollappmoduletypes.StoreKey,
		sequencermoduletypes.StoreKey,
		streamermoduletypes.StoreKey,
		packetforwardtypes.StoreKey,
		delayedacktypes.StoreKey,
		// this line is used by starport scaffolding # stargate/app/storeKey

		// ethermint keys
		evmtypes.StoreKey, feemarkettypes.StoreKey,

		// osmosis keys
		lockuptypes.StoreKey,
		epochstypes.StoreKey,
		gammtypes.StoreKey,
		poolmanagertypes.StoreKey,
		incentivestypes.StoreKey,
		txfeestypes.StoreKey,
	)

	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey, evmtypes.TransientKey, feemarkettypes.TransientKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	// load state streaming if enabled
	if _, _, err := streaming.LoadStreamingServices(bApp, appOpts, appCodec, keys); err != nil {
		panic("failed to load state streaming services: " + err.Error())
	}

	app := &App{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	// bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	// this line is used by starport scaffolding # stargate/app/scopedKeeper

	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, keys[authtypes.StoreKey], app.GetSubspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authz.ModuleName], appCodec, app.MsgServiceRouter(), app.AccountKeeper,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, keys[banktypes.StoreKey], app.AccountKeeper, app.GetSubspace(banktypes.ModuleName), app.ModuleAccountAddrs(),
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName),
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], app.GetSubspace(minttypes.ModuleName), &stakingKeeper,
		app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, authtypes.FeeCollectorName,
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], &stakingKeeper, app.GetSubspace(slashingtypes.ModuleName),
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, app.BaseApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	// Create Ethermint keepers
	app.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec, authtypes.NewModuleAddress(govtypes.ModuleName),
		keys[feemarkettypes.StoreKey], tkeys[feemarkettypes.TransientKey], app.GetSubspace(feemarkettypes.ModuleName),
	)

	// Create evmos keeper
	tracer := cast.ToString(appOpts.Get(flags.EVMTracer))
	app.EvmKeeper = evmkeeper.NewKeeper(
		appCodec, keys[evmtypes.StoreKey], tkeys[evmtypes.TransientKey], authtypes.NewModuleAddress(govtypes.ModuleName),
		app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.FeeMarketKeeper,
		tracer, app.GetSubspace(evmtypes.ModuleName),
	)

	// Osmosis keepers

	app.LockupKeeper = lockupkeeper.NewKeeper(
		app.keys[lockuptypes.StoreKey],
		app.GetSubspace(lockuptypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
	)

	app.LockupKeeper.SetHooks(
		lockuptypes.NewMultiLockupHooks(
		// insert lockup hooks receivers here
		),
	)

	app.EpochsKeeper = epochskeeper.NewKeeper(app.keys[epochstypes.StoreKey])

	gammKeeper := gammkeeper.NewKeeper(
		appCodec, keys[gammtypes.StoreKey],
		app.GetSubspace(gammtypes.ModuleName),
		app.AccountKeeper,
		// TODO: Add a mintcoins restriction
		app.BankKeeper, app.DistrKeeper)
	app.GAMMKeeper = &gammKeeper

	app.PoolManagerKeeper = poolmanagerkeeper.NewKeeper(
		keys[poolmanagertypes.StoreKey],
		app.GAMMKeeper,
		app.BankKeeper,
		app.AccountKeeper,
	)

	txfeeskeeper := txfeeskeeper.NewKeeper(
		app.keys[txfeestypes.StoreKey],
		app.GetSubspace(txfeestypes.ModuleName),
		app.AccountKeeper,
		app.EpochsKeeper,
		app.BankKeeper,
		app.PoolManagerKeeper,
		app.GAMMKeeper,
	)
	app.TxFeesKeeper = &txfeeskeeper
	app.GAMMKeeper.SetPoolManager(app.PoolManagerKeeper)
	app.GAMMKeeper.SetTxFees(app.TxFeesKeeper)

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		app.GetSubspace(ibchost.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
	)

	app.IncentivesKeeper = incentiveskeeper.NewKeeper(
		app.keys[incentivestypes.StoreKey],
		app.GetSubspace(incentivestypes.ModuleName),
		app.BankKeeper,
		app.LockupKeeper,
		app.EpochsKeeper,
		app.DistrKeeper,
		app.TxFeesKeeper,
	)

	app.RollappKeeper = *rollappmodulekeeper.NewKeeper(
		appCodec,
		keys[rollappmoduletypes.StoreKey],
		keys[rollappmoduletypes.MemStoreKey],
		app.GetSubspace(rollappmoduletypes.ModuleName),
	)

	app.SequencerKeeper = *sequencermodulekeeper.NewKeeper(
		appCodec,
		keys[sequencermoduletypes.StoreKey],
		keys[sequencermoduletypes.MemStoreKey],
		app.GetSubspace(sequencermoduletypes.ModuleName),

		app.BankKeeper,
		app.RollappKeeper,
	)

	app.StreamerKeeper = *streamermodulekeeper.NewKeeper(
		keys[streamermoduletypes.StoreKey],
		app.GetSubspace(streamermoduletypes.ModuleName),

		app.BankKeeper,
		app.EpochsKeeper,
		app.AccountKeeper,
		app.IncentivesKeeper,
	)

	app.DelayedAckKeeper = *delayedackkeeper.NewKeeper(
		appCodec,
		keys[delayedacktypes.StoreKey],
		keys[delayedacktypes.MemStoreKey],
		app.RollappKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.IBCKeeper.ClientKeeper,
	)

	/* -------------------------------- set hooks ------------------------------- */
	// Set hooks
	app.GAMMKeeper.SetHooks(
		gammtypes.NewMultiGammHooks(
			// insert gamm hooks receivers here
			app.StreamerKeeper.Hooks(),
			app.TxFeesKeeper.Hooks(),
		),
	)

	app.IncentivesKeeper.SetHooks(
		incentivestypes.NewMultiIncentiveHooks(
		// insert incentive hooks receivers here
		),
	)
	app.EpochsKeeper.SetHooks(
		epochstypes.NewMultiEpochHooks(
			// insert epochs hooks receivers here
			app.IncentivesKeeper.Hooks(),
			app.StreamerKeeper.Hooks(),
			app.TxFeesKeeper.Hooks(),
		),
	)

	sequencerModule := sequencermodule.NewAppModule(appCodec, app.SequencerKeeper, app.AccountKeeper, app.BankKeeper)
	rollappModule := rollappmodule.NewAppModule(appCodec, &app.RollappKeeper, app.AccountKeeper, app.BankKeeper)
	streamerModule := streamermodule.NewAppModule(app.StreamerKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper)
	delayedackModule := delayedackmodule.NewAppModule(appCodec, app.DelayedAckKeeper)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/x/gov/spec/01_concepts.md#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(streamermoduletypes.RouterKey, streamermodule.NewStreamerProposalHandler(app.StreamerKeeper))

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
	)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], &app.StakingKeeper, app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	govConfig := govtypes.DefaultConfig()
	app.GovKeeper = govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, govRouter, app.MsgServiceRouter(), govConfig,
	)

	// this line is used by starport scaffolding # stargate/app/keeperDefinition

	app.PacketForwardMiddlewareKeeper = packetforwardkeeper.NewKeeper(
		appCodec, keys[packetforwardtypes.StoreKey],
		app.GetSubspace(packetforwardtypes.ModuleName),
		app.TransferKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.DistrKeeper,
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
	)

	transferModule := ibctransfer.NewAppModule(app.TransferKeeper)

	var transferStack ibcporttypes.IBCModule
	transferStack = ibctransfer.NewIBCModule(app.TransferKeeper)
	transferStack = packetforwardmiddleware.NewIBCMiddleware(transferStack, app.PacketForwardMiddlewareKeeper, 0, packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp, packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp)
	transferStack = denommetadatamodule.NewIBCMiddleware(transferStack, app.IBCKeeper.ChannelKeeper, app.TransferKeeper, app.RollappKeeper, app.BankKeeper)
	transferStack = delayedackmodule.NewIBCMiddleware(transferStack, app.DelayedAckKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)
	// this line is used by starport scaffolding # ibc/app/router
	app.IBCKeeper.SetRouter(ibcRouter)

	// register the rollapp hooks
	app.RollappKeeper.SetHooks(rollappmoduletypes.NewMultiRollappHooks(
		// insert rollapp hooks receivers here
		app.SequencerKeeper.RollappHooks(),
		transferStack.(delayedackmodule.IBCMiddleware),
	))

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		packetforwardmiddleware.NewAppModule(app.PacketForwardMiddlewareKeeper),
		transferModule,
		rollappModule,
		sequencerModule,
		streamerModule,
		delayedackModule,
		// this line is used by starport scaffolding # stargate/app/appModule

		// Ethermint app modules
		evm.NewAppModule(app.EvmKeeper, app.AccountKeeper, app.GetSubspace(evmtypes.ModuleName).WithKeyTable(evmtypes.ParamKeyTable())),
		feemarket.NewAppModule(app.FeeMarketKeeper, app.GetSubspace(feemarkettypes.ModuleName).WithKeyTable(feemarkettypes.ParamKeyTable())),

		// osmosis modules
		lockup.NewAppModule(*app.LockupKeeper, app.AccountKeeper, app.BankKeeper),
		epochs.NewAppModule(*app.EpochsKeeper),
		gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		poolmanager.NewAppModule(*app.PoolManagerKeeper, app.GAMMKeeper),
		incentives.NewAppModule(*app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		txfees.NewAppModule(*app.TxFeesKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		epochstypes.ModuleName,
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		feemarkettypes.ModuleName,
		evmtypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		packetforwardtypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		rollappmoduletypes.ModuleName,
		sequencermoduletypes.ModuleName,
		streamermoduletypes.ModuleName,
		delayedacktypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
		lockuptypes.ModuleName,
		gammtypes.ModuleName,
		poolmanagertypes.ModuleName,
		incentivestypes.ModuleName,
		txfeestypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		feemarkettypes.ModuleName,
		evmtypes.ModuleName,
		slashingtypes.ModuleName,
		vestingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		packetforwardtypes.ModuleName,
		rollappmoduletypes.ModuleName,
		sequencermoduletypes.ModuleName,
		streamermoduletypes.ModuleName,
		delayedacktypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/endBlockers
		epochstypes.ModuleName,
		lockuptypes.ModuleName,
		gammtypes.ModuleName,
		poolmanagertypes.ModuleName,
		incentivestypes.ModuleName,
		txfeestypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		slashingtypes.ModuleName,
		feemarkettypes.ModuleName,
		evmtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		packetforwardtypes.ModuleName,
		feegrant.ModuleName,
		rollappmoduletypes.ModuleName,
		sequencermoduletypes.ModuleName,
		streamermoduletypes.ModuleName,
		delayedacktypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/initGenesis

		epochstypes.ModuleName,
		lockuptypes.ModuleName,
		gammtypes.ModuleName,
		poolmanagertypes.ModuleName,
		incentivestypes.ModuleName,
		txfeestypes.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	maxGasWanted := cast.ToUint64(appOpts.Get(flags.EVMMaxTxGasWanted))
	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:          &app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		IBCKeeper:              app.IBCKeeper,
		FeeMarketKeeper:        app.FeeMarketKeeper,
		StakingKeeper:          app.StakingKeeper,
		DistributionKeeper:     app.DistrKeeper,
		EvmKeeper:              app.EvmKeeper,
		FeegrantKeeper:         app.FeeGrantKeeper,
		TxFeesKeeper:           app.TxFeesKeeper,
		SignModeHandler:        encodingConfig.TxConfig.SignModeHandler(),
		MaxTxGasWanted:         maxGasWanted,
		SigGasConsumer:         evmante.SigVerificationGasConsumer,
		ExtensionOptionChecker: nil,
	})
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	//FraudProof verifier
	// var fraudProofVerifier fraudproof.FraudProofVerifier = nil
	fraudProofVerifier := fraudproof.New(bApp, "rollapp_fraudproof", logger)
	app.RollappKeeper.SetFraudProofVerifier(fraudProofVerifier)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	// this line is used by starport scaffolding # stargate/app/beforeInitReturn

	return app
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	//exclude the streamer and txfees modules
	modAccAddrs[authtypes.NewModuleAddress(streamermoduletypes.ModuleName).String()] = false
	modAccAddrs[authtypes.NewModuleAddress(txfeestypes.ModuleName).String()] = false
	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
	HealthcheckRegister(clientCtx, apiSvr.Router)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *App) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
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
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(rollappmoduletypes.ModuleName)
	paramsKeeper.Subspace(sequencermoduletypes.ModuleName)
	paramsKeeper.Subspace(streamermoduletypes.ModuleName)
	// this line is used by starport scaffolding # stargate/app/paramSubspace

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

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return nil
}

// GetIBCKeeper implements ibctesting.TestingApp
func (app *App) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements ibctesting.TestingApp
func (app *App) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetStakingKeeper implements ibctesting.TestingApp
func (app *App) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetTxConfig implements ibctesting.TestingApp
func (app *App) GetTxConfig() client.TxConfig {
	return simappparams.MakeTestEncodingConfig().TxConfig
}

func (app *App) ExportState(ctx sdk.Context) map[string]json.RawMessage {
	return app.mm.ExportGenesis(ctx, app.AppCodec())
}

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"

	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/cosmos/cosmos-sdk/runtime"

	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	denommetadatamoduleclient "github.com/dymensionxyz/dymension/v3/x/denommetadata/client"
	dymnsmoduleclient "github.com/dymensionxyz/dymension/v3/x/dymns/client"
	sequencermoduleclient "github.com/dymensionxyz/dymension/v3/x/sequencer/client"

	v5 "github.com/dymensionxyz/dymension/v3/app/upgrades/v5"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/x/crisis"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/dymensionxyz/dymension/v3/app/ante"
	"github.com/dymensionxyz/dymension/v3/app/params"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"

	/* ------------------------------ ethermint imports ----------------------------- */

	"github.com/evmos/ethermint/server/flags"
	ethermint "github.com/evmos/ethermint/types"
	evmclient "github.com/evmos/ethermint/x/evm/client"
	/* ----------------------------- osmosis imports ---------------------------- */ /* ---------------------------- upgrade handlers ---------------------------- */)

var (
	_ servertypes.Application = (*App)(nil)
	_ runtime.AppI            = (*App)(nil)
	_ ibctesting.TestingApp   = (*App)(nil)

	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// Upgrades contains the upgrade handlers for the application
	Upgrades = []upgrades.Upgrade{v5.Upgrade}
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

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// keepers
	AppKeepers
	// the module manager
	mm                 *module.Manager
	BasicModuleManager module.BasicManager

	// module configurator
	configurator module.Configurator
	// simulation manager
	sm *module.SimulationManager
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	encoding := params.MakeEncodingConfig()
	appCodec := encoding.Codec
	legacyAmino := encoding.Amino
	txConfig := encoding.TxConfig
	interfaceRegistry := encoding.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appparams.Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		AppKeepers:        AppKeepers{},
	}

	app.AppKeepers.GenerateKeys()

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, app.keys); err != nil {
		panic(fmt.Errorf("failed to register streaming services: %w", err))
	}

	app.AppKeepers.InitKeepers(appCodec, legacyAmino, bApp, logger, ModuleAccountAddrs(), appOpts)
	app.AppKeepers.SetupHooks()
	app.AppKeepers.InitTransferStack()

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(app.SetupModules(appCodec, bApp, skipGenesisInvariants)...)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.mm,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
					sequencermoduleclient.PunishSequencerHandler,
					denommetadatamoduleclient.CreateDenomMetadataHandler,
					denommetadatamoduleclient.UpdateDenomMetadataHandler,
					dymnsmoduleclient.MigrateChainIdsProposalHandler,
					dymnsmoduleclient.UpdateAliasesProposalHandler,
					evmclient.UpdateVirtualFrontierBankContractProposalHandler,
				},
			),
		})

	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	// NOTE: upgrade module is required to be prioritized
	app.mm.SetOrderPreBlockers(PreBlockers...)
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// TODO: use "github.com/osmosis-labs/osmosis/osmoutils/partialord" to order modules
	app.mm.SetOrderBeginBlockers(BeginBlockers...)
	app.mm.SetOrderEndBlockers(EndBlockers...)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(InitGenesis...)
	app.mm.SetOrderExportGenesis(InitGenesis...)

	// Uncomment if you want to set a custom migration order here.
	// app.mm.SetOrderMigrations(custom order)

	app.mm.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err := app.mm.RegisterServices(app.configurator)
	if err != nil {
		panic(fmt.Errorf("failed to register services: %w", err))
	}

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.ModuleManager` and `app.configurator` are set.
	app.setupUpgradeHandlers()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(fmt.Errorf("failed to create reflection service: %w", err))
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	/****  Simulations ****/
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.mm.Modules, overrideModules)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(KVStoreKeys)
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	/* ---------------------------- set ante handler ---------------------------- */
	maxGasWanted := cast.ToUint64(appOpts.Get(flags.EVMMaxTxGasWanted))
	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		ExtensionOptionChecker: nil, // uses default
		FeegrantKeeper:         app.FeeGrantKeeper,
		SignModeHandler:        txConfig.SignModeHandler(),
		SigGasConsumer:         nil,
		TxFeeChecker:           nil,
		IBCKeeper:              app.IBCKeeper,
		FeeMarketKeeper:        app.FeeMarketKeeper,
		EvmKeeper:              app.EvmKeeper,
		TxFeesKeeper:           app.TxFeesKeeper,
		MaxTxGasWanted:         maxGasWanted,
		RollappKeeper:          *app.RollappKeeper,
		LightClientKeeper:      &app.LightClientKeeper,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create ante handler: %w", err))
	}

	app.SetAnteHandler(anteHandler)

	// At startup, after all modules have been registered, check that all proto
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return app
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

func (a *App) Configurator() module.Configurator {
	return a.configurator
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	if err != nil {
		panic(err)
	}
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *App) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns SimApp's TxConfig
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// AutoCliOpts returns the autocli options for the app.
func (app *App) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.mm.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *App) DefaultGenesis() map[string]json.RawMessage {
	return a.BasicModuleManager.DefaultGenesis(a.appCodec)
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new CometBFT queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
	HealthcheckRegister(clientCtx, apiSvr.Router)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		cmtApp.Query,
	)
}

func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

func (app *App) setupUpgradeHandlers() {
	for _, u := range Upgrades {
		app.setupUpgradeHandler(u)
	}
}

func (app *App) setupUpgradeHandler(upgrade upgrades.Upgrade) {
	app.UpgradeKeeper.SetUpgradeHandler(
		upgrade.Name,
		upgrade.CreateHandler(
			app.mm,
			app.configurator,
			&upgrades.UpgradeKeepers{
				LockupKeeper:      app.AppKeepers.LockupKeeper,
				IROKeeper:         app.AppKeepers.IROKeeper,
				GAMMKeeper:        app.AppKeepers.GAMMKeeper,
				GovKeeper:         app.AppKeepers.GovKeeper,
				IncentivesKeeper:  app.AppKeepers.IncentivesKeeper,
				RollappKeeper:     app.AppKeepers.RollappKeeper,
				SponsorshipKeeper: &app.AppKeepers.SponsorshipKeeper,
				ParamsKeeper:      &app.AppKeepers.ParamsKeeper,
				DelayedAckKeeper:  &app.AppKeepers.DelayedAckKeeper,
				EIBCKeeper:        &app.AppKeepers.EIBCKeeper,
				DymNSKeeper:       &app.AppKeepers.DymNSKeeper,
				StreamerKeeper:    &app.AppKeepers.StreamerKeeper,
			},
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}

	if upgradeInfo.Name == upgrade.Name && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// configure store loader with the store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
	}
}

func (app *App) ExportState(ctx sdk.Context) map[string]json.RawMessage {
	export, err := app.mm.ExportGenesis(ctx, app.AppCodec())
	if err != nil {
		panic(err)
	}
	return export
}

/* -------------------------------------------------------------------------- */
/*                            ibc testing interface                           */
/* -------------------------------------------------------------------------- */

func (app *App) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetTxConfig implements ibctesting.TestingApp.
func (app *App) GetTxConfig() client.TxConfig {
	return app.txConfig
}

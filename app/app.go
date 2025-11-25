package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	// 1. Cosmos SDK/Consensus Imports
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/gogoproto/proto"
	
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	
	"github.com/spf13/cast"

	// 2. Ethermint/EVM Imports
	"github.com/evmos/ethermint/server/flags"
	ethermint "github.com/evmos/ethermint/types"
	evmclient "github.com/evmos/ethermint/x/evm/client"

	// Force-load the tracer engines to trigger registration due to Go-Ethereum v1.10.15 changes
	_ "github.com/ethereum/go-ethereum/eth/tracers/js"
	_ "github.com/ethereum/go-ethereum/eth/tracers/native"

	// 3. Dymension Local Imports
	"github.com/dymensionxyz/dymension/v3/app/ante"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	v5 "github.com/dymensionxyz/dymension/v3/app/upgrades/v5"
	denommetadatamoduleclient "github.com/dymensionxyz/dymension/v3/x/denommetadata/client"
)

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

	// Set the global power reduction for the app (used by EVM)
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
	mm *module.Manager
	// the module basic manager
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
	encoding := appparams.MakeEncodingConfig()
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

	// Initialize the app's keys (MaccPerms, etc.)
	app.GenerateKeys()

	// Register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, app.keys); err != nil {
		panic(fmt.Errorf("failed to register streaming services: %w", err))
	}

	// Initialize Keepers (Account, Bank, Staking, etc.)
	app.InitKeepers(appCodec, legacyAmino, bApp, logger, ModuleAccountAddrs(), appOpts)
	app.SetupHooks()
	app.InitTransferStack()

	// Skip genesis invariants check if specified in options
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// Initialize the Module Manager
	// NOTE: Any module instantiated here that is later modified must be passed by reference.
	app.mm = module.NewManager(app.SetupModules(appCodec, bApp, skipGenesisInvariants)...)

	// Initialize the Basic Module Manager for codec registration and genesis verification.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.mm,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
					denommetadatamoduleclient.CreateDenomMetadataHandler,
					denommetadatamoduleclient.UpdateDenomMetadataHandler,
					evmclient.UpdateVirtualFrontierBankContractProposalHandler,
				},
			),
		})

	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	// Set module ordering for block execution
	app.mm.SetOrderPreBlockers(PreBlockers...)
	app.mm.SetOrderBeginBlockers(BeginBlockers...)
	app.mm.SetOrderEndBlockers(EndBlockers...)

	// Set module ordering for InitGenesis and ExportGenesis
	app.mm.SetOrderInitGenesis(InitGenesis...)
	app.mm.SetOrderExportGenesis(InitGenesis...)

	// Set a custom migration order
	app.mm.SetOrderMigrations(CustomMigrationOrder(app.mm.ModuleNames())...)

	app.mm.RegisterInvariants(app.CrisisKeeper)

	// Register all module services (gRPC query/message services)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err := app.mm.RegisterServices(app.configurator)
	if err != nil {
		panic(fmt.Errorf("failed to register module services: %w", err))
	}

	// Register upgrade handlers after ModuleManager and configurator are set.
	app.setupUpgradeHandlers()

	// Register AutoCLI and Reflection Services
	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(fmt.Errorf("failed to create reflection service: %w", err))
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// Setup Simulation Manager
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.mm.Modules, overrideModules)
	app.sm.RegisterStoreDecoders()

	// Initialize KV/Transient/Memory stores
	app.MountKVStores(KVStoreKeys)
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	// Initialize BaseApp run phases
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// Set Ante Handler
	maxGasWanted := cast.ToUint64(appOpts.Get(flags.EVMMaxTxGasWanted))
	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper: app.AccountKeeper,
		BankKeeper: app.BankKeeper,
		ExtensionOptionChecker: nil,
		FeegrantKeeper: app.FeeGrantKeeper,
		SignModeHandler: txConfig.SignModeHandler(),
		IBCKeeper: app.IBCKeeper,
		FeeMarketKeeper: app.FeeMarketKeeper,
		EvmKeeper: app.EvmKeeper,
		TxFeesKeeper: app.TxFeesKeeper,
		MaxTxGasWanted: maxGasWanted,
		RollappKeeper: *app.RollappKeeper,
		LightClientKeeper: &app.LightClientKeeper,
		CircuitKeeper: &app.CircuitBreakerKeeper,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create ante handler: %w", err))
	}

	app.SetAnteHandler(anteHandler)

	// Validate Protobuf annotations for message services
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Log warning for annotation issues (does not halt the application)
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

// PreBlocker executes application logic before block execution
func (app *App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx)
}

// BeginBlocker executes application logic at the beginning of the block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker executes application logic at the end of the block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

func (a *App) Configurator() module.Configurator {
	return a.configurator
}

// InitChainer initializes the chain with the genesis state
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

// LegacyAmino returns SimApp's amino codec. NOTE: Used solely for testing.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns an app codec. NOTE: Used solely for testing.
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

// RegisterAPIRoutes registers all application module routes with the provided API server.
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

	// Register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
	HealthcheckRegister(clientCtx, apiSvr.Router)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.interfaceRegistry,
		cmtApp.Query,
	)
}

func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// setupUpgradeHandlers registers the handlers for on-chain upgrades.
func (app *App) setupUpgradeHandlers() {
	for _, u := range Upgrades {
		app.setupUpgradeHandler(u)
	}
}

// setupUpgradeHandler registers a specific upgrade handler.
func (app *App) setupUpgradeHandler(upgrade upgrades.Upgrade) {
	app.UpgradeKeeper.SetUpgradeHandler(
		upgrade.Name,
		upgrade.CreateHandler(
			app.mm,
			app.configurator,
			&upgrades.UpgradeKeepers{
				AccountKeeper:      &app.AccountKeeper,
				CircuitBreakKeeper: &app.CircuitBreakerKeeper,
				LockupKeeper:       app.LockupKeeper,
				IROKeeper:          app.IROKeeper,
				GAMMKeeper:         app.GAMMKeeper,
				GovKeeper:          app.GovKeeper,
				IncentivesKeeper:   app.IncentivesKeeper,
				RollappKeeper:      app.RollappKeeper,
				SponsorshipKeeper:  &app.SponsorshipKeeper,
				ParamsKeeper:       &app.ParamsKeeper,
				DelayedAckKeeper:   &app.DelayedAckKeeper,
				EIBCKeeper:         &app.EIBCKeeper,
				DymNSKeeper:        &app.DymNSKeeper,
				StreamerKeeper:     &app.StreamerKeeper,
				OTCBuybackKeeper:   app.OTCBuybackKeeper,
				SequencerKeeper:    app.SequencerKeeper,
				MintKeeper:         &app.MintKeeper,
				SlashingKeeper:     &app.SlashingKeeper,
				ConsensusKeeper:    &app.ConsensusParamsKeeper,
				RateLimitingKeeper: &app.RateLimitingKeeper,
				TxfeesKeeper:       app.TxFeesKeeper,
			},
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}

	if upgradeInfo.Name == upgrade.Name && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// Configure store loader with the store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
	}
}

// ExportState exports the application's current state (genesis).
func (app *App) ExportState(ctx sdk.Context) map[string]json.RawMessage {
	export, err := app.mm.ExportGenesis(ctx, app.AppCodec())
	if err != nil {
		panic(err)
	}
	return export
}

/* --- IBC Testing Interface --- */

// GetBaseApp implements ibctesting.TestingApp.
func (app *App) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetTxConfig implements ibctesting.TestingApp.
func (app *App) GetTxConfig() client.TxConfig {
	return app.txConfig
}

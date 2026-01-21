package app

// ... (imports remain similar but organized by internal/external)

/**
 * Dymension App Architecture - Security Optimized
 * Focus: Gas safety, State isolation, and Modular registration
 */

// App constants and interfaces
var (
	_ servertypes.Application = (*App)(nil)
	_ runtime.AppI            = (*App)(nil)
	_ ibctesting.TestingApp   = (*App)(nil)
)

// New initializes the Dymension blockchain application.
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	// 1. Initial Encoding & BaseApp Setup
	encoding := appparams.MakeEncodingConfig()
	bApp := baseapp.NewBaseApp(
		appparams.Name, 
		logger, 
		db, 
		encoding.TxConfig.TxDecoder(), 
		baseAppOptions...,
	)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       encoding.Amino,
		appCodec:          encoding.Codec,
		txConfig:          encoding.TxConfig,
		interfaceRegistry: encoding.InterfaceRegistry,
		AppKeepers:        AppKeepers{},
	}

	// 2. Keeper & Store Initialization
	// Logic isolated into helper methods for readability and auditability
	app.InitKeepers(app.appCodec, app.legacyAmino, bApp, logger, ModuleAccountAddrs(), appOpts)
	
	// 3. Module Management
	// Grouping modules into a central manager with strict order of execution
	skipInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))
	app.mm = module.NewManager(app.SetupModules(app.appCodec, bApp, skipInvariants)...)

	// Explicitly set order for critical blockchain phases
	app.setExecutionOrder()

	// 4. AnteHandler: The Security Gatekeeper
	// Optimized with strict gas limits for EVM transactions
	app.setupAnteHandler(appOpts, encoding.TxConfig)

	// 5. Upgrade Handlers
	app.setupUpgradeHandlers()

	// 6. Loading state
	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			logger.Error("Failed to load latest version", "err", err)
			panic(err) // Critical failure: cannot start without state
		}
	}

	return app
}

// setExecutionOrder centralizes the order of pre-blockers, begin-blockers, and end-blockers.
func (app *App) setExecutionOrder() {
	app.mm.SetOrderPreBlockers(PreBlockers...)
	app.mm.SetOrderBeginBlockers(BeginBlockers...)
	app.mm.SetOrderEndBlockers(EndBlockers...)
	app.mm.SetOrderInitGenesis(InitGenesis...)
	app.mm.SetOrderExportGenesis(InitGenesis...)
}

// setupAnteHandler configures the transaction filtering logic.
func (app *App) setupAnteHandler(appOpts servertypes.AppOptions, txConfig client.TxConfig) {
	maxGasWanted := cast.ToUint64(appOpts.Get(flags.EVMMaxTxGasWanted))
	
	options := ante.HandlerOptions{
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		FeegrantKeeper:         app.FeeGrantKeeper,
		SignModeHandler:        txConfig.SignModeHandler(),
		IBCKeeper:              app.IBCKeeper,
		FeeMarketKeeper:        app.FeeMarketKeeper,
		EvmKeeper:              app.EvmKeeper,
		TxFeesKeeper:           app.TxFeesKeeper,
		MaxTxGasWanted:         maxGasWanted,
		RollappKeeper:          *app.RollappKeeper,
		LightClientKeeper:      &app.LightClientKeeper,
		CircuitKeeper:          &app.CircuitBreakerKeeper,
		ExtensionOptionChecker: nil, 
	}

	handler, err := ante.NewAnteHandler(options)
	if err != nil {
		panic(fmt.Errorf("ante handler setup failed: %w", err))
	}
	app.SetAnteHandler(handler)
}

// ... Additional helper methods for state export and API routes

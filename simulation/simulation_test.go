package simulation_test

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time" // Added for better random seed generation if needed

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app"
)

// Ensure all simulator flags are registered before test execution
func init() {
	simcli.GetSimulatorFlags()
}

const (
	SimulationAppChainID = "dymension_100-1"
	DefaultBondDenom     = "adym"
	DefaultPowerReduction = 1000000
)

// setupConfiguration sets the required Cosmos SDK address prefixes, default denominations, and power reduction.
func setupConfiguration() {
	// Configure SDK address prefixes
	sdkConfig := types.GetConfig()
	sdkConfig.SetBech32PrefixForAccount("dym", "dympub")
	sdkConfig.SetBech32PrefixForValidator("dymvaloper", "dymvaloperpub")
	sdkConfig.SetBech32PrefixForConsensusNode("dymvalcons", "dymvalconspub")
	sdkConfig.Seal()

	// Configure default denominations and power reduction
	types.DefaultBondDenom = DefaultBondDenom
	types.DefaultPowerReduction = math.NewIntFromUint64(DefaultPowerReduction)
}

/*
To execute a completely pseudo-random simulation (from the root of the repository):

	go test ./simulation \
		-run=TestFullAppSimulation \
		-Enabled=true \
		-NumBlocks=100 \
		...
*/
func TestFullAppSimulation(t *testing.T) {
	// OPTIMIZATION: Centralize configuration setup
	setupConfiguration() 
	
	config := simcli.NewConfigFromFlags()
	config.ChainID = SimulationAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	// Configure application-specific options for the simulation run
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	// The period at which invariants are checked
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue 
	// Default bond denomination used in the genesis state
	appOptions[cli.FlagDefaultBondDenom] = DefaultBondDenom 

	// Initialize the Dymension application instance
	dymdApp := app.New(logger, db, nil, true, appOptions, baseapp.SetChainID(SimulationAppChainID))
	require.Equal(t, "dymension", dymdApp.Name())

	// prepareGenesis must be defined elsewhere and returns the initial state bytes.
	genesis, err := prepareGenesis(dymdApp.AppCodec()) 
	require.NoError(t, err)

	// Run randomized simulation from the configured seed
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		dymdApp.BaseApp,
		simtestutil.AppStateFn(dymdApp.AppCodec(), dymdApp.SimulationManager(), genesis),
		simulationtypes.RandomAccounts,
		simtestutil.SimulationOperations(dymdApp, dymdApp.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		dymdApp.AppCodec(),
	)

	// Export state and simParams before checking the simulation error
	err = simtestutil.CheckExportSimulation(dymdApp, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	appHash := base64.StdEncoding.EncodeToString(dymdApp.LastCommitID().Hash)
	fmt.Println("App hash:", appHash)
}

/*
TestAppStateDeterminism runs a simulation multiple times with the same seed
to ensure that the application is deterministic (i.e., the resulting app hash is the same each time).
*/
func TestAppStateDeterminism(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	// OPTIMIZATION: Centralize configuration setup
	setupConfiguration() 

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = SimulationAppChainID
	
	// If no seed is provided via flags, use a truly random seed for better test coverage
	if config.Seed == simcli.DefaultSeedValue {
		rand.Seed(time.Now().UnixNano())
		config.Seed = rand.Int63() 
	}

	// Default run settings for determinism test
	const (
		numTimesToRunPerSeed = 5
		numSeeds             = 1 
	)

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	// The period at which invariants are checked
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue 
	// Default bond denomination used in the genesis state
	appOptions[cli.FlagDefaultBondDenom] = DefaultBondDenom 

	for i := 0; i < numSeeds; i++ {
		fmt.Println("config.Seed: ", config.Seed)

		appHashList := make([]string, numTimesToRunPerSeed)

		for j := 0; j < numTimesToRunPerSeed; j++ {
			// Setup simulation environment for a clean run
			db, _, logger, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
			require.NoError(t, err, "simulation setup failed")
			
			// Initialize application
			dymdApp := app.New(logger, db, nil, true, appOptions, baseapp.SetChainID(SimulationAppChainID))
			require.Equal(t, "dymension", dymdApp.Name())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			genesis, err := prepareGenesis(dymdApp.AppCodec())
			require.NoError(t, err)

			// Run simulation with the same seed
			_, _, err = simulation.SimulateFromSeed(
				t,
				os.Stdout,
				dymdApp.BaseApp,
				simtestutil.AppStateFn(dymdApp.AppCodec(), dymdApp.SimulationManager(), genesis),
				simulationtypes.RandomAccounts,
				simtestutil.SimulationOperations(dymdApp, dymdApp.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				dymdApp.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simtestutil.PrintStats(db)
			}

			// Capture the final application hash
			appHash := base64.StdEncoding.EncodeToString(dymdApp.LastCommitID().Hash)
			fmt.Printf("Seed: %d, attempt: %d/%d, app hash: %s\n", config.Seed, j+1, numTimesToRunPerSeed, appHash)
			appHashList[j] = appHash

			// Check determinism: compare current hash with the first hash captured for this seed
			if j != 0 {
				require.Equal(
					t, appHashList[0], appHashList[j],
					"non-determinism in seed %d: %d/%d, attempt: %d/%d. Hashes: %s vs %s\n", 
					config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed, appHashList[0], appHashList[j],
				)
			}
		}
	}
}

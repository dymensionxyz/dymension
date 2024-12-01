package simulation_test

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"testing"

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
	appParams "github.com/dymensionxyz/dymension/v3/app/params"
)

func init() {
	simcli.GetSimulatorFlags()
}

const SimulationAppChainID = "dymension_100-1"

/*
To execute a completely pseudo-random simulation (from the root of the repository):

	go test ./simulation \
		-run=TestFullAppSimulation \
		-Enabled=true \
		-NumBlocks=100 \
		-BlockSize=200 \
		-Commit=true \
		-Seed=99 \
		-Period=1 \
		-PrintAllInvariants=true \
		-v -timeout 24h

To export the simulation params to a file at a given block height:

	go test ./simulation \
		-run=TestFullAppSimulation \
		-Enabled=true \
		-NumBlocks=100 \
		-BlockSize=200 \
		-Commit=true \
		-Seed=99 \
		-Period=1 \
		-PrintAllInvariants=true \
		-ExportParamsPath=/path/to/params.json \
		-ExportParamsHeight=50 \
		-v -timeout 24h

To export the simulation app state (i.e genesis) to a file:

	go test ./simulation \
		-run=TestFullAppSimulation \
		-Enabled=true \
		-NumBlocks=100 \
		-BlockSize=200 \
		-Commit=true \
		-Seed=99 \
		-Period=1 \
		-PrintAllInvariants=true \
		-ExportStatePath=/path/to/genesis.json \
		-v -timeout 24h
*/
func TestFullAppSimulation(t *testing.T) {
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

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue // period at which invariants are checked
	appOptions[cli.FlagDefaultBondDenom] = "adym"
	types.DefaultBondDenom = "adym"
	types.DefaultPowerReduction = math.NewIntFromUint64(1000000) // overwrite evm module's default power reduction

	encoding := app.MakeEncodingConfig()

	appParams.SetAddressPrefixes()

	dymdApp := app.New(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		appOptions,
		baseapp.SetChainID(SimulationAppChainID),
	)
	require.Equal(t, "dymension", dymdApp.Name())

	genesis, err := prepareGenesis(dymdApp.AppCodec())
	require.NoError(t, err)

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		dymdApp.BaseApp,
		simtestutil.AppStateFn(dymdApp.AppCodec(), dymdApp.SimulationManager(), genesis),
		simulationtypes.RandomAccounts,
		simtestutil.SimulationOperations(dymdApp, dymdApp.AppCodec(), config),
		dymdApp.ModuleAccountAddrs(),
		config,
		dymdApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
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
TestAppStateDeterminism runs a simulation to ensure that the application is deterministic.
It generates a random seed and runs the simulation multiple times with the same seed to ensure
that the resulting app hash is the same each time. You may manually specify a seed by using
the -Seed flag. The test may take a few minutes to run.

	go test ./simulation \
		-run=TestAppStateDeterminism \
		-Enabled=true \
		-NumBlocks=50 \
		-BlockSize=300 \
		-Commit=true \
		-Period=0 \
		-v -timeout 24h
*/
func TestAppStateDeterminism(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = SimulationAppChainID

	numSeeds := 1
	numTimesToRunPerSeed := 5

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue // period at which invariants are checked
	appOptions[cli.FlagDefaultBondDenom] = "adym"
	types.DefaultBondDenom = "adym"
	types.DefaultPowerReduction = math.NewIntFromUint64(1000000) // overwrite evm module's default power reduction

	encoding := app.MakeEncodingConfig()

	for i := 0; i < numSeeds; i++ {
		if config.Seed == simcli.DefaultSeedValue {
			// overwrite default seed
			config.Seed = rand.Int63() //nolint:gosec
		}

		fmt.Println("config.Seed: ", config.Seed)

		appHashList := make([]string, numTimesToRunPerSeed)

		for j := 0; j < numTimesToRunPerSeed; j++ {
			db, _, logger, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
			require.NoError(t, err, "simulation setup failed")

			dymdApp := app.New(
				logger,
				db,
				nil,
				true,
				map[int64]bool{},
				app.DefaultNodeHome,
				0,
				encoding,
				appOptions,
				baseapp.SetChainID(SimulationAppChainID),
			)
			require.Equal(t, "dymension", dymdApp.Name())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			genesis, err := prepareGenesis(dymdApp.AppCodec())
			require.NoError(t, err)

			_, _, err = simulation.SimulateFromSeed(
				t,
				os.Stdout,
				dymdApp.BaseApp,
				simtestutil.AppStateFn(dymdApp.AppCodec(), dymdApp.SimulationManager(), genesis),
				simulationtypes.RandomAccounts,
				simtestutil.SimulationOperations(dymdApp, dymdApp.AppCodec(), config),
				dymdApp.ModuleAccountAddrs(),
				config,
				dymdApp.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simtestutil.PrintStats(db)
			}

			appHash := base64.StdEncoding.EncodeToString(dymdApp.LastCommitID().Hash)
			fmt.Printf("Seed: %d, appempt: %d/%d, app hash: %s\n", config.Seed, j+1, numTimesToRunPerSeed, appHash)
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, appHashList[0], appHashList[j],
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}

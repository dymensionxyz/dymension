package simulation_test

import (
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
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/testing/simapp"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app"
	appParams "github.com/dymensionxyz/dymension/v3/app/params"
)

func init() {
	simcli.GetSimulatorFlags()
}

const SimulationAppChainID = "dymension_100-1"

/*
To execute a completely pseudo-random simulation:

	 $ go test . \
		-run=TestFullAppSimulation \
		-Enabled=true \
		-NumBlocks=100 \
		-BlockSize=200 \
		-Commit=true \
		-Seed=99 \
		-Period=5 \
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
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue
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
}

// BenchmarkSimulation run the chain simulation
// Running using starport command:
// `starport chain simulate -v --numBlocks 200 --blockSize 50`
// Running as go benchmark test:
// `go test -benchmem -run=^$ -bench ^BenchmarkSimulation ./app -NumBlocks=200 -BlockSize 50 -Commit=true -Verbose=true -Enabled=true`
func BenchmarkSimulation(b *testing.B) {
	simcli.FlagEnabledValue = true
	simcli.FlagCommitValue = true
	config := simcli.NewConfigFromFlags()
	config.ChainID = "simulation-app"

	db, dir, logger, _, err := simtestutil.SetupSimulation(config, "goleveldb-app-sim", "Simulation", simcli.FlagEnabledValue, simcli.FlagEnabledValue)
	require.NoError(b, err, "simulation setup failed")

	b.Cleanup(func() {
		err := db.Close()
		require.NoError(b, err)
		err = os.RemoveAll(dir)
		require.NoError(b, err)
	})

	encoding := app.MakeEncodingConfig()

	dymdApp := app.New(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		simapp.EmptyAppOptions{},
		baseapp.SetChainID(SimulationAppChainID),
	)

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		dymdApp.BaseApp,
		simtestutil.AppStateFn(dymdApp.AppCodec(), dymdApp.SimulationManager(), app.NewDefaultGenesisState(dymdApp.AppCodec())),
		simulationtypes.RandomAccounts,
		simtestutil.SimulationOperations(dymdApp, dymdApp.AppCodec(), config),
		dymdApp.ModuleAccountAddrs(),
		config,
		dymdApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(dymdApp, config, simParams)
	require.NoError(b, err)
	require.NoError(b, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

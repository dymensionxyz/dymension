package simulation_test

import (
	"os"
	"testing"

	simapp "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/stretchr/testify/require"
)

func init() {
	simcli.GetSimulatorFlags()
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

	db, dir, logger, _, err := simapp.SetupSimulation(config, "goleveldb-app-sim", "Simulation", simcli.FlagEnabledValue, simcli.FlagEnabledValue)
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
	)

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		dymdApp.BaseApp,
		simapp.AppStateFn(dymdApp.AppCodec(), dymdApp.SimulationManager(), app.NewDefaultGenesisState(dymdApp.AppCodec())),
		simulationtypes.RandomAccounts,
		simapp.SimulationOperations(dymdApp, dymdApp.AppCodec(), config),
		dymdApp.ModuleAccountAddrs(),
		config,
		dymdApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(dymdApp, config, simParams)
	require.NoError(b, err)
	require.NoError(b, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}
}

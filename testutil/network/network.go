package network

import (
	"context"
	"testing"
	"time"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/params"
)

// New creates a test network instance with a fully configured Cosmos network.
// It accepts optional config, which will override the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	t.Helper()

	if len(configs) > 1 {
		panic("at most one config should be provided")
	}

	cfg := DefaultConfig()
	if len(configs) > 0 {
		cfg = configs[0]
	}

	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err, "failed to create test network")

	// Ensure cleanup is called regardless of test outcome
	t.Cleanup(func() {
		testCleanup(t, net)
	})

	return net
}

func testCleanup(t *testing.T, net *network.Network) {
	t.Log("cleaning up test network resources")
	net.Cleanup()
}

// DefaultConfig initializes the config for the Dymension test network with the custom
// application setup and single validator configuration.
func DefaultConfig() network.Config {
	cfg := network.DefaultConfig(nil)
	encoding := params.MakeEncodingConfig()

	// Use a unique ChainID for each test run to prevent potential conflicts.
	// Using a timestamp ensures near-uniqueness.
	cfg.ChainID = "dymension_" + time.Now().Format("20060102150405") + "-1"

	// Configure the application constructor to use the custom Dymension App
	cfg.AppConstructor = func(val network.ValidatorI) servertypes.Application {
		return app.New(
			val.GetCtx().Logger, dbm.NewMemDB(), nil, true,
			sims.EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			// Use the context for baseapp options to pass necessary config during test initialization
			baseapp.SetContext(context.Background()),
		)
	}

	// Set the correct default genesis state for the Dymension application
	cfg.GenesisState = app.DefaultGenesis(encoding.Codec)

	// Custom Genesis Configuration: Force disable EVM EnableCreate for testing
	if evmGenesisStateJSON, found := cfg.GenesisState[evmtypes.ModuleName]; found {
		cfg.GenesisState[evmtypes.ModuleName] = setupEvmGenesis(encoding, evmGenesisStateJSON)
	}

	cfg.NumValidators = 1

	return cfg
}

// setupEvmGenesis modifies the EVM module genesis state to disable 'EnableCreate'.
func setupEvmGenesis(encoding params.EncodingConfig, evmGenesisStateJSON []byte) []byte {
	var evmGenesisState evmtypes.GenesisState
	
	// Unmarshal the existing EVM genesis state
	encoding.Codec.MustUnmarshalJSON(evmGenesisStateJSON, &evmGenesisState)
	
	// Force disable EVM Create functionality
	evmGenesisState.Params.EnableCreate = false
	
	// Marshal the modified state back into JSON
	return encoding.Codec.MustMarshalJSON(&evmGenesisState)
}

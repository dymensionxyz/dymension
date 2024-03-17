package network

import (
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/stretchr/testify/require"
	tmdb "github.com/tendermint/tm-db"

	"github.com/dymensionxyz/dymension/v3/app"
)

type (
	Network = network.Network
	Config  = network.Config
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()
	} else {
		cfg = configs[0]
	}
	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)

	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func DefaultConfig() network.Config {
	cfg := network.DefaultConfig()
	encoding := app.MakeEncodingConfig()

	// FIXME: add rand tmrand.Uint64() to chainID
	cfg.ChainID = "dymension_1000-1"
	cfg.AppConstructor = func(val network.Validator) servertypes.Application {
		return app.New(
			val.Ctx.Logger, tmdb.NewMemDB(), nil, true, map[int64]bool{}, val.Ctx.Config.RootDir, 0,
			encoding,
			simapp.EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}

	cfg.GenesisState = app.ModuleBasics.DefaultGenesis(encoding.Codec)
	if evmGenesisStateJson, found := cfg.GenesisState[evmtypes.ModuleName]; found {
		// force disable Enable Create of x/evm
		var evmGenesisState evmtypes.GenesisState
		encoding.Codec.MustUnmarshalJSON(evmGenesisStateJson, &evmGenesisState)
		evmGenesisState.Params.EnableCreate = false
		cfg.GenesisState[evmtypes.ModuleName] = encoding.Codec.MustMarshalJSON(&evmGenesisState)
	}

	cfg.NumValidators = 1

	return cfg
}

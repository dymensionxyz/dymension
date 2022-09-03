package app

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
)

type SimApp interface {
	cosmoscmd.App
	GetBaseApp() *baseapp.BaseApp
	AppCodec() codec.Codec
	SimulationManager() *module.SimulationManager
	ModuleAccountAddrs() map[string]bool
	Name() string
	LegacyAmino() *codec.LegacyAmino
	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock
	EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock
	InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain
}

var defaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

func setup(withGenesis bool, invCheckPeriod uint) (*App, GenesisState) {
	simapp.FlagEnabledValue = true
	simapp.FlagCommitValue = true

	_, db, _, logger, _, _ := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
	// require.NoError(b, err, "simulation setup failed")

	// b.Cleanup(func() {
	// 	db.Close()
	// 	err = os.RemoveAll(dir)
	// 	require.NoError(b, err)
	// })

	encoding := cosmoscmd.MakeEncodingConfig(ModuleBasics)

	myApp := NewSim(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		encoding,
		simapp.EmptyAppOptions{},
	)

	newApp, ok := myApp.(*App)
	_ = ok
	//require.True(b, ok, "can't use simapp")

	if withGenesis {
		return newApp, NewDefaultGenesisState(newApp.AppCodec())
	}
	return newApp, GenesisState{}
}

func AppCodec() {
	panic("unimplemented")
}

// Setup initializes a new SimApp. A Nop logger is set in SimApp.
func Setup(isCheckTx bool) *App {
	dymSimApp, genesisState := setup(!isCheckTx, 5)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		(*dymSimApp).InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: defaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return dymSimApp
}

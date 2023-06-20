package app

// import (
// 	"encoding/json"
// 	"time"

// 	"github.com/cosmos/cosmos-sdk/simapp"

// 	abci "github.com/tendermint/tendermint/abci/types"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
// 	tmtypes "github.com/tendermint/tendermint/types"

// 	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
// 	"github.com/tendermint/tendermint/libs/log"
// 	dbm "github.com/tendermint/tm-db"
// )

// var defaultConsensusParams = &abci.ConsensusParams{
// 	Block: &abci.BlockParams{
// 		MaxBytes: 200000,
// 		MaxGas:   2000000,
// 	},
// 	Evidence: &tmproto.EvidenceParams{
// 		MaxAgeNumBlocks: 302400,
// 		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
// 		MaxBytes:        10000,
// 	},
// 	Validator: &tmproto.ValidatorParams{
// 		PubKeyTypes: []string{
// 			tmtypes.ABCIPubKeyTypeEd25519,
// 		},
// 	},
// }

// func SetupTestApp(withGenesis bool) (*App, GenesisState, simtypes.Config, dbm.DB, string, log.Logger) {
// 	simapp.GetSimulatorFlags()
// 	simapp.FlagEnabledValue = true
// 	simapp.FlagCommitValue = true

// 	config, db, dir, logger, _, err := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
// 	if err != nil {
// 		panic(err)
// 	}

// 	encoding := MakeEncodingConfig()

// 	app := New(
// 		logger,
// 		db,
// 		nil,
// 		true,
// 		map[int64]bool{},
// 		DefaultNodeHome,
// 		0,
// 		encoding,
// 		simapp.EmptyAppOptions{},
// 	)

// 	genesisState := GenesisState{}
// 	if withGenesis {
// 		genesisState = NewDefaultGenesisState(app.AppCodec())
// 	}

// 	return app, genesisState, config, db, dir, logger
// }

// // Setup initializes a new test App. A Nop logger is set in App.
// func Setup(isCheckTx bool) *App {
// 	simApp, genesisState, _, _, _, _ := SetupTestApp(!isCheckTx)
// 	if !isCheckTx {
// 		// init chain must be called to stop deliverState from being nil
// 		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
// 		if err != nil {
// 			panic(err)
// 		}

// 		// Initialize the chain
// 		(*simApp).InitChain(
// 			abci.RequestInitChain{
// 				ChainId:         "dymension_100-1",
// 				Validators:      []abci.ValidatorUpdate{},
// 				ConsensusParams: defaultConsensusParams,
// 				AppStateBytes:   stateBytes,
// 			},
// 		)
// 	}

// 	return simApp
// }

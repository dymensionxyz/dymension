package app

import (
	"bytes"

	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"

	app "github.com/dymensionxyz/dymension/app"

	abci "github.com/tendermint/tendermint/abci/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// set up a new baseapp from given params
func setupBaseAppFromParams(app *BaseApp, db dbm.DB, storeKeyToIAVLTree map[string]*iavltree.DeepSubTree, blockHeight int64, storeKeys []storetypes.StoreKey, options ...func(*BaseApp)) (*BaseApp, error) {
	// This initial height is used in `BeginBlock` in `validateHeight`
	options = append(options, SetInitialHeight(blockHeight))

	appName := app.Name() + "FromFraudProof"

	//FIXME: need to run the rollapp app here
	newApp := baseapp.NewBaseApp(appName, app.logger, db, app.txDecoder, options...)

	// newApp.msgServiceRouter = app.msgServiceRouter
	// newApp.SetBeginBlocker(rollapp.BeginBlocker())
	// newApp.beginBlocker = app.beginBlocker
	// newApp.endBlocker = app.endBlocker
	// stores are mounted

	newApp.CommitMultiStore()

	newApp.MountStores(storeKeys...)

	//FIXME: should use deep iavl tree
	cmsStore := newApp.CommitMultiStore().(*rootmulti.Store)
	for storeKey, iavlTree := range storeKeyToIAVLTree {
		cmsStore.SetDeepIAVLTree(storeKey, iavlTree)
	}
	err := newApp.LoadLatestVersion()
	return newApp, err
}

// set up a new baseapp from a fraudproof
func SetupBaseAppFromFraudProof(db dbm.DB, fraudProof *types.FraudProof, storeKeys []storetypes.StoreKey, options ...func(*BaseApp)) (*app.BaseApp, error) {
	storeKeyToIAVLTree, err := fraudProof.GetDeepIAVLTrees()
	if err != nil {
		return nil, err
	}
	return setupBaseAppFromParams(db, storeKeyToIAVLTree, fraudProof.BlockHeight, storeKeys, options...)
}

// VerifyFraudProof implements the ABCI application interface. It loads a fresh BaseApp using
// the given Fraud Proof, runs the given fraudulent state transition within the Fraud Proof,
// and gets the app hash representing state of the resulting BaseApp. It returns a boolean
// representing whether this app hash is equivalent to the expected app hash given.
func VerifyFraudProof(fraudProof *types.FraudProof) error {
	// Store and subtore level verification
	success, err := fraudProof.VerifyFraudProof()
	if err != nil {
		return err
	}

	if success {
		// State execution verification
		modules := fraudProof.GetModules()
		iavlStoreKeys := make([]storetypes.StoreKey, 0, len(modules))

		//FIXME: get module names from L2 app
		// for _, module := range modules {
		// 	iavlStoreKeys = append(iavlStoreKeys, storeKeys[module])
		// }

		// Setup a new app from fraud proof
		appFromFraudProof, err := SetupBaseAppFromFraudProof(
			// app,
			db.NewMemDB(),
			fraudProof,
			iavlStoreKeys,
			// options...,
		)
		if err != nil {
			panic(err)
		}
		appFromFraudProof.InitChain(abci.RequestInitChain{})
		appHash := appFromFraudProof.GetAppHash(abci.RequestGetAppHash{}).AppHash

		if !bytes.Equal(fraudProof.PreStateAppHash, appHash) {
			//TODO: rename error
			return types.ErrInvalidAppHash
		}

		//FIXME: let's try to run all

		/*
			// Execute fraudulent state transition
			if fraudProof.fraudulentBeginBlock != nil {
				appFromFraudProof.BeginBlock(*fraudProof.fraudulentBeginBlock)
			} else {
				// Need to add some dummy begin block here since its a new app
				appFromFraudProof.beginBlocker = nil
				appFromFraudProof.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: fraudProof.blockHeight}})
				if fraudProof.fraudulentDeliverTx != nil {
					resp := appFromFraudProof.DeliverTx(*fraudProof.fraudulentDeliverTx)
					if !resp.IsOK() {
						panic(resp.Log)
					}
				} else {
					appFromFraudProof.EndBlock(*fraudProof.fraudulentEndBlock)
				}
			}
		*/

		appHash = appFromFraudProof.GetAppHash(abci.RequestGetAppHash{}).AppHash
		success = bytes.Equal(appHash, req.ExpectedValidAppHash)
	}
	res = abci.ResponseVerifyFraudProof{
		Success: success,
	}
	return res
}

package fraudproof

import (
	"bytes"
	"context"
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"

	abci "github.com/tendermint/tendermint/abci/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/app/fraudproof/types"
)

type FraudProofVerifier struct {
	ctx context.Context
	app *baseapp.BaseApp
}

// New creates a new FraudProofVerifier
func New(ctx context.Context, appName string, logger log.Logger, txDecoder sdk.TxDecoder) *FraudProofVerifier {
	newApp := baseapp.NewBaseApp(appName, logger, db.NewMemDB(), txDecoder)
	return &FraudProofVerifier{
		ctx: ctx,
		app: newApp,
	}
}

// InitFromFraudProof initializes the FraudProofVerifier from a fraud proof
func (fpv *FraudProofVerifier) InitFromFraudProof(fraudProof *types.FraudProof) error {
	// check app is initialized
	if fpv.app == nil {
		// return types.ErrAppNotInitialized
		return fmt.Errorf("app not initialized")
	}

	_, err := fraudProof.VerifyFraudProof()
	if err != nil {
		return err
	}

	storeKeyToIAVLTree, err := fraudProof.GetDeepIAVLTrees()
	if err != nil {
		return err
	}

	modules := fraudProof.GetModules()
	iavlStoreKeys := make([]storetypes.StoreKey, 0, len(modules))
	//FIXME: get module names from L2 app
	// modules := fraudProof.getModules()
	// iavlStoreKeys := make([]types.StoreKey, 0, len(modules))
	// for _, module := range modules {
	// 	iavlStoreKeys = append(iavlStoreKeys, app.cms.(*rootmulti.Store).StoreKeysByName()[module])
	// }

	fpv.app.CommitMultiStore()

	fpv.app.MountStores(iavlStoreKeys...)

	cmsStore := fpv.app.CommitMultiStore().(*rootmulti.Store)
	for storeKey, iavlTree := range storeKeyToIAVLTree {
		cmsStore.SetDeepIAVLTree(storeKey, iavlTree)
	}
	_ = cmsStore
	err = fpv.app.LoadLatestVersion()
	if err != nil {
		return err
	}
	return nil

}

// VerifyFraudProof implements the ABCI application interface. It loads a fresh BaseApp using
// the given Fraud Proof, runs the given fraudulent state transition within the Fraud Proof,
// and gets the app hash representing state of the resulting BaseApp. It returns a boolean
// representing whether this app hash is equivalent to the expected app hash given.
func (fpv *FraudProofVerifier) VerifyFraudProof(fraudProof *types.FraudProof) error {

	//TODO: check app is initialized

	//TODO: pass rollapp name as well
	fpv.app.InitChain(abci.RequestInitChain{InitialHeight: fraudProof.BlockHeight})
	appHash := fpv.app.GetAppHash(abci.RequestGetAppHash{}).AppHash

	if !bytes.Equal(fraudProof.PreStateAppHash, appHash) {
		return types.ErrInvalidPreStateAppHash
	}

	//FIXME: let's try to run all

	/*
		// Execute fraudulent state transition
		if fraudProof.fraudulentBeginBlock != nil {
			fpv.app.BeginBlock(*fraudProof.fraudulentBeginBlock)
		} else {
			// Need to add some dummy begin block here since its a new app
			fpv.app.beginBlocker = nil
			fpv.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: fraudProof.blockHeight}})
			if fraudProof.fraudulentDeliverTx != nil {
				resp := fpv.app.DeliverTx(*fraudProof.fraudulentDeliverTx)
				if !resp.IsOK() {
					panic(resp.Log)
				}
			} else {
				fpv.app.EndBlock(*fraudProof.fraudulentEndBlock)
			}
		}
	*/

	appHash = fpv.app.GetAppHash(abci.RequestGetAppHash{}).AppHash
	if !bytes.Equal(appHash, fraudProof.ExpectedValidAppHash) {
		return types.ErrInvalidAppHash
	}
	return nil
}

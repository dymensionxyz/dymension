package fraudproof

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"

	abci "github.com/tendermint/tendermint/abci/types"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	rollappevm "github.com/dymensionxyz/rollapp-evm/app"

	appparams "github.com/dymensionxyz/dymension/app/params"

	_ "github.com/evmos/evmos/v12/crypto/codec"
	_ "github.com/evmos/evmos/v12/crypto/ethsecp256k1"
	_ "github.com/evmos/evmos/v12/types"
)

var (
	ErrInvalidPreStateAppHash = errors.New("invalid pre state app hash")
	ErrInvalidAppHash         = errors.New("invalid app hash")
)

type FraudProofVerifier interface {
	InitFromFraudProof(fraudProof *fraudtypes.FraudProof) error
	VerifyFraudProof(fraudProof *fraudtypes.FraudProof) error
}

type RollappFPV struct {
	host *baseapp.BaseApp
	app  *baseapp.BaseApp
}

var _ FraudProofVerifier = (*RollappFPV)(nil)

// New creates a new FraudProofVerifier
func New(host *baseapp.BaseApp, appName string, logger log.Logger, _ appparams.EncodingConfig) *RollappFPV {
	//TODO: use logger?
	//TODO: default home directory?

	//TODO: test with dymension app for working reference
	// encCdc := rollappparams.MakeEncodingConfig()
	// encCdc := rollappparams.MakeEncodingConfig()
	// rollapp := rollappevm.NewRollapp(log.NewNopLogger(), db.NewMemDB(), nil, true, map[int64]bool{}, "/tmp", 0, encCdc, nil)

	// rollapp := app.New(log.NewNopLogger(), db.NewMemDB(), nil, true, map[int64]bool{}, "/tmp", 0, encCdc, nil)

	//TODO: need to create here a rollapp app, and use it's begin block and txdecoder

	// cfg := appparams.MakeEncodingConfigFP()

	// encodingConfig := simappparams.MakeTestEncodingConfig()

	cfg := rollappevm.MakeEncodingConfig()

	rollapp := rollappevm.NewBaseAppRollapp(log.NewNopLogger(), dbm.NewMemDB(), nil, false, map[int64]bool{}, "/tmp", 0, cfg, simapp.EmptyAppOptions{})
	// _ = rollappevm.GetMaccPerms()
	// baseapp := rollapp.GetBaseApp()

	// rollapp := baseapp.NewBaseApp(appName, log.NewNopLogger(), dbm.NewMemDB(), encodingCfg.TxConfig.TxDecoder())
	//FIXME: remove this
	// if host != nil {
	// 	newApp.SetMsgServiceRouter(host.MsgServiceRouter())
	// newApp.SetBeginBlocker(host.GetBeginBlocker())
	// 	newApp.SetEndBlocker(host.GetEndBlocker())
	// }

	return &RollappFPV{
		host: host,
		app:  rollapp,
	}
}

// InitFromFraudProof initializes the FraudProofVerifier from a fraud proof
func (fpv *RollappFPV) InitFromFraudProof(fraudProof *fraudtypes.FraudProof) error {
	// check app is initialized
	if fpv.app == nil {
		// return types.ErrAppNotInitialized
		return fmt.Errorf("app not initialized")
	}

	_, err := fraudProof.ValidateBasic()
	if err != nil {
		return err
	}

	fpv.app.SetInitialHeight(fraudProof.BlockHeight)

	cmsHost := fpv.host.CommitMultiStore().(*rootmulti.Store)
	storeKeys := cmsHost.StoreKeysByName()
	// modules := fraudProof.GetModules()
	// iavlStoreKeys := make([]storetypes.StoreKey, 0, len(modules))
	// for _, module := range modules {
	// iavlStoreKeys = append(iavlStoreKeys, storeKeys[module])
	// }

	iavlStoreKeys := make([]storetypes.StoreKey, 0, len(storeKeys))
	for _, storeKey := range storeKeys {
		iavlStoreKeys = append(iavlStoreKeys, storeKey)
	}
	fpv.app.MountStores(iavlStoreKeys...)

	storeKeyToIAVLTree, err := fraudProof.GetDeepIAVLTrees()
	if err != nil {
		return err
	}
	cmsStore := fpv.app.CommitMultiStore().(*rootmulti.Store)
	for storeKey, iavlTree := range storeKeyToIAVLTree {
		cmsStore.SetDeepIAVLTree(storeKey, iavlTree)
	}

	err = fpv.app.LoadLatestVersion()
	if err != nil {
		return err
	}

	//is it enough? rollkit uses:
	//	// This initial height is used in `BeginBlock` in `validateHeight`
	// options = append(options, SetInitialHeight(blockHeight))
	fpv.app.InitChain(abci.RequestInitChain{})

	// fpv.app.ResetDeliverState()

	return nil

}

// VerifyFraudProof implements the ABCI application interface. It loads a fresh BaseApp using
// the given Fraud Proof, runs the given fraudulent state transition within the Fraud Proof,
// and gets the app hash representing state of the resulting BaseApp. It returns a boolean
// representing whether this app hash is equivalent to the expected app hash given.
func (fpv *RollappFPV) VerifyFraudProof(fraudProof *fraudtypes.FraudProof) error {

	//TODO: check app is initialized

	//TODO: pass rollapp name as well
	appHash := fpv.app.GetAppHashInternal()
	fmt.Println("appHash - prestate", hex.EncodeToString(appHash))

	if !bytes.Equal(fraudProof.PreStateAppHash, appHash) {
		return ErrInvalidPreStateAppHash
	}

	//TODO: verifyy all data exists in fraud proof

	// Execute fraudulent state transition
	if fraudProof.FraudulentBeginBlock != nil {
		fpv.app.BeginBlock(*fraudProof.FraudulentBeginBlock)
		fmt.Println("appHash - beginblock", hex.EncodeToString(fpv.app.GetAppHashInternal()))
	} else {
		// Need to add some dummy begin block here since its a new app
		fpv.app.ResetDeliverState()
		fpv.app.SetBeginBlocker(nil)
		fpv.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: fraudProof.BlockHeight}})
		fmt.Println("appHash - dummy beginblock", hex.EncodeToString(fpv.app.GetAppHashInternal()))

		// skip IncrementSequenceDecorator check in AnteHandler
		fpv.app.SetAnteHandler(nil)

		if fraudProof.FraudulentDeliverTx != nil {
			resp := fpv.app.DeliverTx(*fraudProof.FraudulentDeliverTx)
			if !resp.IsOK() {
				panic(resp.Log)
			}
			fmt.Println("appHash - posttx", hex.EncodeToString(fpv.app.GetAppHashInternal()))

		} else {
			fpv.app.EndBlock(*fraudProof.FraudulentEndBlock)
			fmt.Println("appHash - endblock", hex.EncodeToString(fpv.app.GetAppHashInternal()))
		}
	}

	appHash = fpv.app.GetAppHashInternal()
	fmt.Println("appHash - final", hex.EncodeToString(appHash))
	if !bytes.Equal(appHash, fraudProof.ExpectedValidAppHash) {
		return types.ErrInvalidAppHash
	}
	return nil
}

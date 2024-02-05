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
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	rollappevm "github.com/dymensionxyz/rollapp-evm/app"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"

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
	keys map[string]storetypes.StoreKey
}

var _ FraudProofVerifier = (*RollappFPV)(nil)

// New creates a new FraudProofVerifier
func New(host *baseapp.BaseApp, appName string, logger log.Logger, _ appparams.EncodingConfig) *RollappFPV {
	//TODO: use logger?
	//TODO: default home directory?

	cfg := rollappevm.MakeEncodingConfig()

	//FIXNE: the export key hack doesnt work. need to get it from the app after a mount
	// rollapp, exportKeys := rollappevm.NewBaseAppRollapp(log.NewNopLogger(), dbm.NewMemDB(), nil, false, map[int64]bool{}, "/tmp", 0, cfg, simapp.EmptyAppOptions{})
	rollappApp := rollappevm.NewRollapp(log.NewNopLogger(), dbm.NewMemDB(), nil, false, map[int64]bool{}, "/tmp", 0, cfg, simapp.EmptyAppOptions{})

	rollapp := baseapp.NewBaseApp(appName, log.NewNopLogger(), dbm.NewMemDB(), cfg.TxConfig.TxDecoder())
	//FIXME: remove this
	// if host != nil {
	rollapp.SetMsgServiceRouter(rollappApp.MsgServiceRouter())
	rollapp.SetBeginBlocker(rollappApp.GetBeginBlocker())
	rollapp.SetEndBlocker(rollappApp.GetEndBlocker())
	// }

	cms := rollappApp.CommitMultiStore().(*rootmulti.Store)
	storeKeys := cms.StoreKeysByName()

	return &RollappFPV{
		host: host,
		app:  rollapp,
		keys: storeKeys,
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

	fpv.app.SetInitialHeight(fraudProof.BlockHeight + 1) //FIXME: why +1?

	cms := fpv.app.CommitMultiStore().(*rootmulti.Store)
	storeKeys := fpv.keys
	modules := fraudProof.GetModules()
	iavlStoreKeys := make([]storetypes.StoreKey, 0, len(modules))
	for _, module := range modules {
		iavlStoreKeys = append(iavlStoreKeys, storeKeys[module])
	}

	fpv.app.MountStores(iavlStoreKeys...)

	storeKeyToIAVLTree, err := fraudProof.GetDeepIAVLTrees()
	if err != nil {
		return err
	}
	for storeKey, iavlTree := range storeKeyToIAVLTree {
		cms.SetDeepIAVLTree(storeKey, iavlTree)
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
		panic("fraudulent begin block not supported")
		// fpv.app.BeginBlock(*fraudProof.FraudulentBeginBlock)
		// fmt.Println("appHash - beginblock", hex.EncodeToString(fpv.app.GetAppHashInternal()))
	} else {
		// Need to add some dummy begin block here since its a new app
		fpv.app.ResetDeliverState()
		fpv.app.SetBeginBlocker(nil)
		fpv.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: fraudProof.BlockHeight + 1}}) //FIXME: why +1?
		fmt.Println("appHash - dummy beginblock", hex.EncodeToString(fpv.app.GetAppHashInternal()))

		if fraudProof.FraudulentDeliverTx != nil {
			//WE NEED TO OVERWRITE THE SDK CONFIG HERE, as txs will fail due to the singleton config
			// skip IncrementSequenceDecorator check in AnteHandler
			fpv.app.SetAnteHandler(nil)
			SetRollappAddressPrefixes("ethm")

			resp := fpv.app.DeliverTx(*fraudProof.FraudulentDeliverTx)
			if !resp.IsOK() {
				panic(resp.Log)
			}
			fmt.Println("appHash - posttx", hex.EncodeToString(fpv.app.GetAppHashInternal()))
			SetRollappAddressPrefixes("dym")
		} else {
			panic("fraudulent end block not supported")
			// fpv.app.EndBlock(*fraudProof.FraudulentEndBlock)
			// fmt.Println("appHash - endblock", hex.EncodeToString(fpv.app.GetAppHashInternal()))
		}
	}

	appHash = fpv.app.GetAppHashInternal()
	fmt.Println("appHash - final", hex.EncodeToString(appHash))
	if !bytes.Equal(appHash, fraudProof.ExpectedValidAppHash) {
		return types.ErrInvalidAppHash
	}
	return nil
}

func SetRollappAddressPrefixes(prefix string) {
	// Set prefixes
	accountPubKeyPrefix := prefix + "pub"
	validatorAddressPrefix := prefix + "valoper"
	validatorPubKeyPrefix := prefix + "valoperpub"
	consNodeAddressPrefix := prefix + "valcons"
	consNodePubKeyPrefix := prefix + "valconspub"

	// Set config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccountNoAssert(prefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidatorNoAssert(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNodeNoAssert(consNodeAddressPrefix, consNodePubKeyPrefix)
}

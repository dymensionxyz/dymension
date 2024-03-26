package fraudproof

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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
	rollappevmparams "github.com/dymensionxyz/rollapp-evm/app/params"

	_ "github.com/evmos/evmos/v12/crypto/codec"
	_ "github.com/evmos/evmos/v12/crypto/ethsecp256k1"
	_ "github.com/evmos/evmos/v12/types"
)

// TODO: Move to types package
var (
	ErrInvalidPreStateAppHash = errors.New("invalid pre state app hash")
	ErrInvalidAppHash         = errors.New("invalid app hash") // TODO(danwt): use or delete
)

type Verifier struct {
	name           string
	storeKeys      map[string]storetypes.StoreKey
	encCfg         rollappevmparams.EncodingConfig
	rollappBaseApp *baseapp.BaseApp
	runningApp     *baseapp.BaseApp
}

// NewVerifier creates a new Verifier
func NewVerifier(appName string) *Verifier {
	cfg := rollappevm.MakeEncodingConfig()

	// TODO: use logger? default home directory?
	rollappApp := rollappevm.NewRollapp(log.NewNopLogger(), dbm.NewMemDB(), nil, false, map[int64]bool{}, "/tmp", 0, cfg, simapp.EmptyAppOptions{})
	storeKeys := rollappApp.CommitMultiStore().(*rootmulti.Store).StoreKeysByName()

	return &Verifier{
		name:           appName,
		encCfg:         cfg,
		storeKeys:      storeKeys,
		rollappBaseApp: rollappApp.GetBaseApp(),
	}
}

func (fpv *Verifier) initCleanInstance() {
	rollapp := baseapp.NewBaseApp(fpv.name, log.NewNopLogger(), dbm.NewMemDB(), fpv.encCfg.TxConfig.TxDecoder())
	rollapp.SetMsgServiceRouter(fpv.rollappBaseApp.MsgServiceRouter())
	rollapp.SetBeginBlocker(fpv.rollappBaseApp.GetBeginBlocker())
	rollapp.SetEndBlocker(fpv.rollappBaseApp.GetEndBlocker())
	fpv.runningApp = rollapp
}

// Init initializes the Verifier from a fraud proof
func (fpv *Verifier) Init(fraudProof *fraudtypes.FraudProof) error {
	// check app is initialized
	if fpv.rollappBaseApp == nil {
		return fmt.Errorf("app not initialized")
	}

	fpv.initCleanInstance()

	fpv.runningApp.SetInitialHeight(fraudProof.GetFraudulentBlockHeight())

	cms := fpv.runningApp.CommitMultiStore().(*rootmulti.Store)
	storeKeys := fpv.storeKeys
	modules := fraudProof.GetModules()
	iavlStoreKeys := make([]storetypes.StoreKey, 0, len(modules))
	for _, module := range modules {
		iavlStoreKeys = append(iavlStoreKeys, storeKeys[module])
	}

	// FIXME: make sure non is nil

	fpv.runningApp.MountStores(iavlStoreKeys...)

	storeKeyToIAVLTree, err := fraudProof.GetDeepIAVLTrees()
	if err != nil {
		return err
	}
	for storeKey, iavlTree := range storeKeyToIAVLTree {
		cms.SetDeepIAVLTree(storeKey, iavlTree)
	}

	err = fpv.runningApp.LoadLatestVersion()
	if err != nil {
		return err
	}

	fpv.runningApp.InitChain(abci.RequestInitChain{})

	return nil
}

// Verify checks the validity of a given fraud proof.
//
// The function takes a FraudProof object as an argument and returns an error if the fraud proof is invalid.
//
// The function performs the following checks:
// 1. It checks if the pre-state application hash of the fraud proof matches the current application hash.
// 2. It executes a fraudulent state transition.
// 3. Finally, it checks if the post-state application hash matches the expected valid application hash in the fraud proof.
//
// If any of these checks fail, the function returns an error. Otherwise, it returns nil.
//
// Note: This function modifies the state of the Verifier object it's called on.
func (fpv *Verifier) Verify(fraudProof *fraudtypes.FraudProof) error {
	appHash := fpv.runningApp.GetAppHashInternal()
	fmt.Println("appHash - prestate", hex.EncodeToString(appHash))

	if !bytes.Equal(fraudProof.PreStateAppHash, appHash) {
		return ErrInvalidPreStateAppHash
	}

	SetRollappAddressPrefixes("ethm")

	// Execute fraudulent state transition
	if fraudProof.FraudulentBeginBlock != nil {
		panic("fraudulent begin block not supported")
		// fpv.app.BeginBlock(*fraudProof.FraudulentBeginBlock)
		// fmt.Println("appHash - beginblock", hex.EncodeToString(fpv.app.GetAppHashInternal()))
	} else {
		// Need to add some dummy begin block here since its a new app
		fpv.runningApp.ResetDeliverState()
		fpv.runningApp.SetBeginBlocker(nil)
		fpv.runningApp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: fraudProof.BlockHeight + 1}})
		fmt.Println("appHash - dummy beginblock", hex.EncodeToString(fpv.runningApp.GetAppHashInternal()))

		if fraudProof.FraudulentDeliverTx != nil {
			// skip IncrementSequenceDecorator check in AnteHandler
			fpv.runningApp.SetAnteHandler(nil)

			resp := fpv.runningApp.DeliverTx(*fraudProof.FraudulentDeliverTx)
			if !resp.IsOK() {
				panic(resp.Log)
			}
			fmt.Println("appHash - posttx", hex.EncodeToString(fpv.runningApp.GetAppHashInternal()))
			SetRollappAddressPrefixes("dym")
		} else {
			panic("fraudulent end block not supported")
			// fpv.app.EndBlock(*fraudProof.FraudulentEndBlock)
			// fmt.Println("appHash - endblock", hex.EncodeToString(fpv.app.GetAppHashInternal()))
		}
	}

	appHash = fpv.runningApp.GetAppHashInternal()
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

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
	appName string
	// the encoding config used to initialize the verification process with a fresh app TODO: could just use decoder
	encCfg rollappevmparams.EncodingConfig
	// the base app that is used to initialize the verification process on each verification attempt
	baseApp *baseapp.BaseApp
	// the mutable base app that is used to perform the verification process
	mutableBaseApp *baseapp.BaseApp
}

// NewVerifier creates a new Verifier
func NewVerifier(appName string) *Verifier {
	cfg := rollappevm.MakeEncodingConfig()

	// TODO: use logger? default home directory?
	rollappApp := rollappevm.NewRollapp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		false,
		map[int64]bool{},
		"/tmp",
		0,
		cfg, simapp.EmptyAppOptions{},
	)

	return &Verifier{
		appName: appName,
		encCfg:  cfg,
		baseApp: rollappApp.GetBaseApp(),
	}
}

func (v *Verifier) moduleStoreKey(module string) storetypes.StoreKey {
	return v.baseApp.CommitMultiStore().(*rootmulti.Store).StoreKeysByName()[module]
}

func (v *Verifier) initCleanInstance() {
	rollapp := baseapp.NewBaseApp(v.appName, log.NewNopLogger(), dbm.NewMemDB(), v.encCfg.TxConfig.TxDecoder())
	rollapp.SetMsgServiceRouter(v.baseApp.MsgServiceRouter())
	rollapp.SetBeginBlocker(v.baseApp.GetBeginBlocker())
	rollapp.SetEndBlocker(v.baseApp.GetEndBlocker())
	v.mutableBaseApp = rollapp
}

func (v *Verifier) Run(fp fraudtypes.FraudProof) error {
	err := v.InitMutableChain(fp)
	if err != nil {
		return err
	}
	err = v.ExecuteProofOnMutableChain(fp)
	if err != nil {
		return err
	}
	return nil
}

// InitMutableChain initializes the Verifier from a fraud proof
// It creates a new mutable app and inits the chain with all required store data
//
// This is inspired by https://github.com/rollkit/cosmos-sdk-old/blob/f6c90a66ed7d8006713ce0781ee0c770d5cc9b71/baseapp/abci.go#L266-L298
func (v *Verifier) InitMutableChain(fp fraudtypes.FraudProof) error {
	if v.baseApp == nil {
		return fmt.Errorf("base app nil")
	}

	v.initCleanInstance()

	if v.mutableBaseApp == nil { // TODO: can remove? Michael had left a nil check todo
		return fmt.Errorf("mutable base app nil")
	}

	v.mutableBaseApp.SetInitialHeight(fp.GetFraudulentBlockHeight())

	cms := v.mutableBaseApp.CommitMultiStore().(*rootmulti.Store)
	modules := fp.GetModules()

	// ~~~~
	// TODO(danwt): this bit is a bit weird, can't we just use the keys from the fp?
	iavlStoreKeys := make([]storetypes.StoreKey, 0, len(modules))
	for _, module := range modules {
		iavlStoreKeys = append(iavlStoreKeys, v.moduleStoreKey(module))
	}
	v.mutableBaseApp.MountStores(iavlStoreKeys...)
	// ~~~~

	moduleStoreKeyToDeepIAVLTree, err := fp.GetModuleStoreKeysToDeepIAVLTree()
	if err != nil {
		return fmt.Errorf("get deep iavl trees: %w", err)
	}
	for moduleStoreKey, iavlTree := range moduleStoreKeyToDeepIAVLTree {
		cms.SetDeepIAVLTree(moduleStoreKey, iavlTree)
	}

	err = v.mutableBaseApp.LoadLatestVersion()
	if err != nil {
		return err
	}

	v.mutableBaseApp.InitChain(abci.RequestInitChain{})

	return nil
}

// ExecuteProofOnMutableChain checks the validity of a given fraud proof.
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
// Note: This function mutates the Verifier
//
// This is inspired by https://github.com/rollkit/cosmos-sdk-old/blob/f6c90a66ed7d8006713ce0781ee0c770d5cc9b71/baseapp/abci.go#L300-L315
func (v *Verifier) ExecuteProofOnMutableChain(fp fraudtypes.FraudProof) error {
	appHash := v.mutableBaseApp.GetAppHashInternal()
	fmt.Println("appHash - prestate", hex.EncodeToString(appHash)) // TODO: remove

	if !bytes.Equal(fp.PreStateAppHash, appHash) {
		return ErrInvalidPreStateAppHash
	}

	setRollappAddressPrefixes("ethm")

	// Execute fraudulent state transition
	if fp.FraudulentBeginBlock != nil {
		panic("fraudulent begin block not supported")
		// v.app.BeginBlock(*fp.FraudulentBeginBlock)
		// fmt.Println("appHash - beginblock", hex.EncodeToString(v.app.GetAppHashInternal()))
	} else {
		// Need to add some dummy begin block here since it's a new app
		v.mutableBaseApp.ResetDeliverState()
		v.mutableBaseApp.SetBeginBlocker(nil)
		v.mutableBaseApp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: fp.GetFraudulentBlockHeight()}})
		fmt.Println("appHash - dummy beginblock", hex.EncodeToString(v.mutableBaseApp.GetAppHashInternal())) // TODO: remove

		if fp.FraudulentDeliverTx != nil {
			// skip IncrementSequenceDecorator check in AnteHandler
			v.mutableBaseApp.SetAnteHandler(nil)

			resp := v.mutableBaseApp.DeliverTx(*fp.FraudulentDeliverTx)
			if !resp.IsOK() {
				panic(resp.Log)
			}
			fmt.Println("appHash - posttx", hex.EncodeToString(v.mutableBaseApp.GetAppHashInternal()))
			setRollappAddressPrefixes("dym")
		} else {
			panic("fraudulent end block not supported")
			// v.app.EndBlock(*fp.FraudulentEndBlock)
			// fmt.Println("appHash - endblock", hex.EncodeToString(v.app.GetAppHashInternal()))
		}
	}

	appHash = v.mutableBaseApp.GetAppHashInternal()
	fmt.Println("appHash - final", hex.EncodeToString(appHash))
	if !bytes.Equal(appHash, fp.ExpectedValidAppHash) {
		return types.ErrInvalidAppHash
	}
	return nil
}

// setRollappAddressPrefixes sets the address prefixes for the rollapp chain
func setRollappAddressPrefixes(prefix string) {
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

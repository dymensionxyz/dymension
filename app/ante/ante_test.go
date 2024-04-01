package ante_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/ethereum/eip712"
	"github.com/evmos/ethermint/testutil"
	ethermint "github.com/evmos/ethermint/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/ante"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/app/params"
)

type AnteTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.App
	clientCtx   client.Context
	anteHandler sdk.AnteHandler
	txBuilder   client.TxBuilder
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *AnteTestSuite) SetupTest(isCheckTx bool) {
	s.app = apptesting.Setup(s.T(), isCheckTx)
	s.ctx = s.app.BaseApp.NewContext(isCheckTx, tmproto.Header{}).WithBlockHeight(1).WithChainID(apptesting.TestChainID)

	txConfig := s.app.GetTxConfig()
	s.clientCtx = client.Context{}.
		WithTxConfig(txConfig).
		WithCodec(s.app.AppCodec())

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   &s.app.AccountKeeper,
			BankKeeper:      s.app.BankKeeper,
			IBCKeeper:       s.app.IBCKeeper,
			EvmKeeper:       s.app.EvmKeeper,
			FeeMarketKeeper: s.app.FeeMarketKeeper,
			TxFeesKeeper:    s.app.TxFeesKeeper,
			FeegrantKeeper:  s.app.FeeGrantKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
		},
	)

	s.Require().NoError(err)
	s.anteHandler = anteHandler
}

func (suite *AnteTestSuite) TestCosmosAnteHandlerEip712() {
	suite.SetupTest(false)
	privkey, _ := ethsecp256k1.GenerateKey()
	key, err := privkey.ToECDSA()
	suite.Require().NoError(err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	
	amt := sdk.NewInt(100)
	err = testutil.FundAccount(
		suite.app.BankKeeper,
		suite.ctx,
		privkey.PubKey().Address().Bytes(),
		sdk.NewCoins(sdk.NewCoin(params.DisplayDenom, amt)),
	)
	suite.Require().NoError(err)

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
	suite.Require().NoError(acc.SetSequence(1))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	from := acc.GetAddress()
	recipient := sdk.AccAddress(common.Address{}.Bytes())
	msgSend := banktypes.NewMsgSend(from, recipient, sdk.NewCoins(sdk.NewCoin(params.DisplayDenom, sdk.NewInt(1))))

	txBuilder := suite.CreateTestEIP712CosmosTxBuilder(privkey, []sdk.Msg{msgSend})
	_, err = suite.anteHandler(suite.ctx, txBuilder.GetTx(), false)

	suite.Require().NoError(err)
}

func (suite *AnteTestSuite) CreateTestEIP712CosmosTxBuilder(
	priv cryptotypes.PrivKey, msgs []sdk.Msg,
) client.TxBuilder {
	txConfig := suite.clientCtx.TxConfig
	coinAmount := sdk.NewCoin(params.DisplayDenom, sdk.NewInt(20))
	fees := sdk.NewCoins(coinAmount)

	pc, err := ethermint.ParseChainID(suite.ctx.ChainID())
	suite.Require().NoError(err)
	chainIDNum := pc.Uint64()

	from := sdk.AccAddress(priv.PubKey().Address().Bytes())
	acc := suite.app.AccountKeeper.GetAccount(suite.ctx, from)
	accNumber := acc.GetAccountNumber()
	nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, from)
	suite.Require().NoError(err)

	suite.txBuilder = txConfig.NewTxBuilder()
	builder, ok := suite.txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok, "txBuilder could not be casted to authtx.ExtensionOptionsTxBuilder type")
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(200000)

	err = builder.SetMsgs(msgs...)
	suite.Require().NoError(err)

	txBytes := legacytx.StdSignBytes(
		suite.ctx.ChainID(),
		accNumber,
		nonce,
		0,
		legacytx.StdFee{
			Amount: fees,
			Gas:    200000,
		},
		msgs, "", nil,
	)

	feeDelegation := &eip712.FeeDelegationOptions{
		FeePayer: from,
	}

	data, err := eip712.LegacyWrapTxToTypedData(
		suite.clientCtx.Codec,
		chainIDNum,
		msgs[0],
		txBytes,
		feeDelegation,
	)
	suite.Require().NoError(err)
	sigHash, _, err := apitypes.TypedDataAndHash(data)
	suite.Require().NoError(err)

	keyringSigner := NewSigner(priv)
	signature, pubKey, err := keyringSigner.SignByAddress(from, sigHash)
	suite.Require().NoError(err)

	signature[crypto.RecoveryIDOffset] += 27

	option, err := codectypes.NewAnyWithValue(&ethermint.ExtensionOptionsWeb3Tx{
		FeePayer:         from.String(),
		TypedDataChainID: chainIDNum,
		FeePayerSig:      signature,
	})
	suite.Require().NoError(err)

	builder.SetExtensionOptions(option)

	sigsV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		},
		Sequence: nonce,
	}
	err = builder.SetSignatures(sigsV2)
	suite.Require().NoError(err)

	return builder
}

// Signer defines a type that is used on testing for signing MsgEthereumTx
type Signer struct {
	privKey cryptotypes.PrivKey
}

func NewSigner(sk cryptotypes.PrivKey) keyring.Signer {
	return &Signer{
		privKey: sk,
	}
}

// Sign signs the message using the underlying private key
func (s Signer) Sign(_ string, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	if s.privKey.Type() != ethsecp256k1.KeyType {
		return nil, nil, fmt.Errorf(
			"invalid private key type for signing ethereum tx; expected %s, got %s",
			ethsecp256k1.KeyType,
			s.privKey.Type(),
		)
	}

	sig, err := s.privKey.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, s.privKey.PubKey(), nil
}

// SignByAddress sign byte messages with a user key providing the address.
func (s Signer) SignByAddress(address sdk.Address, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	signer := sdk.AccAddress(s.privKey.PubKey().Address())
	if !signer.Equals(address) {
		return nil, nil, fmt.Errorf("address mismatch: signer %s ≠ given address %s", signer, address)
	}

	return s.Sign("", msg)
}

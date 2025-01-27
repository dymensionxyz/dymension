package ante_test

import (
	"encoding/json"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	feegrant "github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/ethereum/eip712"
	ethermint "github.com/evmos/ethermint/types"
)

func (s *AnteTestSuite) getMsgSend(from sdk.AccAddress) sdk.Msg {
	privkey2, _ := ethsecp256k1.GenerateKey()
	to := sdk.AccAddress(privkey2.PubKey().Address())
	return banktypes.NewMsgSend(from, to, sdk.NewCoins(sdk.NewCoin(params.DisplayDenom, sdk.NewInt(1))))
}

func (s *AnteTestSuite) getMsgCreateValidator(from sdk.AccAddress) sdk.Msg {
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(from),
		ed25519.GenPrivKey().PubKey(),
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1_000_000_000)),
		stakingtypes.NewDescription("moniker", "indentity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(),
	)
	s.Assert().NoError(err)
	return msgCreate
}

func (s *AnteTestSuite) getMsgGrantEIBC(from sdk.AccAddress) *authz.MsgGrant {
	privkey2, _ := ethsecp256k1.GenerateKey()
	to := sdk.AccAddress(privkey2.PubKey().Address())

	crit := eibctypes.NewRollappCriteria(
		"rollappID",
		[]string{"denom"},
		sdk.DecProto{Dec: sdk.NewDec(1)},
		sdk.Coins{sdk.NewCoin("denom", sdk.NewInt(1))},
		sdk.Coins{sdk.NewCoin("denom", sdk.NewInt(1))},
		sdk.DecProto{Dec: sdk.NewDec(1)},
		true,
	)
	expDate := time.Now().Add(1 * time.Hour)
	msg, err := authz.NewMsgGrant(
		from,
		to,
		eibctypes.NewFulfillOrderAuthorization([]*eibctypes.RollappCriteria{crit}),
		&expDate,
	)
	if err != nil {
		panic(err)
	}
	return msg
}

func (s *AnteTestSuite) getMsgGrant(from sdk.AccAddress) *authz.MsgGrant {
	privkey2, _ := ethsecp256k1.GenerateKey()
	to := sdk.AccAddress(privkey2.PubKey().Address())

	// msgTypeUrl := sdk.MsgTypeURL(&authz.MsgExec{})
	msgTypeUrl := "/dymensionxyz.dymension.gamm.poolmodels.balancer.v1beta1.MsgCreateBalancerPool"
	expDate := time.Now().Add(1 * time.Hour)
	msg, err := authz.NewMsgGrant(
		from,
		to,
		authz.NewGenericAuthorization(msgTypeUrl),
		&expDate,
	)
	if err != nil {
		panic(err)
	}
	return msg
}

func (s *AnteTestSuite) getMsgSubmitProposal(from sdk.AccAddress) sdk.Msg {
	proposal, ok := govtypes.ContentFromProposalType("My proposal", "My description", govtypes.ProposalTypeText)
	s.Require().True(ok)
	deposit := sdk.NewCoins(sdk.NewCoin(params.DisplayDenom, sdk.NewInt(10)))
	msgSubmit, err := govtypes.NewMsgSubmitProposal(proposal, deposit, from)
	s.Require().NoError(err)
	return msgSubmit
}

func (s *AnteTestSuite) getMsgGrantAllowance(from sdk.AccAddress) sdk.Msg {
	spendLimit := sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 10000000))
	threeHours := time.Now().Add(3 * time.Hour)
	basic := &feegrant.BasicAllowance{
		SpendLimit: spendLimit,
		Expiration: &threeHours,
	}

	privkey2, _ := ethsecp256k1.GenerateKey()
	grantee := sdk.AccAddress(privkey2.PubKey().Address())
	msgGrant, err := feegrant.NewMsgGrantAllowance(basic, from, grantee)
	s.Require().NoError(err)
	return msgGrant
}

func (s *AnteTestSuite) getMsgCreateRollapp(from string, tokenless bool, metadata *rollapptypes.RollappMetadata) sdk.Msg {
	genesisInfo := &rollapptypes.GenesisInfo{
		Bech32Prefix:    strings.ToLower(rand.Str(3)),
		GenesisChecksum: "1234567890abcdefg",
		InitialSupply:   sdk.NewInt(1000),
		NativeDenom: rollapptypes.DenomMetadata{
			Display:  "DEN",
			Base:     "aden",
			Exponent: 18,
		},
	}

	if metadata == nil {
		metadata = &rollapptypes.RollappMetadata{
			Website:     "https://dymension.xyz",
			Description: "Sample description",
			LogoUrl:     "https://dymension.xyz/logo.png",
			Telegram:    "https://t.me/rolly",
			X:           "https://x.dymension.xyz",
		}
	}

	if tokenless {
		genesisInfo.InitialSupply = math.ZeroInt()
		genesisInfo.NativeDenom = rollapptypes.DenomMetadata{
			Display:  "",
			Base:     "",
			Exponent: 1,
		}
	}

	return &rollapptypes.MsgCreateRollapp{
		Creator:          from,
		RollappId:        "test_1000-1",
		InitialSequencer: "*",
		MinSequencerBond: rollapptypes.DefaultMinSequencerBondGlobalCoin,
		Alias:            strings.ToLower(rand.Str(7)),
		VmType:           rollapptypes.Rollapp_EVM,
		GenesisInfo:      genesisInfo,
		Metadata:         metadata,
	}
}

func (s *AnteTestSuite) TestEIP712() {
	s.SetupTestCheckTx(false)
	privkey, _ := ethsecp256k1.GenerateKey()
	acc := sdk.AccAddress(privkey.PubKey().Address())

	amt := sdk.NewInt(10000).MulRaw(1e18)
	err := bankutil.FundAccount(s.app.BankKeeper, s.ctx, privkey.PubKey().Address().Bytes(), sdk.NewCoins(sdk.NewCoin(params.BaseDenom, amt)))
	s.Require().Nil(err)

	from := acc
	testCases := []struct {
		description string
		msg         sdk.Msg
		output      bool
	}{
		{"MsgSend", s.getMsgSend(from), false},
		{"MsgCreateRollapp (native denom)", s.getMsgCreateRollapp(from.String(), false, nil), false},
		{"MsgCreateRollapp (tokenless)", s.getMsgCreateRollapp(from.String(), true, nil), false},
		{"MsgGrant", s.getMsgGrant(from), false},
		{"MsgGrantAllowance", s.getMsgGrantAllowance(from), false},
		{"MsgSubmitProposal", s.getMsgSubmitProposal(from), false},
		{"MsgGrantEIBC", s.getMsgGrantEIBC(from), false},
		{"MsgCreateValidator", s.getMsgCreateValidator(from), false},
	}

	for _, tc := range testCases {
		s.Run(tc.description, func() {
			data, err := s.DumpEIP712TypedData(from, []sdk.Msg{tc.msg})
			s.Require().NoError(err)

			// Dump the json string to t.log
			if tc.output {
				str, err := json.MarshalIndent(data, "", "  ") // Indent with 2 spaces
				s.Assert().NoError(err)
				s.T().Log(string(str))
			}
		})
	}
}

func (suite *AnteTestSuite) DumpEIP712TypedData(from sdk.AccAddress, msgs []sdk.Msg) (apitypes.TypedData, error) {
	txConfig := suite.clientCtx.TxConfig
	coinAmount := sdk.NewCoin(params.BaseDenom, sdk.NewInt(20).MulRaw(1e18))
	fees := sdk.NewCoins(coinAmount)

	pc, err := ethermint.ParseChainID(suite.ctx.ChainID())
	suite.Require().NoError(err)
	chainIDNum := pc.Uint64()

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
	_, _, err = apitypes.TypedDataAndHash(data)
	suite.Require().NoError(err)

	return data, nil
}

package ibctesting_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmtypes "github.com/tendermint/tendermint/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/cosmos/ibc-go/v6/testing/mock"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ChainIDPrefix defines the default chain ID prefix for Evmos test chains
var ChainIDPrefix = "evmos_9000-"

func init() {
	ibctesting.ChainIDPrefix = ChainIDPrefix
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		return apptesting.SetupTestingApp()
	}
}

func ConvertToApp(chain *ibctesting.TestChain) *app.App {
	app, ok := chain.App.(*app.App)
	require.True(chain.T, ok)

	return app
}

// IBCTestUtilSuite is a testing suite to test keeper functions.
type IBCTestUtilSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator
}

func hubChainID() string {
	return ibctesting.GetChainID(1)
}

func cosmosChainID() string {
	return ibctesting.GetChainID(2)
}

func rollappChainID() string {
	return ibctesting.GetChainID(3)
}

func (suite *IBCTestUtilSuite) hubChain() *ibctesting.TestChain {
	return suite.coordinator.GetChain(hubChainID())
}

func (suite *IBCTestUtilSuite) cosmosChain() *ibctesting.TestChain {
	return suite.coordinator.GetChain(cosmosChainID())
}

func (suite *IBCTestUtilSuite) rollappChain() *ibctesting.TestChain {
	return suite.coordinator.GetChain(rollappChainID())
}

func (suite *IBCTestUtilSuite) hubApp() *app.App {
	return ConvertToApp(suite.hubChain())
}

func (suite *IBCTestUtilSuite) cosmosApp() *app.App {
	return ConvertToApp(suite.cosmosChain())
}

func (suite *IBCTestUtilSuite) rollappApp() *app.App {
	return ConvertToApp(suite.rollappChain())
}

func (suite *IBCTestUtilSuite) rollappMsgServer() rollapptypes.MsgServer {
	return rollappkeeper.NewMsgServerImpl(suite.hubApp().RollappKeeper)
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *IBCTestUtilSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2) // initializes test chains
	suite.coordinator.Chains[rollappChainID()] = suite.newTestChainWithSingleValidator(suite.T(), suite.coordinator, rollappChainID())
}

// CreateRollappWithFinishedGenesis creates a rollapp whose 'genesis' protocol is complete:
// that is, they have finished all genesis transfers and their bridge is enabled.
func (suite *IBCTestUtilSuite) CreateRollappWithFinishedGenesis(canonicalChannelID string) {
	suite.CreateRollapp(true, &canonicalChannelID)
}

func (suite *IBCTestUtilSuite) CreateRollapp(transfersEnabled bool, channelID *string) {
	msgCreateRollapp := rollapptypes.NewMsgCreateRollapp(
		suite.hubChain().SenderAccount.GetAddress().String(),
		suite.rollappChain().ChainID,
		10,
		[]string{},

		// in most cases we want to test when the genesis bridge setup is already complete
		transfersEnabled,
	)
	_, err := suite.hubChain().SendMsgs(msgCreateRollapp)
	suite.Require().NoError(err) // message committed
	if channelID != nil {
		app := suite.hubApp()
		ra := app.RollappKeeper.MustGetRollapp(suite.hubChain().GetContext(), suite.rollappChain().ChainID)
		ra.ChannelId = *channelID
		app.RollappKeeper.SetRollapp(suite.hubChain().GetContext(), ra)
	}
}

func (suite *IBCTestUtilSuite) RegisterSequencer() {
	bond := sequencertypes.DefaultParams().MinBond
	// fund account
	err := bankutil.FundAccount(suite.hubApp().BankKeeper, suite.hubChain().GetContext(), suite.hubChain().SenderAccount.GetAddress(), sdk.NewCoins(bond))
	suite.Require().Nil(err)

	// using validator pubkey as the dymint pubkey
	pk, err := cryptocodec.FromTmPubKeyInterface(suite.rollappChain().Vals.Validators[0].PubKey)
	suite.Require().Nil(err)

	msgCreateSequencer, err := sequencertypes.NewMsgCreateSequencer(
		suite.hubChain().SenderAccount.GetAddress().String(),
		pk,
		suite.rollappChain().ChainID,
		&sequencertypes.Description{},
		bond,
	)
	suite.Require().NoError(err) // message committed
	_, err = suite.hubChain().SendMsgs(msgCreateSequencer)
	suite.Require().NoError(err) // message committed
}

func (suite *IBCTestUtilSuite) UpdateRollappState(endHeight uint64) {
	// Get the start index and start height based on the latest state info
	rollappKeeper := suite.hubApp().RollappKeeper
	latestStateInfoIndex, _ := rollappKeeper.GetLatestStateInfoIndex(suite.hubChain().GetContext(), suite.rollappChain().ChainID)
	stateInfo, found := rollappKeeper.GetStateInfo(suite.hubChain().GetContext(), suite.rollappChain().ChainID, latestStateInfoIndex.Index)
	startHeight := uint64(1)
	if found {
		startHeight = stateInfo.StartHeight + stateInfo.NumBlocks
	}
	numBlocks := endHeight - startHeight + 1
	// populate the block descriptors
	blockDescriptors := &rollapptypes.BlockDescriptors{BD: make([]rollapptypes.BlockDescriptor, numBlocks)}
	for i := 0; i < int(numBlocks); i++ {
		blockDescriptors.BD[i] = rollapptypes.BlockDescriptor{
			Height:    startHeight + uint64(i),
			StateRoot: bytes.Repeat([]byte{byte(startHeight) + byte(i)}, 32),
		}
	}
	// Update the state
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		suite.hubChain().SenderAccount.GetAddress().String(),
		suite.rollappChain().ChainID,
		startHeight,
		endHeight-startHeight+1, // numBlocks
		"mock-da-path",
		0,
		blockDescriptors,
	)
	err := msgUpdateState.ValidateBasic()
	suite.Require().NoError(err)
	_, err = suite.rollappMsgServer().UpdateState(suite.hubChain().GetContext(), msgUpdateState)
	suite.Require().NoError(err)
}

func (suite *IBCTestUtilSuite) FinalizeRollappState(index uint64, endHeight uint64) (sdk.Events, error) {
	rollappKeeper := suite.hubApp().RollappKeeper
	ctx := suite.hubChain().GetContext()

	stateInfoIdx := rollapptypes.StateInfoIndex{RollappId: suite.rollappChain().ChainID, Index: index}
	stateInfo, found := rollappKeeper.GetStateInfo(ctx, suite.rollappChain().ChainID, stateInfoIdx.Index)
	suite.Require().True(found)
	stateInfo.NumBlocks = endHeight - stateInfo.StartHeight + 1
	stateInfo.Status = common.Status_FINALIZED
	// update the status of the stateInfo
	rollappKeeper.SetStateInfo(ctx, stateInfo)
	// update the LatestStateInfoIndex of the rollapp
	rollappKeeper.SetLatestFinalizedStateIndex(ctx, stateInfoIdx)
	err := rollappKeeper.GetHooks().AfterStateFinalized(
		ctx,
		suite.rollappChain().ChainID,
		&stateInfo,
	)

	return ctx.EventManager().Events(), err
}

func (suite *IBCTestUtilSuite) NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}

func (suite *IBCTestUtilSuite) GetRollappToHubIBCDenomFromPacket(packet channeltypes.Packet) string {
	var data transfertypes.FungibleTokenPacketData
	err := eibctypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data)
	suite.Require().NoError(err)
	return suite.GetIBCDenomForChannel(packet.GetDestChannel(), data.Denom)
}

func (suite *IBCTestUtilSuite) GetIBCDenomForChannel(channel string, denom string) string {
	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := types.GetDenomPrefix("transfer", channel)
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + denom
	// construct the denomination trace from the full raw denomination
	denomTrace := types.ParseDenomTrace(prefixedDenom)
	return denomTrace.IBCDenom()
}

func (suite *IBCTestUtilSuite) newTestChainWithSingleValidator(t *testing.T, coord *ibctesting.Coordinator, chainID string) *ibctesting.TestChain {
	genAccs := []authtypes.GenesisAccount{}
	genBals := []banktypes.Balance{}
	senderAccs := []ibctesting.SenderAccount{}

	// generate genesis accounts

	valPrivKey := mock.NewPV()
	valPubKey, err := valPrivKey.GetPubKey()
	suite.Require().NoError(err)

	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)

	amount, ok := sdk.NewIntFromString("10000000000000000000")
	suite.Require().True(ok)

	// add sender account
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amount)),
	}

	genAccs = append(genAccs, acc)
	genBals = append(genBals, balance)

	senderAcc := ibctesting.SenderAccount{
		SenderAccount: acc,
		SenderPrivKey: senderPrivKey,
	}

	senderAccs = append(senderAccs, senderAcc)

	var validators []*tmtypes.Validator
	signersByAddress := make(map[string]tmtypes.PrivValidator, 1)

	validators = append(validators, tmtypes.NewValidator(valPubKey, 1))

	signersByAddress[valPubKey.Address().String()] = valPrivKey
	valSet := tmtypes.NewValidatorSet(validators)

	app := ibctesting.SetupWithGenesisValSet(t, valSet, genAccs, chainID, sdk.DefaultPowerReduction, genBals...)

	// create current header and call begin block
	header := tmproto.Header{
		ChainID: chainID,
		Height:  1,
		Time:    coord.CurrentTime.UTC(),
	}

	txConfig := app.GetTxConfig()

	// create an account to send transactions from
	chain := &ibctesting.TestChain{
		T:              t,
		Coordinator:    coord,
		ChainID:        chainID,
		App:            app,
		CurrentHeader:  header,
		QueryServer:    app.GetIBCKeeper(),
		TxConfig:       txConfig,
		Codec:          app.AppCodec(),
		Vals:           valSet,
		NextVals:       valSet,
		Signers:        signersByAddress,
		SenderPrivKey:  senderAcc.SenderPrivKey,
		SenderAccount:  senderAcc.SenderAccount,
		SenderAccounts: senderAccs,
	}

	coord.CommitBlock(chain)

	return chain
}

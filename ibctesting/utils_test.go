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

// chainIDPrefix defines the default chain ID prefix for Evmos test chains
var chainIDPrefix = "evmos_9000-"

func init() {
	ibctesting.ChainIDPrefix = chainIDPrefix
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		return apptesting.SetupTestingApp()
	}
}

func convertToApp(chain *ibctesting.TestChain) *app.App {
	a, ok := chain.App.(*app.App)
	require.True(chain.T, ok)

	return a
}

// utilSuite is a testing suite to test keeper functions.
type utilSuite struct {
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

func (s *utilSuite) hubChain() *ibctesting.TestChain {
	return s.coordinator.GetChain(hubChainID())
}

func (s *utilSuite) cosmosChain() *ibctesting.TestChain {
	return s.coordinator.GetChain(cosmosChainID())
}

func (s *utilSuite) rollappChain() *ibctesting.TestChain {
	return s.coordinator.GetChain(rollappChainID())
}

func (s *utilSuite) hubApp() *app.App {
	return convertToApp(s.hubChain())
}

func (s *utilSuite) rollappApp() *app.App {
	return convertToApp(s.rollappChain())
}

func (s *utilSuite) hubCtx() sdk.Context {
	return s.hubChain().GetContext()
}

func (s *utilSuite) cosmosCtx() sdk.Context {
	return s.cosmosChain().GetContext()
}

func (s *utilSuite) rollappCtx() sdk.Context {
	return s.rollappChain().GetContext()
}

func (s *utilSuite) rollappMsgServer() rollapptypes.MsgServer {
	return rollappkeeper.NewMsgServerImpl(s.hubApp().RollappKeeper)
}

// SetupTest creates a coordinator with 2 test chains.
func (s *utilSuite) SetupTest() {
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2) // initializes test chains
	s.coordinator.Chains[rollappChainID()] = s.newTestChainWithSingleValidator(s.T(), s.coordinator, rollappChainID())
}

// CreateRollappWithFinishedGenesis creates a rollapp whose 'genesis' protocol is complete:
// that is, they have finished all genesis transfers and their bridge is enabled.
func (s *utilSuite) createRollappWithFinishedGenesis(canonicalChannelID string) {
	s.createRollapp(true, &canonicalChannelID)
}

func (s *utilSuite) createRollapp(transfersEnabled bool, channelID *string) {
	msgCreateRollapp := rollapptypes.NewMsgCreateRollapp(s.hubChain().SenderAccount.GetAddress().String(), rollappChainID(), 10, []string{})
	_, err := s.hubChain().SendMsgs(msgCreateRollapp)
	s.Require().NoError(err) // message committed
	if channelID != nil {
		a := s.hubApp()
		ra := a.RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
		ra.ChannelId = *channelID
		ra.GenesisState.TransfersEnabled = transfersEnabled
		a.RollappKeeper.SetRollapp(s.hubCtx(), ra)
	}
}

func (s *utilSuite) registerSequencer() {
	bond := sequencertypes.DefaultParams().MinBond
	// fund account
	err := bankutil.FundAccount(s.hubApp().BankKeeper, s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), sdk.NewCoins(bond))
	s.Require().Nil(err)

	// using validator pubkey as the dymint pubkey
	pk, err := cryptocodec.FromTmPubKeyInterface(s.rollappChain().Vals.Validators[0].PubKey)
	s.Require().Nil(err)

	msgCreateSequencer, err := sequencertypes.NewMsgCreateSequencer(
		s.hubChain().SenderAccount.GetAddress().String(),
		pk,
		rollappChainID(),
		&sequencertypes.Description{},
		bond,
	)
	s.Require().NoError(err) // message committed
	_, err = s.hubChain().SendMsgs(msgCreateSequencer)
	s.Require().NoError(err) // message committed
}

func (s *utilSuite) updateRollappState(endHeight uint64) {
	// Get the start index and start height based on the latest state info
	rollappKeeper := s.hubApp().RollappKeeper
	latestStateInfoIndex, _ := rollappKeeper.GetLatestStateInfoIndex(s.hubCtx(), rollappChainID())
	stateInfo, found := rollappKeeper.GetStateInfo(s.hubCtx(), rollappChainID(), latestStateInfoIndex.Index)
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
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		startHeight,
		endHeight-startHeight+1, // numBlocks
		"mock-da-path",
		0,
		blockDescriptors,
	)
	err := msgUpdateState.ValidateBasic()
	s.Require().NoError(err)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Require().NoError(err)
}

func (s *utilSuite) finalizeRollappState(index uint64, endHeight uint64) (sdk.Events, error) {
	rollappKeeper := s.hubApp().RollappKeeper
	ctx := s.hubCtx()

	stateInfoIdx := rollapptypes.StateInfoIndex{RollappId: rollappChainID(), Index: index}
	stateInfo, found := rollappKeeper.GetStateInfo(ctx, rollappChainID(), stateInfoIdx.Index)
	s.Require().True(found)
	stateInfo.NumBlocks = endHeight - stateInfo.StartHeight + 1
	stateInfo.Status = common.Status_FINALIZED
	// update the status of the stateInfo
	rollappKeeper.SetStateInfo(ctx, stateInfo)
	// update the LatestStateInfoIndex of the rollapp
	rollappKeeper.SetLatestFinalizedStateIndex(ctx, stateInfoIdx)
	err := rollappKeeper.GetHooks().AfterStateFinalized(
		ctx,
		rollappChainID(),
		&stateInfo,
	)

	return ctx.EventManager().Events(), err
}

func (s *utilSuite) newTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}

func (s *utilSuite) getRollappToHubIBCDenomFromPacket(packet channeltypes.Packet) string {
	var data transfertypes.FungibleTokenPacketData
	err := eibctypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data)
	s.Require().NoError(err)
	return s.getIBCDenomForChannel(packet.GetDestChannel(), data.Denom)
}

func (s *utilSuite) getIBCDenomForChannel(channel string, denom string) string {
	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := types.GetDenomPrefix("transfer", channel)
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + denom
	// construct the denomination trace from the full raw denomination
	denomTrace := types.ParseDenomTrace(prefixedDenom)
	return denomTrace.IBCDenom()
}

func (s *utilSuite) newTestChainWithSingleValidator(t *testing.T, coord *ibctesting.Coordinator, chainID string) *ibctesting.TestChain {
	genAccs := []authtypes.GenesisAccount{}
	genBals := []banktypes.Balance{}
	senderAccs := []ibctesting.SenderAccount{}

	// generate genesis accounts

	valPrivKey := mock.NewPV()
	valPubKey, err := valPrivKey.GetPubKey()
	s.Require().NoError(err)

	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)

	amount, ok := sdk.NewIntFromString("10000000000000000000")
	s.Require().True(ok)

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

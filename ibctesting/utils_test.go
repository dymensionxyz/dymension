package ibctesting_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cometbfttypes "github.com/cometbft/cometbft/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/cosmos/ibc-go/v8/testing/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	denomutils "github.com/dymensionxyz/dymension/v3/utils/denom"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// chainIDPrefix defines the default chain ID prefix for Evmos test chains
var chainIDPrefix = "evmos_9000"

func init() {
	ibctesting.ChainIDPrefix = chainIDPrefix
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		return apptesting.SetupTestingApp()
	}
}

func convertToApp(chain *ibctesting.TestChain) *app.App {
	a, ok := chain.App.(*app.App)
	require.True(chain.TB, ok)

	return a
}

// ibcTestingSuite is a testing suite to test IBC middlewares and functions.
type ibcTestingSuite struct {
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

func (s *ibcTestingSuite) hubChain() *ibctesting.TestChain {
	return s.coordinator.GetChain(hubChainID())
}

func (s *ibcTestingSuite) cosmosChain() *ibctesting.TestChain {
	return s.coordinator.GetChain(cosmosChainID())
}

func (s *ibcTestingSuite) rollappChain() *ibctesting.TestChain {
	return s.coordinator.GetChain(rollappChainID())
}

func (s *ibcTestingSuite) hubApp() *app.App {
	return convertToApp(s.hubChain())
}

func (s *ibcTestingSuite) rollappApp() *app.App {
	return convertToApp(s.rollappChain())
}

func (s *ibcTestingSuite) hubCtx() sdk.Context {
	return s.hubChain().GetContext()
}

func (s *ibcTestingSuite) cosmosCtx() sdk.Context {
	return s.cosmosChain().GetContext()
}

func (s *ibcTestingSuite) rollappCtx() sdk.Context {
	return s.rollappChain().GetContext()
}

func (s *ibcTestingSuite) rollappMsgServer() rollapptypes.MsgServer {
	return rollappkeeper.NewMsgServerImpl(s.hubApp().RollappKeeper)
}

// SetupTest creates a coordinator with 2 test chains.
func (s *ibcTestingSuite) SetupTest() {
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2) // initializes test chains
	s.coordinator.Chains[rollappChainID()] = s.newTestChainWithSingleValidator(s.T(), s.coordinator, rollappChainID())
}

// CreateRollappWithFinishedGenesis creates a rollapp whose 'genesis' protocol is complete:
// that is, they have finished all genesis transfers and their bridge is enabled.
func (s *ibcTestingSuite) createRollappWithFinishedGenesis(canonicalChannelID string) {
	s.createRollapp(true, &canonicalChannelID)
}

func (s *ibcTestingSuite) createRollapp(transfersEnabled bool, channelID *string) {
	msgCreateRollapp := rollapptypes.NewMsgCreateRollapp(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		s.hubChain().SenderAccount.GetAddress().String(),
		rollapptypes.DefaultMinSequencerBondGlobalCoin,
		strings.ToLower(tmrand.Str(7)),
		rollapptypes.Rollapp_EVM,
		&rollapptypes.RollappMetadata{
			Website:     "http://example.com",
			Description: "Some description",
			LogoUrl:     "https://dymension.xyz/logo.png",
			Telegram:    "https://t.me/rolly",
			X:           "https://x.dymension.xyz",
		},
		&rollapptypes.GenesisInfo{
			GenesisChecksum: "checksum",
			Bech32Prefix:    "ethm",
			NativeDenom: rollapptypes.DenomMetadata{
				Display:  "DEN",
				Base:     "aden",
				Exponent: 18,
			},
			InitialSupply: math.NewInt(1_000_000_000).MulRaw(1e18),
		},
	)

	apptesting.FundForAliasRegistration(
		s.hubApp(), s.hubCtx(), msgCreateRollapp.Alias, msgCreateRollapp.Creator,
	)

	_, err := s.hubChain().SendMsgs(msgCreateRollapp)
	s.Require().NoError(err) // message committed
	if channelID != nil {
		a := s.hubApp()
		ra := a.RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
		ra.ChannelId = *channelID
		ra.GenesisState.TransferProofHeight = 0
		if transfersEnabled {
			ra.GenesisState.TransferProofHeight = 1
		}
		a.RollappKeeper.SetRollapp(s.hubCtx(), ra)
	}

	// for some reason, the ibctesting frameworks creates headers with App version=2
	// we use this field as revision number, so it breaks the tests as the expected revision number is 0
	// this is an hack to fix the tests
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	rollapp.Revisions[0].Number = 2
	s.hubApp().RollappKeeper.SetRollapp(s.hubCtx(), rollapp)
}

// necessary for tests which do not execute the entire light client flow, and just need to make transfers work
// (all tests except the light client tests themselves)
func (s *ibcTestingSuite) setRollappLightClientID(chainID, clientID string) {
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), chainID, clientID)
}

func (s *ibcTestingSuite) registerSequencer() {
	bond := rollapptypes.DefaultMinSequencerBondGlobalCoin
	// fund account
	apptesting.FundAccount(s.hubApp(), s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), sdk.NewCoins(bond))

	// using validator pubkey as the dymint pubkey
	pk, err := cryptocodec.FromTmPubKeyInterface(s.rollappChain().Vals.Validators[0].PubKey)
	s.Require().Nil(err)

	msgCreateSequencer, err := sequencertypes.NewMsgCreateSequencer(
		s.hubChain().SenderAccount.GetAddress().String(),
		pk,
		rollappChainID(),
		&sequencertypes.SequencerMetadata{
			Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
			EvmRpcs:     []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
			RestApiUrls: []string{"https://api.wpd.evm.rollapp.noisnemyd.xyz:443"},
		},
		bond,
		s.hubChain().SenderAccount.GetAddress().String(),
		[]string{},
	)
	s.Require().NoError(err) // message committed
	_, err = s.hubChain().SendMsgs(msgCreateSequencer)
	s.Require().NoError(err) // message committed
}

// this method sets "dummy" block descriptors for the rollapp
// make sure to setup s.hubApp().LightClientKeeper.SetEnabled(false) before calling this method
func (s *ibcTestingSuite) updateRollappState(endHeight uint64) {
	// Get the start index and start height based on the latest state info
	rollappKeeper := s.hubApp().RollappKeeper
	revision := rollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).Revisions[0].Number
	latestStateInfoIndex, _ := rollappKeeper.GetLatestStateInfoIndex(s.hubCtx(), rollappChainID())
	stateInfo, found := rollappKeeper.GetStateInfo(s.hubCtx(), rollappChainID(), latestStateInfoIndex.Index)
	startHeight := uint64(1)
	if found {
		startHeight = stateInfo.StartHeight + stateInfo.NumBlocks
	}
	numBlocks := endHeight - startHeight + 1
	// populate the block descriptors
	blockDescriptors := &rollapptypes.BlockDescriptors{BD: make([]rollapptypes.BlockDescriptor, numBlocks)}
	for i := uint64(0); i < numBlocks; i++ {
		blockDescriptors.BD[i] = rollapptypes.BlockDescriptor{
			Height:     startHeight + i,
			StateRoot:  bytes.Repeat([]byte{byte(startHeight) + byte(i)}, 32),
			Timestamp:  time.Now().UTC(),
			DrsVersion: 1,
		}
	}
	// Update the state
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		startHeight,
		endHeight-startHeight+1, // numBlocks
		revision,
		blockDescriptors,
	)
	err := msgUpdateState.ValidateBasic()
	s.Require().NoError(err)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Require().NoError(err)
}

// NOTE: does not use process the queue, it uses intrusive method which breaks invariants
func (s *ibcTestingSuite) finalizeRollappState(index uint64, endHeight uint64) (sdk.Events, error) {
	rollappKeeper := s.hubApp().RollappKeeper
	ctx := s.hubCtx()

	stateInfoIdx := rollapptypes.StateInfoIndex{RollappId: rollappChainID(), Index: index}
	stateInfo, found := rollappKeeper.GetStateInfo(ctx, rollappChainID(), stateInfoIdx.Index)
	s.Require().True(found)
	// this is a hack to increase the finalized height by modifying the last state info instead of submitting a new one
	stateInfo.NumBlocks = endHeight - stateInfo.StartHeight + 1
	stateInfo.BDs.BD = make([]rollapptypes.BlockDescriptor, stateInfo.NumBlocks)
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

func (s *ibcTestingSuite) newTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

func (s *ibcTestingSuite) getRollappToHubIBCDenomFromPacket(packet channeltypes.Packet) string {
	var data transfertypes.FungibleTokenPacketData
	err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data)
	s.Require().NoError(err)

	return denomutils.GetIncomingTransferDenom(packet, data)
}

func (s *ibcTestingSuite) newTestChainWithSingleValidator(t *testing.T, coord *ibctesting.Coordinator, chainID string) *ibctesting.TestChain {
	genAccs := []authtypes.GenesisAccount{}
	genBals := []banktypes.Balance{}
	senderAccs := []ibctesting.SenderAccount{}

	// generate genesis accounts

	valPrivKey := mock.NewPV()
	valPubKey, err := valPrivKey.GetPubKey()
	s.Require().NoError(err)

	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)

	amount, ok := math.NewIntFromString("10000000000000000000")
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

	var validators []*cometbfttypes.Validator
	signersByAddress := make(map[string]cometbfttypes.PrivValidator, 1)

	validators = append(validators, cometbfttypes.NewValidator(valPubKey, 1))

	signersByAddress[valPubKey.Address().String()] = valPrivKey
	valSet := cometbfttypes.NewValidatorSet(validators)

	app := ibctesting.SetupWithGenesisValSet(t, valSet, genAccs, chainID, sdk.DefaultPowerReduction, genBals...)

	// create current header and call begin block
	header := cometbftproto.Header{
		ChainID: chainID,
		Height:  1,
		Time:    coord.CurrentTime.UTC(),
	}

	txConfig := app.GetTxConfig()

	// create an account to send transactions from
	chain := &ibctesting.TestChain{
		TB:             t,
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

func (s *ibcTestingSuite) finalizeRollappPacketsByAddress(address string) sdk.Events {
	s.T().Helper()
	// Query all pending packets by address
	querier := delayedackkeeper.NewQuerier(s.hubApp().DelayedAckKeeper)
	resp, err := querier.GetPendingPacketsByAddress(s.hubCtx(), &delayedacktypes.QueryPendingPacketsByAddressRequest{
		Address: address,
	})
	s.Require().NoError(err)
	// Finalize all packets and collect events
	events := make(sdk.Events, len(resp.RollappPackets))
	for _, packet := range resp.RollappPackets {
		k := common.EncodePacketKey(packet.RollappPacketKey())
		handler := s.hubApp().MsgServiceRouter().Handler(new(delayedacktypes.MsgFinalizePacketByPacketKey))
		resp, err := handler(s.hubCtx(), &delayedacktypes.MsgFinalizePacketByPacketKey{
			Sender:    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			PacketKey: k,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		events = append(events, resp.GetEvents()...)
	}
	return events
}

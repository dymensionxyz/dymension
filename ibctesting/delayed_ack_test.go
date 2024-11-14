package ibctesting_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcmerkle "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/cosmos/ibc-go/v7/testing/simapp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	disabledTimeoutTimestamp = uint64(0)
)

type delayedAckSuite struct {
	utilSuite
}

func TestDelayedAckTestSuite(t *testing.T) {
	suite.Run(t, new(delayedAckSuite))
}

func (s *delayedAckSuite) SetupTest() {
	s.utilSuite.SetupTest()
	s.hubApp().LightClientKeeper.SetEnabled(false)

	s.hubApp().BankKeeper.SetDenomMetaData(s.hubCtx(), banktypes.Metadata{
		Base: sdk.DefaultBondDenom,
	})
}

// Transfer from cosmos chain to the hub. No delay expected
func (s *delayedAckSuite) TestTransferCosmosToHub() {
	// setup between cosmosChain and hubChain
	path := s.newTransferPath(s.hubChain(), s.cosmosChain())
	s.coordinator.Setup(path)

	cosmosEndpoint := path.EndpointB
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	s.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	// send from cosmosChain to hubChain
	msg := types.NewMsgTransfer(cosmosEndpoint.ChannelConfig.PortID, cosmosEndpoint.ChannelID, coinToSendToB, s.cosmosChain().SenderAccount.GetAddress().String(), s.hubChain().SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := s.cosmosChain().SendMsgs(msg)
	s.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	s.Require().NoError(err) // relay committed

	found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
}

func (s *delayedAckSuite) TestTransferHubToCosmos() {
	// setup between cosmosChain and hubChain
	path := s.newTransferPath(s.hubChain(), s.cosmosChain())
	s.coordinator.Setup(path)

	hubEndpoint := path.EndpointA
	cosmosIBCKeeper := s.cosmosChain().App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	s.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	// send from cosmosChain to hubChain
	msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, s.hubChain().SenderAccount.GetAddress().String(), s.cosmosChain().SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	s.Require().NoError(err) // relay committed

	found := cosmosIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.cosmosCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
}

func (s *delayedAckSuite) TestTransferRollappToHubNotFinalized() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	rollappEndpoint := path.EndpointB
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()

	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight()))

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	s.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	msg := types.NewMsgTransfer(
		rollappEndpoint.ChannelConfig.PortID,
		rollappEndpoint.ChannelID,
		coinToSendToB,
		s.rollappChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		"",
	)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	s.Require().Error(err) // expecting error as no AcknowledgePacket expected
	found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().False(found)
}

func (s *delayedAckSuite) TestTransferRollappToHubFinalization() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()

	rollappEndpoint := path.EndpointB
	rollappIBCKeeper := s.rollappChain().App.GetIBCKeeper()

	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()

	// Update rollapp state
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight())
	s.updateRollappState(currentRollappBlockHeight)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	s.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	/* --------------------- initiating transfer on rollapp --------------------- */
	msg := types.NewMsgTransfer(rollappEndpoint.ChannelConfig.PortID, rollappEndpoint.ChannelID, coinToSendToB, s.rollappChain().SenderAccount.GetAddress().String(), s.hubChain().SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err) // message committed
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	found := rollappIBCKeeper.ChannelKeeper.HasPacketCommitment(s.rollappCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found)

	// relay send
	err = path.RelayPacket(packet)
	s.Require().Error(err) // expecting error as no AcknowledgePacket expected to return

	found = hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().False(found)

	// Finalize the rollapp state
	currentRollappBlockHeight = uint64(s.rollappCtx().BlockHeight())
	_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
	s.Require().NoError(err)

	// manually finalize packets through x/delayedack
	s.finalizeRollappPacketsByAddress(s.hubChain().SenderAccount.GetAddress().String())

	// Validate ack is found
	found = hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
}

// TestHubToRollappTimeout tests the scenario where a packet is sent from the hub to the rollapp and the rollapp times out the packet.
// The packet should actually get timed out and funds returned to the user only after the rollapp state is finalized.
func (s *delayedAckSuite) TestHubToRollappTimeout() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	// Setup endpoints
	hubEndpoint := path.EndpointA
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()
	// Create rollapp and update its initial state
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight()))
	// Set the timeout height
	timeoutHeight := clienttypes.GetSelfHeight(s.rollappCtx())
	amount, ok := sdk.NewIntFromString("1000000000000000000") // 1DYM
	s.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	// Setup accounts
	senderAccount := s.hubChain().SenderAccount.GetAddress()
	receiverAccount := s.rollappChain().SenderAccount.GetAddress()
	// Check balances
	bankKeeper := s.hubApp().BankKeeper
	preSendBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	// send from hubChain to rollappChain
	msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, senderAccount.String(), receiverAccount.String(), timeoutHeight, disabledTimeoutTimestamp, "")
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found)
	// Check balance decreased
	postSendBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().Equal(preSendBalance.Amount.Sub(coinToSendToB.Amount), postSendBalance.Amount)
	// Update the client to create timeout
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)
	// Timeout the packet. Shouldn't release funds until rollapp height is finalized
	err = path.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err)
	// Validate funds are still not returned to the sender
	postTimeoutBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().Equal(postSendBalance.Amount, postTimeoutBalance.Amount)
	// Finalize the rollapp state
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight())
	_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
	s.Require().NoError(err)
	// manually finalize packets through x/delayedack
	s.finalizeRollappPacketsByAddress(senderAccount.String())
	// Validate funds are returned to the sender
	postFinalizeBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().Equal(preSendBalance.Amount, postFinalizeBalance.Amount)
}

// TestHardFork tests the hard fork handling for outgoing packets from the hub to the rollapp.
// we assert the packets commitments are restored and the pending packets are ackable after the hard fork.
func (s *delayedAckSuite) TestHardFork_HubToRollapp() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	// Setup endpoints
	var (
		hubEndpoint     = path.EndpointA
		hubIBCKeeper    = s.hubChain().App.GetIBCKeeper()
		senderAccount   = s.hubChain().SenderAccount.GetAddress()
		receiverAccount = s.rollappChain().SenderAccount.GetAddress()

		amount, _     = sdk.NewIntFromString("1000000000000000000") // 1DYM
		coinToSendToB = sdk.NewCoin(sdk.DefaultBondDenom, amount)
		timeoutHeight = clienttypes.Height{RevisionNumber: 1, RevisionHeight: 50}
	)

	// Create rollapp and update its initial state
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight()))

	// send from hubChain to rollappChain
	balanceBefore := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, senderAccount.String(), receiverAccount.String(), timeoutHeight, disabledTimeoutTimestamp, "")
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// assert commitments are created
	found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found)

	// Update the client
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	err = path.RelayPacket(packet)
	s.Require().NoError(err) // expecting error as no AcknowledgePacket expected to return

	// progress the rollapp chain
	s.coordinator.CommitNBlocks(s.rollappChain(), 110)

	// Update the client
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	// write ack optimistically
	err = path.EndpointA.AcknowledgePacket(packet, []byte{0x1})
	s.Require().NoError(err)

	// assert commitments are no longer available
	found = hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().False(found)

	// timeout the packet, can't check for error (ErrNoOp). we assert the balance refund
	err = path.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err)
	balanceAfter := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().NotEqual(balanceBefore.String(), balanceAfter.String())

	// hard fork
	err = s.hubApp().DelayedAckKeeper.OnHardFork(s.hubCtx(), s.rollappCtx().ChainID(), 5)
	s.Require().NoError(err)

	// assert commitments are created again
	found = hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found)

	// Update the client
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	// timeout the packet. we expect for verification error
	timeoutMsg := getTimeOutPacket(hubEndpoint, packet)
	_, _, err = simapp.SignAndDeliver(
		path.EndpointA.Chain.T,
		path.EndpointA.Chain.TxConfig,
		path.EndpointA.Chain.App.GetBaseApp(),
		path.EndpointA.Chain.GetContext().BlockHeader(),
		[]sdk.Msg{timeoutMsg},
		path.EndpointA.Chain.ChainID,
		[]uint64{path.EndpointA.Chain.SenderAccount.GetAccountNumber()},
		[]uint64{path.EndpointA.Chain.SenderAccount.GetSequence()},
		true, false, path.EndpointA.Chain.SenderPrivKey,
	)
	s.Require().ErrorIs(err, ibcmerkle.ErrInvalidProof)
}

func getTimeOutPacket(endpoint *ibctesting.Endpoint, packet channeltypes.Packet) *channeltypes.MsgTimeout {
	packetKey := host.PacketReceiptKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	counterparty := endpoint.Counterparty
	proof, proofHeight := counterparty.QueryProof(packetKey)
	nextSeqRecv, found := counterparty.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceRecv(counterparty.Chain.GetContext(), counterparty.ChannelConfig.PortID, counterparty.ChannelID)
	require.True(endpoint.Chain.T, found)

	timeoutMsg := channeltypes.NewMsgTimeout(
		packet, nextSeqRecv,
		proof, proofHeight, endpoint.Chain.SenderAccount.GetAddress().String(),
	)

	return timeoutMsg
}

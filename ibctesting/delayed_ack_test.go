package ibctesting_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcmerkle "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	disabledTimeoutTimestamp = uint64(0)
	// Standard amount for 10 DYM
	TenDymAmountString = "10000000000000000000"
	// Standard amount for 1 DYM
	OneDymAmountString = "1000000000000000000"
)

type delayedAckSuite struct {
	ibcTestingSuite
}

func TestDelayedAckTestSuite(t *testing.T) {
	suite.Run(t, new(delayedAckSuite))
}

func (s *delayedAckSuite) SetupTest() {
	s.ibcTestingSuite.SetupTest()
	// Disable light client verification on the hub for delayed ack testing
	s.hubApp().LightClientKeeper.SetEnabled(false)

	s.hubApp().BankKeeper.SetDenomMetaData(s.hubCtx(), banktypes.Metadata{
		Base: sdk.DefaultBondDenom,
	})
}

// createTransferCoin creates a standard coin to send (10DYM).
func (s *delayedAckSuite) createTransferCoin(amountStr string) sdk.Coin {
	amount, ok := math.NewIntFromString(amountStr)
	s.Require().True(ok, "failed to create Int from string: %s", amountStr)
	return sdk.NewCoin(sdk.DefaultBondDenom, amount)
}

// extractPacket extracts the IBC packet from the transaction events.
func (s *delayedAckSuite) extractPacket(res *sdk.Result) channeltypes.Packet {
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err, "failed to parse packet from events")
	return packet
}

// Transfer from cosmos chain to the hub. No delay expected
func (s *delayedAckSuite) TestTransferCosmosToHub() {
	// setup between cosmosChain and hubChain
	path := s.newTransferPath(s.hubChain(), s.cosmosChain())
	s.coordinator.Setup(path)

	cosmosEndpoint := path.EndpointB
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	coinToSendToB := s.createTransferCoin(TenDymAmountString)

	// send from cosmosChain to hubChain
	// Final "" is for memo field
	msg := types.NewMsgTransfer(
		cosmosEndpoint.ChannelConfig.PortID,
		cosmosEndpoint.ChannelID,
		coinToSendToB,
		s.cosmosChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		"",
	)
	res, err := s.cosmosChain().SendMsgs(msg)
	s.Require().NoError(err, "transfer message committed on source chain")

	packet := s.extractPacket(res)

	// relay send (should succeed immediately as it's not a rollapp)
	err = path.RelayPacket(packet)
	s.Require().NoError(err, "relay committed successfully")

	// Validate ack is found on the destination chain (Hub)
	found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found, "expected packet acknowledgement to be found on Hub")
}

// Transfer from hub to cosmos chain. No delay expected
func (s *delayedAckSuite) TestTransferHubToCosmos() {
	// setup between cosmosChain and hubChain
	path := s.newTransferPath(s.hubChain(), s.cosmosChain())
	s.coordinator.Setup(path)

	hubEndpoint := path.EndpointA
	cosmosIBCKeeper := s.cosmosChain().App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	coinToSendToB := s.createTransferCoin(TenDymAmountString)

	// send from hubChain to cosmosChain
	msg := types.NewMsgTransfer(
		hubEndpoint.ChannelConfig.PortID,
		hubEndpoint.ChannelID,
		coinToSendToB,
		s.hubChain().SenderAccount.GetAddress().String(),
		s.cosmosChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		"",
	)
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err, "transfer message committed on source chain")

	packet := s.extractPacket(res)

	// relay send (should succeed immediately as it's not a rollapp)
	err = path.RelayPacket(packet)
	s.Require().NoError(err, "relay committed successfully")

	// Validate ack is found on the destination chain (Cosmos)
	found := cosmosIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.cosmosCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found, "expected packet acknowledgement to be found on Cosmos chain")
}

// Tests a transfer from a rollapp to the hub where the rollapp state is NOT finalized.
func (s *delayedAckSuite) TestTransferRollappToHubNotFinalized() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	rollappEndpoint := path.EndpointB
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()

	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()

	// Use safe block height conversion for rollapp context
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight()) 
	s.updateRollappState(currentRollappBlockHeight) 

	timeoutHeight := clienttypes.NewHeight(100, 110)
	coinToSendToB := s.createTransferCoin(TenDymAmountString)

	// Initiate transfer on the rollapp (source chain)
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
	s.Require().NoError(err, "transfer message committed on rollapp")

	packet := s.extractPacket(res)

	// relay send (should fail because the acknowledgement is delayed/not finalized)
	err = path.RelayPacket(packet)
	s.Require().Error(err, "expected relay to fail as no AcknowledgePacket is expected yet")

	// Validate ack is NOT found on the destination chain (Hub)
	found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().False(found, "expected packet acknowledgement NOT to be found on Hub")
}

// Tests a complete transfer cycle from rollapp to hub, including finalization.
func (s *delayedAckSuite) TestTransferRollappToHubFinalization() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()
	rollappEndpoint := path.EndpointB
	rollappIBCKeeper := s.rollappChain().App.GetIBCKeeper()

	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()

	// 1. Update rollapp state (to get initial proof)
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight()) 
	s.updateRollappState(currentRollappBlockHeight)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	coinToSendToB := s.createTransferCoin(TenDymAmountString)

	/* --------------------- initiating transfer on rollapp --------------------- */
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
	s.Require().NoError(err, "transfer message committed on rollapp")
	packet := s.extractPacket(res)

	// Assert packet commitment exists on the source chain (rollapp)
	found := rollappIBCKeeper.ChannelKeeper.HasPacketCommitment(s.rollappCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found, "packet commitment not found on rollapp")

	// 2. Initial relay attempt (should fail due to non-finalized state)
	err = path.RelayPacket(packet)
	s.Require().Error(err, "expected initial relay to fail")

	// Validate ack is NOT found on the destination chain (Hub)
	found = hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().False(found, "expected packet acknowledgement NOT to be found on Hub before finalization")

	// 3. Finalize the rollapp state
	currentRollappBlockHeight = uint64(s.rollappCtx().BlockHeight()) 
	_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
	s.Require().NoError(err, "rollapp state finalization failed")

	// 4. Manually finalize packets through x/delayedack (This is the final ACK step)
	s.finalizeRollappPacketsByAddress(s.hubChain().SenderAccount.GetAddress().String())

	// Validate ack is found after manual finalization
	found = hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found, "expected packet acknowledgement to be found on Hub after finalization")
}

// TestHubToRollappTimeout tests the scenario where a packet is sent from the hub to the rollapp and times out.
// Funds should only be returned after the rollapp state is finalized.
func (s *delayedAckSuite) TestHubToRollappTimeout() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	
	hubEndpoint := path.EndpointA
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()
	
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()
	
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight()) 
	s.updateRollappState(currentRollappBlockHeight) 
	
	// Timeout height must be based on the destination chain's height.
	timeoutHeight := clienttypes.GetSelfHeight(s.rollappCtx())
	coinToSendToB := s.createTransferCoin(OneDymAmountString)
	
	senderAccount := s.hubChain().SenderAccount.GetAddress()
	receiverAccount := s.rollappChain().SenderAccount.GetAddress()
	
	bankKeeper := s.hubApp().BankKeeper
	preSendBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	
	// 1. Send from hubChain to rollappChain
	msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, senderAccount.String(), receiverAccount.String(), timeoutHeight, disabledTimeoutTimestamp, "")
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet := s.extractPacket(res)
	
	found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found, "packet commitment not found on hub after send")
	
	// Check balance decreased (funds were escrowed)
	postSendBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().Equal(preSendBalance.Amount.Sub(coinToSendToB.Amount).String(), postSendBalance.Amount.String(), "funds were not escrowed correctly")
	
	// 2. Update the client and trigger the timeout
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)
	
	// Timeout the packet. Funds should NOT be released yet due to delayed ack rules.
	err = path.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err)
	
	// Validate funds are still NOT returned to the sender (still escrowed)
	postTimeoutBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().Equal(postSendBalance.Amount.String(), postTimeoutBalance.Amount.String(), "funds should remain escrowed before finalization")
	
	// 3. Finalize the rollapp state (to enable refund)
	currentRollappBlockHeight = uint64(s.rollappCtx().BlockHeight()) 
	_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
	s.Require().NoError(err)
	
	// 4. Manually finalize packets through x/delayedack (Refund step)
	s.finalizeRollappPacketsByAddress(senderAccount.String())
	
	// Validate funds are returned to the sender (Refund successful)
	postFinalizeBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().Equal(preSendBalance.Amount.String(), postFinalizeBalance.Amount.String(), "funds were not returned after finalization")
}

// TestHardFork_HubToRollapp tests the hard fork handling for outgoing packets from the hub to the rollapp.
// We assert the packets commitments are restored and the pending packets are ackable after the hard fork.
func (s *delayedAckSuite) TestHardFork_HubToRollapp() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	// Setup endpoints and constants
	hubEndpoint := path.EndpointA
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()
	senderAccount := s.hubChain().SenderAccount.GetAddress()
	receiverAccount := s.rollappChain().SenderAccount.GetAddress()
	coinToSendToB := s.createTransferCoin(OneDymAmountString)
	timeoutHeight := clienttypes.Height{RevisionNumber: 1, RevisionHeight: 50}

	// Create rollapp and update its initial state
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight())) //nolint:gosec

	// 1. Send from hubChain to rollappChain
	balanceBefore := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, senderAccount.String(), receiverAccount.String(), timeoutHeight, disabledTimeoutTimestamp, "")
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet := s.extractPacket(res)

	// Assert commitments are created
	found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found, "initial packet commitment not found")

	// Update the client
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	// 2. Relay (RecvPacket on Rollapp)
	err = path.RelayPacket(packet)
	// Expecting error here if no acknowledgement is immediately returned (standard rollapp flow)
	s.Require().Error(err, "expected error as no AcknowledgePacket is expected to return immediately") 

	// Progress the rollapp chain
	s.coordinator.CommitNBlocks(s.rollappChain(), 110)

	// Update the client
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	// 3. Write ACK optimistically (simulate pre-hard fork state where ACK was written)
	err = path.EndpointA.AcknowledgePacket(packet, []byte{0x1})
	s.Require().NoError(err)

	// Assert commitments are removed due to the optimistic ACK
	found = hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().False(found, "packet commitment should be removed after optimistic ACK")

	// 4. Timeout attempt (should succeed and refund funds before hard fork)
	err = path.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err) // Can't assert ErrNoOp here, so we check balance refund below
	balanceAfter := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().NotEqual(balanceBefore.String(), balanceAfter.String(), "funds should have been refunded before hard fork due to previous ACK removal")

	// 5. Hard Fork: Restore commitments
	err = s.hubApp().DelayedAckKeeper.OnHardFork(s.hubCtx(), s.rollappCtx().ChainID(), 5)
	s.Require().NoError(err, "OnHardFork failed")

	// Assert commitments are restored after the hard fork
	found = hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found, "packet commitment should be restored after hard fork")

	// 6. Verification: Try to timeout the restored packet
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	// Create timeout message using the corrected helper function
	timeoutMsg := s.getTimeoutPacket(hubEndpoint, packet) 

	// Send timeout message: Expect verification error since the proof points to a pre-hard fork state
	_, err = s.hubChain().SendMsgs(timeoutMsg)
	s.Require().ErrorContains(err, ibcmerkle.ErrInvalidProof.Error(), "expected invalid proof error after hard fork")
}

// getTimeoutPacket is the corrected helper function name.
func (s *delayedAckSuite) getTimeoutPacket(endpoint *ibctesting.Endpoint, packet channeltypes.Packet) *channeltypes.MsgTimeout {
	packetKey := host.PacketReceiptKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	counterparty := endpoint.Counterparty
	
	// Query the proof from the counterparty chain
	proof, proofHeight := counterparty.QueryProof(packetKey)
	
	// Get next sequence receive from the counterparty chain
	nextSeqRecv, found := counterparty.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceRecv(counterparty.Chain.GetContext(), counterparty.ChannelConfig.PortID, counterparty.ChannelID)
	// Use s.Require().True for assertion within the suite method
	s.Require().True(found, "expected next sequence receive to be found")

	timeoutMsg := channeltypes.NewMsgTimeout(
		packet, nextSeqRecv,
		proof, proofHeight, endpoint.Chain.SenderAccount.GetAddress().String(),
	)

	return timeoutMsg
}

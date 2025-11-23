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

// Constants for test parameters
const (
	disabledTimeoutTimestamp = uint64(0)
	defaultTimeoutHeight     = 110
)

var (
	// 10 DYM amount for general transfer tests
	tenDYMAmount = math.NewInt(10_000_000_000_000_000_000)
	// 1 DYM amount for specific balance check tests
	oneDYMAmount = math.NewInt(1_000_000_000_000_000_000)
)

type delayedAckSuite struct {
	ibcTestingSuite
}

func TestDelayedAckTestSuite(t *testing.T) {
	suite.Run(t, new(delayedAckSuite))
}

func (s *delayedAckSuite) SetupTest() {
	s.ibcTestingSuite.SetupTest()
	// Disable light client verification for simplicity in the hub for testing delayed acks
	s.hubApp().LightClientKeeper.SetEnabled(false)

	s.hubApp().BankKeeper.SetDenomMetaData(s.hubCtx(), banktypes.Metadata{
		Base: sdk.DefaultBondDenom,
	})
}

// executeTransferAndVerifyAck is a helper function to perform a transfer, relay the packet, and verify the acknowledgment.
func (s *delayedAckSuite) executeTransferAndVerifyAck(
	path *ibctesting.Path,
	sourcePortID, sourceChannelID string,
	sender, receiver sdk.AccAddress,
	coin sdk.Coin,
	timeoutHeight clienttypes.Height,
	// The chain where the ACK is expected to be found
	ackChain *ibctesting.TestChain,
) {
	msg := types.NewMsgTransfer(
		sourcePortID, sourceChannelID, coin,
		sender.String(), receiver.String(), timeoutHeight, disabledTimeoutTimestamp, "",
	)
	
	// Send the message from the source chain
	res, err := path.EndpointA.Chain.SendMsgs(msg)
	s.Require().NoError(err, "MsgTransfer should be committed successfully")

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err, "Packet must be parsed from events")

	// Relay the packet to the destination chain (which includes processing RecvPacket and AckPacket)
	err = path.RelayPacket(packet)
	s.Require().NoError(err, "RelayPacket should succeed and return ACK immediately for standard chains")

	// Verify the acknowledgment is found on the destination chain
	ibcKeeper := ackChain.App.GetIBCKeeper()
	found := ibcKeeper.ChannelKeeper.HasPacketAcknowledgement(
		ackChain.GetContext(), 
		packet.GetDestPort(), 
		packet.GetDestChannel(), 
		packet.GetSequence(),
	)
	s.Require().True(found, "Packet acknowledgment must be found on the destination chain")
}

// TestTransferCosmosToHub tests standard IBC transfer where ACK is immediate (no delayed ack mechanism).
func (s *delayedAckSuite) TestTransferCosmosToHub() {
	// Setup transfer path between cosmosChain (Source/EndpointB) and hubChain (Destination/EndpointA)
	path := s.newTransferPath(s.hubChain(), s.cosmosChain())
	s.coordinator.Setup(path)

	cosmosEndpoint := path.EndpointB
	timeoutHeight := clienttypes.NewHeight(100, defaultTimeoutHeight)
	coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, tenDYMAmount)

	// Source: cosmosChain (EndpointB), Destination: hubChain (EndpointA)
	s.executeTransferAndVerifyAck(
		path,
		cosmosEndpoint.ChannelConfig.PortID,
		cosmosEndpoint.ChannelID,
		s.cosmosChain().SenderAccount.GetAddress(),
		s.hubChain().SenderAccount.GetAddress(),
		coinToSend,
		timeoutHeight,
		s.hubChain(), // ACK expected on the Hub (destination)
	)
}

// TestTransferHubToCosmos tests standard IBC transfer where ACK is immediate.
func (s *delayedAckSuite) TestTransferHubToCosmos() {
	// Setup transfer path between hubChain (Source/EndpointA) and cosmosChain (Destination/EndpointB)
	path := s.newTransferPath(s.hubChain(), s.cosmosChain())
	s.coordinator.Setup(path)

	hubEndpoint := path.EndpointA
	timeoutHeight := clienttypes.NewHeight(100, defaultTimeoutHeight)
	coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, tenDYMAmount)

	// Source: hubChain (EndpointA), Destination: cosmosChain (EndpointB)
	s.executeTransferAndVerifyAck(
		path,
		hubEndpoint.ChannelConfig.PortID,
		hubEndpoint.ChannelID,
		s.hubChain().SenderAccount.GetAddress(),
		s.cosmosChain().SenderAccount.GetAddress(),
		coinToSend,
		timeoutHeight,
		s.cosmosChain(), // ACK expected on Cosmos (destination)
	)
}

// TestTransferRollappToHubNotFinalized tests transfer from a Rollapp to the Hub before the state is finalized.
// ACK relay is expected to fail.
func (s *delayedAckSuite) TestTransferRollappToHubNotFinalized() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	rollappEndpoint := path.EndpointB
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()

	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()
	// Update Rollapp state to commit the block header to the Hub client
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight()))

	timeoutHeight := clienttypes.NewHeight(100, defaultTimeoutHeight)
	coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, tenDYMAmount)

	msg := types.NewMsgTransfer(
		rollappEndpoint.ChannelConfig.PortID,
		rollappEndpoint.ChannelID,
		coinToSend,
		s.rollappChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		disabledTimeoutTimestamp,
		"",
	)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err, "MsgTransfer should commit successfully on Rollapp")

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// Attempt to relay the packet (RecvPacket on Hub)
	err = path.RelayPacket(packet)
	// Expecting an error because the ACK is 'delayed' and not returned by RecvPacket processing immediately.
	// This usually means the Hub's RecvPacket handler correctly delays the ACK processing.
	s.Require().Error(err, "Relaying packet should result in an error or non-ACK return since ACK is delayed")
	
	// Validate that the ACK is NOT found on the Hub (destination chain)
	found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(
		s.hubCtx(), 
		packet.GetDestPort(), 
		packet.GetDestChannel(), 
		packet.GetSequence(),
	)
	s.Require().False(found, "Packet acknowledgment should be delayed and not found yet")
}

// TestTransferRollappToHubFinalization tests transfer from a Rollapp to the Hub, ensuring the ACK is processed 
// only after the Rollapp state is finalized on the Hub.
func (s *delayedAckSuite) TestTransferRollappToHubFinalization() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()
	rollappEndpoint := path.EndpointB

	// Rollapp setup steps
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()

	// 1. Initial Update of rollapp state
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight())
	s.updateRollappState(currentRollappBlockHeight)

	timeoutHeight := clienttypes.NewHeight(100, defaultTimeoutHeight)
	coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, tenDYMAmount)

	/* --------------------- initiating transfer on rollapp --------------------- */
	msg := types.NewMsgTransfer(
		rollappEndpoint.ChannelConfig.PortID,
		rollappEndpoint.ChannelID,
		coinToSend,
		s.rollappChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		disabledTimeoutTimestamp,
		"",
	)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err, "MsgTransfer should commit successfully on Rollapp")
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// Assert the packet commitment exists on the source chain (Rollapp)
	rollappIBCKeeper := s.rollappChain().App.GetIBCKeeper()
	found := rollappIBCKeeper.ChannelKeeper.HasPacketCommitment(
		s.rollappCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(),
	)
	s.Require().True(found, "Packet commitment must exist on Rollapp")

	// 2. Attempt relay: RecvPacket is successful on Hub, but ACK is delayed/not returned
	err = path.RelayPacket(packet)
	s.Require().Error(err, "Relay should fail or return an error as ACK is delayed")

	// Validate ACK is not found yet on the Hub
	found = hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(
		s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
	)
	s.Require().False(found, "Packet acknowledgment should not be found before finalization")

	// 3. Finalize the rollapp state (This makes the packet on the Hub 'finalized')
	currentRollappBlockHeight = uint64(s.rollappCtx().BlockHeight())
	_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
	s.Require().NoError(err, "Rollapp state finalization must succeed")

	// 4. Manually trigger the delayed ack processing through the dedicated module
	s.finalizeRollappPacketsByAddress(s.hubChain().SenderAccount.GetAddress().String())

	// 5. Validate ack is now found
	found = hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(
		s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
	)
	s.Require().True(found, "Packet acknowledgment must be found after finalization and processing")
}

// TestHubToRollappTimeout tests the scenario where a packet is sent from the hub to the rollapp
// and times out. The packet should only be returned to the sender after the rollapp state is finalized.
func (s *delayedAckSuite) TestHubToRollappTimeout() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	
	hubEndpoint := path.EndpointA
	
	// Create rollapp and update its initial state
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight()))

	// Set the timeout height to be current height to trigger timeout immediately after client update
	timeoutHeight := clienttypes.GetSelfHeight(s.rollappCtx())
	coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, oneDYMAmount)
	senderAccount := s.hubChain().SenderAccount.GetAddress()
	receiverAccount := s.rollappChain().SenderAccount.GetAddress()
	bankKeeper := s.hubApp().BankKeeper
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()

	// Check initial balance
	preSendBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	
	// Send from hubChain to rollappChain
	msg := types.NewMsgTransfer(
		hubEndpoint.ChannelConfig.PortID, 
		hubEndpoint.ChannelID, 
		coinToSend, 
		senderAccount.String(), 
		receiverAccount.String(), 
		timeoutHeight, 
		disabledTimeoutTimestamp, 
		"",
	)
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// Assert packet commitment exists on the source chain (Hub)
	found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(
		s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(),
	)
	s.Require().True(found)

	// Check balance decreased (tokens locked in escrow)
	postSendBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	expectedPostSend := preSendBalance.Amount.Sub(coinToSend.Amount)
	s.Require().Equal(expectedPostSend, postSendBalance.Amount, "Balance should decrease after sending (escrowed)")

	// Update the client on the Hub to exceed the timeout height
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)
	
	// 1. Timeout the packet. Funds should NOT be released until rollapp height is finalized
	err = path.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err, "TimeoutPacket should succeed")
	
	// Validate funds are still NOT returned to the sender (Timeout is pending finalization check)
	postTimeoutBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().Equal(postSendBalance.Amount, postTimeoutBalance.Amount, "Balance should NOT change after Timeout until finalization")

	// 2. Finalize the rollapp state (This allows the Timeout to fully process)
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight())
	_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
	s.Require().NoError(err)
	
	// 3. Manually trigger the delayed ack processing/timeout by the module
	s.finalizeRollappPacketsByAddress(senderAccount.String())

	// 4. Validate funds are returned to the sender (Timeout is completed)
	postFinalizeBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().Equal(preSendBalance.Amount, postFinalizeBalance.Amount, "Balance must be restored after finalization and processing")
}

// TestHardFork_HubToRollapp tests the hard fork handling for outgoing packets from the hub to the rollapp.
// It asserts that packet commitments are restored and pending packets are ackable after the hard fork.
func (s *delayedAckSuite) TestHardFork_HubToRollapp() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)

	// Setup variables
	var (
		hubEndpoint     = path.EndpointA
		hubIBCKeeper    = s.hubChain().App.GetIBCKeeper()
		senderAccount   = s.hubChain().SenderAccount.GetAddress()
		receiverAccount = s.rollappChain().SenderAccount.GetAddress()
		coinToSend      = sdk.NewCoin(sdk.DefaultBondDenom, oneDYMAmount)
		// Set a specific timeout height for the hard fork scenario
		timeoutHeight = clienttypes.Height{RevisionNumber: 1, RevisionHeight: 50}
	)

	// Create rollapp and update its initial state
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)
	s.registerSequencer()
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight()))

	// 1. Send packet from Hub to Rollapp
	balanceBefore := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	msg := types.NewMsgTransfer(
		hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSend, 
		senderAccount.String(), receiverAccount.String(), timeoutHeight, disabledTimeoutTimestamp, "",
	)
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// Assert commitments are created on the source chain (Hub)
	found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(
		s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(),
	)
	s.Require().True(found, "Packet commitment must be created after MsgTransfer")

	// 2. Update the client on the Hub
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	// 3. RecvPacket on Rollapp (This will succeed, but ACK is delayed/handled by delayedack module)
	err = path.RelayPacket(packet)
	s.Require().NoError(err, "RelayPacket should succeed up to RecvPacket on the Rollapp")

	// 4. Progress the Rollapp chain, potentially creating the ACK proof
	s.coordinator.CommitNBlocks(s.rollappChain(), 110)

	// 5. Update the client on the Hub again
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	// 6. Write ACK optimistically (simulating the delayed ack module writing the ACK on the Hub)
	// This removes the original packet commitment.
	err = path.EndpointA.AcknowledgePacket(packet, []byte{0x1})
	s.Require().NoError(err)

	// Assert commitments are NO LONGER available (as ACK was written)
	found = hubIBCKeeper.ChannelKeeper.HasPacketCommitment(
		s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(),
	)
	s.Require().False(found, "Packet commitment should be removed after AcknowledgePacket")

	// 7. Timeout the packet. It should NOT release funds as the ACK was already written (No-Op)
	err = path.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err)
	balanceAfter := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
	s.Require().NotEqual(balanceBefore.String(), balanceAfter.String(), "Balance should have decreased post-send and not be refunded by No-Op Timeout")

	// 8. Perform Hard Fork (reverts the state/restores commitments)
	err = s.hubApp().DelayedAckKeeper.OnHardFork(s.hubCtx(), s.rollappCtx().ChainID(), 5)
	s.Require().NoError(err, "OnHardFork must succeed")

	// Assert commitments are CREATED AGAIN (restored by the hard fork logic)
	found = hubIBCKeeper.ChannelKeeper.HasPacketCommitment(
		s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(),
	)
	s.Require().True(found, "Packet commitment must be restored after OnHardFork")

	// 9. Update the client
	err = hubEndpoint.UpdateClient()
	s.Require().NoError(err)

	// 10. Attempt to Timeout the packet again. This must fail because the previous ACK proof is no longer valid
	// after the hard fork (state/proof root changed) and the packet is considered active again.
	timeoutMsg := getTimeoutPacket(hubEndpoint, packet)

	_, err = s.hubChain().SendMsgs(timeoutMsg)
	// We expect a verification error (invalid proof) because the proof from the older, pre-hard-fork state is invalid.
	s.Require().ErrorContains(err, ibcmerkle.ErrInvalidProof.Error(), "Timeout must fail due to invalid proof after hard fork")
}

// getTimeoutPacket constructs a MsgTimeout based on the current state of the counterparty chain.
func getTimeoutPacket(endpoint *ibctesting.Endpoint, packet channeltypes.Packet) *channeltypes.MsgTimeout {
	// The key to query for the absence of the packet receipt on the counterparty chain
	packetKey := host.PacketReceiptKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	counterparty := endpoint.Counterparty
	
	// Query proof of absence (or receipt if it was received) on the counterparty chain
	proof, proofHeight := counterparty.QueryProof(packetKey)
	
	// Get the next sequence expected to be received (used for the proof height)
	nextSeqRecv, found := counterparty.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceRecv(
		counterparty.Chain.GetContext(), counterparty.ChannelConfig.PortID, counterparty.ChannelID,
	)
	require.True(endpoint.Chain.TB, found)

	timeoutMsg := channeltypes.NewMsgTimeout(
		packet, 
		nextSeqRecv,
		proof, 
		proofHeight, 
		endpoint.Chain.SenderAccount.GetAddress().String(),
	)

	return timeoutMsg
}

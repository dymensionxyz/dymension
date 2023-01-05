package irctesting

import (
	"testing"

	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/packet-forward-middleware/v3/router/keeper"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"encoding/json"
)

// TestConnection create connection between dymint rollapp and dymension hub chain
func (suite *IrcTestSuite) SetupConnection(rollappChain *ibctesting.TestChain, hubChain *ibctesting.TestChain) *ibctesting.Path {
	// create chains, chainA is dymint and chainB is dymension hub
	path := ibctesting.NewPath(rollappChain, hubChain)
	// dymension hub chain
	dymensionEndpoint := path.EndpointB
	// rollapp dymint chain
	rollappEndpoint := path.EndpointA
	// create rollapp
	err := suite.CreateRollapp(rollappEndpoint.Chain.ChainID)
	suite.Require().Nil(err)
	// create sequencer
	err = suite.CreateSequencer(rollappEndpoint.Chain.ChainID)
	suite.Require().Nil(err)

	//--
	// Create Clients
	//--

	bds := &rollappChain.TestChainClient.(*DymintTestChainClient).bds
	// update rollapp state
	err = suite.UpdateRollappState(rollappEndpoint.Chain.ChainID, bds)
	// create the client of dymension on the rollapp chain
	err = rollappEndpoint.CreateClient()
	suite.Require().Nil(err)
	// update rollapp state
	err = suite.UpdateRollappState(rollappEndpoint.Chain.ChainID, bds)
	// create the client of the rollapp on dymension
	err = dymensionEndpoint.CreateClient()
	suite.Require().Nil(err)

	//--
	// Create Connection
	//--

	// init a connection on the rollapp
	err = rollappEndpoint.ConnOpenInit()
	suite.Require().Nil(err)
	// update rollapp state
	err = suite.UpdateRollappState(rollappEndpoint.Chain.ChainID, bds)
	// try to open connection on the hub
	err = dymensionEndpoint.ConnOpenTry()
	suite.Require().Nil(err)
	// send ack to to rollapp
	err = rollappEndpoint.ConnOpenAck()
	suite.Require().Nil(err)
	// update rollapp state
	err = suite.UpdateRollappState(rollappEndpoint.Chain.ChainID, bds)
	// send confirmation to the hub
	err = dymensionEndpoint.ConnOpenConfirm()
	suite.Require().Nil(err)
	// ensure rollapp is up to date
	err = rollappEndpoint.UpdateClient()
	suite.Require().Nil(err)

	//--
	// Create Channel
	//--

	// set configuration for transfer channel
	rollappEndpoint.ChannelConfig.PortID = ibctransfertypes.PortID
	rollappEndpoint.ChannelConfig.Version = ibctransfertypes.Version
	dymensionEndpoint.ChannelConfig.PortID = ibctransfertypes.PortID
	dymensionEndpoint.ChannelConfig.Version = ibctransfertypes.Version
	// init a channel on the rollapp
	err = rollappEndpoint.ChanOpenInit()
	suite.Require().Nil(err)
	// update rollapp state
	err = suite.UpdateRollappState(rollappEndpoint.Chain.ChainID, bds)
	// try to open channel on the hub
	err = dymensionEndpoint.ChanOpenTry()
	suite.Require().Nil(err)
	// send ack to to rollapp
	err = rollappEndpoint.ChanOpenAck()
	suite.Require().Nil(err)
	// update rollapp state
	err = suite.UpdateRollappState(rollappEndpoint.Chain.ChainID, bds)
	// send confirmation to the hub
	err = dymensionEndpoint.ChanOpenConfirm()
	suite.Require().Nil(err)
	// ensure counterparty is up to date
	err = rollappEndpoint.UpdateClient()
	suite.Require().Nil(err)

	return path
}

// constructs a send from chainA to chainB on the established channel/connection
// and sends the same coin back from chainB to chainA.
func (suite *IrcTestSuite) TestHandleMsgMiddlewareTransfer() {
	// setup connection between chainA and chainB
	pathAtoB := suite.SetupConnection(suite.chainA, suite.chainB)

	// setup connection between chain to chainB
	// NOTE:
	// pathBtoC.EndpointA = endpoint on chainC
	// pathBtoC.EndpointB = endpoint on chainB
	pathCtoB := suite.SetupConnection(suite.chainC, suite.chainB)

	// rollapp1 dymint chain
	rollapp1Endpoint := pathAtoB.EndpointA
	// rollapp2 dymint chain
	rollapp2Endpoint := pathCtoB.EndpointA

	// block descriptors for update
	rollapp1bds := &suite.chainA.TestChainClient.(*DymintTestChainClient).bds
	rollapp2bds := &suite.chainC.TestChainClient.(*DymintTestChainClient).bds

	amount, ok := sdk.NewIntFromString("1000")
	suite.Require().True(ok)

	//************************************************************************
	//				send from chainA to chainC through B
	//************************************************************************

	// build forward message
	coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	msg := ibctransfertypes.NewMsgTransfer(pathAtoB.EndpointA.ChannelConfig.PortID,
		pathAtoB.EndpointA.ChannelID,
		coinToSend,
		suite.chainA.SenderAccount.GetAddress().String(),
		suite.chainB.SenderAccount.GetAddress().String(),
		timeoutHeight, 0)
	nextMetadata := &keeper.PacketMetadata{
		Forward: &keeper.ForwardMetadata{
			Receiver: suite.chainC.SenderAccount.GetAddress().String(),
			Port:     pathCtoB.EndpointB.ChannelConfig.PortID,
			Channel:  pathCtoB.EndpointB.ChannelID,
		},
	}
	memo, err := json.Marshal(nextMetadata)
	suite.Require().NoError(err)
	msg.Memo = string(memo)

	res, err := suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// update rollapp1 state before updating client
	suite.UpdateRollappState(rollapp1Endpoint.Chain.ChainID, rollapp1bds)
	// relay send from A to B
	err = pathAtoB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	res, err = RecvPacketWithResult(pathAtoB.EndpointB, packet)
	suite.Require().NoError(err)
	events := res.GetEvents()

	// parse the packet of the forward middleware
	forwardPacket, err := ibctesting.ParsePacketFromEvents(events)
	suite.Require().NoError(err)

	// check the balance of the module account escrow address on chain A (holding the funds)
	escrowAddressChainA := ibctransfertypes.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
	balance := GetRollappSimApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), escrowAddressChainA, sdk.DefaultBondDenom)
	suite.Require().Equal(sdk.NewCoin(sdk.DefaultBondDenom, amount), balance)

	// check the balance of the account escrow on chain B (holding the new minted funds on chain B)
	fullDenomPathCahinB := ibctransfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom)
	coinSentFromAToB := sdk.NewCoin(ibctransfertypes.ParseDenomTrace(fullDenomPathCahinB).IBCDenom(), amount)
	escrowAddressChainB := ibctransfertypes.GetEscrowAddress(forwardPacket.GetSourcePort(), forwardPacket.GetSourceChannel())
	balance = GetHubSimApp(suite.chainB).BankKeeper.GetBalance(suite.chainB.GetContext(), escrowAddressChainB, coinSentFromAToB.Denom)
	suite.Require().Equal(coinSentFromAToB, balance)
	print("\n", fullDenomPathCahinB)

	// relay forward middleware packet from B to C
	err = pathCtoB.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	res, err = RecvPacketWithResult(pathCtoB.EndpointA, forwardPacket)
	suite.Require().NoError(err)

	// parse the ack from C
	events = res.GetEvents()
	ack, err := ibctesting.ParseAckFromEvents(events)
	suite.Require().NoError(err)

	// check that the balance is updated on chainC (having the new minted funds on chain C)
	fullDenomPathCahinC := ibctransfertypes.GetPrefixedDenom(forwardPacket.GetDestPort(), forwardPacket.GetDestChannel(), fullDenomPathCahinB)
	coinSentFromBToC := sdk.NewCoin(ibctransfertypes.ParseDenomTrace(fullDenomPathCahinC).IBCDenom(), amount)
	balance = GetRollappSimApp(suite.chainC).BankKeeper.GetBalance(suite.chainC.GetContext(), suite.chainC.SenderAccount.GetAddress(), coinSentFromBToC.Denom)
	suite.Require().Equal(coinSentFromBToC, balance)
	print("\n", fullDenomPathCahinC)

	// update rollapp1 state before updating client
	suite.UpdateRollappState(rollapp2Endpoint.Chain.ChainID, rollapp2bds)
	// relay the ack to C
	err = pathCtoB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	res, err = AcknowledgePacket(pathCtoB.EndpointB, forwardPacket, ack)
	suite.Require().NoError(err)

	// the forward middleware ack also generates the ack of the source send (from A to B)
	events = res.GetEvents()
	ack, err = ibctesting.ParseAckFromEvents(events)
	suite.Require().NoError(err)

	// relay Acknowledgement to chain A
	err = pathAtoB.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	_, err = AcknowledgePacket(pathAtoB.EndpointA, packet, ack)
	suite.Require().NoError(err)

	//************************************************************************
	//				send back from chainC to chainA through B
	//************************************************************************

	// build forward message
	coinToSend = sdk.NewCoin(coinSentFromBToC.Denom, amount)
	msg = ibctransfertypes.NewMsgTransfer(pathCtoB.EndpointA.ChannelConfig.PortID,
		pathCtoB.EndpointA.ChannelID,
		coinToSend,
		suite.chainC.SenderAccount.GetAddress().String(),
		suite.chainB.SenderAccount.GetAddress().String(),
		timeoutHeight, 0)
	nextMetadata = &keeper.PacketMetadata{
		Forward: &keeper.ForwardMetadata{
			Receiver: suite.chainA.SenderAccount.GetAddress().String(),
			Port:     pathAtoB.EndpointB.ChannelConfig.PortID,
			Channel:  pathAtoB.EndpointB.ChannelID,
		},
	}
	memo, err = json.Marshal(nextMetadata)
	suite.Require().NoError(err)
	msg.Memo = string(memo)

	res, err = suite.chainC.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// update rollapp2 state before updating client
	suite.UpdateRollappState(rollapp2Endpoint.Chain.ChainID, rollapp2bds)
	// relay send from C to B
	err = pathCtoB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	res, err = RecvPacketWithResult(pathCtoB.EndpointB, packet)
	suite.Require().NoError(err)
	events = res.GetEvents()

	// parse the packet of the forward middleware
	forwardPacket, err = ibctesting.ParsePacketFromEvents(events)
	suite.Require().NoError(err)

	// check that the balance is updated on chainC (burned to zero)
	balance = GetRollappSimApp(suite.chainC).BankKeeper.GetBalance(suite.chainC.GetContext(), suite.chainC.SenderAccount.GetAddress(), coinSentFromBToC.Denom)
	suite.Require().Zero(balance.Amount.Int64())

	// check the balance of the account escrow on chain B (not holding burned tokens)
	balance = GetHubSimApp(suite.chainB).BankKeeper.GetBalance(suite.chainB.GetContext(), escrowAddressChainB, coinSentFromAToB.Denom)
	suite.Require().Zero(balance.Amount.Int64())

	// relay forward middleware packet from B to A
	err = pathAtoB.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	res, err = RecvPacketWithResult(pathAtoB.EndpointA, forwardPacket)
	suite.Require().NoError(err)

	// parse the ack from A
	events = res.GetEvents()
	ack, err = ibctesting.ParseAckFromEvents(events)
	suite.Require().NoError(err)

	// check that the balance is updated on chainA (having the new minted funds on chain C)
	balance = GetRollappSimApp(suite.chainC).BankKeeper.GetBalance(suite.chainC.GetContext(), suite.chainC.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	//suite.Require().Equal(coinSentFromBToC, balance)

	// check the balance of the module account escrow address on chain A (not holding tokens)
	balance = GetRollappSimApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), escrowAddressChainA, sdk.DefaultBondDenom)
	suite.Require().Zero(balance.Amount.Int64())

	// update rollapp1 state before updating client
	suite.UpdateRollappState(rollapp1Endpoint.Chain.ChainID, rollapp1bds)
	// relay the ack to B (from A)
	err = pathAtoB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	res, err = AcknowledgePacket(pathAtoB.EndpointB, forwardPacket, ack)
	suite.Require().NoError(err)

	// the forward middleware ack also generates the ack of the source send (from C to B)
	events = res.GetEvents()
	ack, err = ibctesting.ParseAckFromEvents(events)
	suite.Require().NoError(err)

	// relay Acknowledgement to chain C
	err = pathCtoB.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	_, err = AcknowledgePacket(pathCtoB.EndpointA, packet, ack)
	suite.Require().NoError(err)
}

// AcknowledgePacket sends a MsgAcknowledgement to the channel associated with the endpoint.
func AcknowledgePacket(endpoint *ibctesting.Endpoint, packet channeltypes.Packet, ack []byte) (*sdk.Result, error) {
	// get proof of acknowledgement on counterparty
	packetKey := host.PacketAcknowledgementKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	proof, proofHeight := endpoint.Counterparty.QueryProof(packetKey)

	ackMsg := channeltypes.NewMsgAcknowledgement(packet, ack, proof, proofHeight, endpoint.Chain.SenderAccount.GetAddress().String())

	return endpoint.Chain.SendMsgs(ackMsg)
}

// RecvPacketWithResult receives a packet on the associated endpoint and the result
// of the transaction is returned.
func RecvPacketWithResult(endpoint *ibctesting.Endpoint, packet channeltypes.Packet) (*sdk.Result, error) {
	// get proof of packet commitment on source
	packetKey := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	proof, proofHeight := endpoint.Counterparty.Chain.QueryProof(packetKey)

	recvMsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, endpoint.Chain.SenderAccount.GetAddress().String())

	// receive on counterparty and update source client
	return endpoint.Chain.SendMsgs(recvMsg)
}

// TestSuite initialize dymint rollapp and dymension hub chain
func TestSuite(t *testing.T) {
	suite.Run(t, &IrcTestSuite{
		chainAConsensusType: exported.Dymint,     // rollapp dymint chain
		chainBConsensusType: exported.Tendermint, // dymension hub chain
		chainCConsensusType: exported.Dymint,     // rollapp dymint chain
	})
}

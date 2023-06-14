package irctesting

import (
	"testing"

	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	routertypes "github.com/strangelove-ventures/packet-forward-middleware/v3/router/types"

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

// TestHandleMsgMiddlewareTransfer constructs a transfer from rollapp1 to rollapp2 through the hub
// by using the packet-forward-middleware  and sends the same coin back from rollapp2 to rollapp1.
func (suite *IrcTestSuite) TestHandleMsgMiddlewareTransfer() {
	// setup connection between rollapp1 and hubChain
	pathRollapp1toHub := suite.SetupConnection(suite.rollapp1, suite.hubChain)

	// setup connection between rollapp2 to hubChain
	pathRollapp2toHub := suite.SetupConnection(suite.rollapp2, suite.hubChain)

	// rollapp1 dymint chain
	rollapp1Endpoint := pathRollapp1toHub.EndpointA
	// rollapp2 dymint chain
	rollapp2Endpoint := pathRollapp2toHub.EndpointA

	// block descriptors for update
	rollapp1bds := &suite.rollapp1.TestChainClient.(*DymintTestChainClient).bds
	rollapp2bds := &suite.rollapp2.TestChainClient.(*DymintTestChainClient).bds

	// This value used to cause panic in the forward middleware and fail the test
	amount, ok := sdk.NewIntFromString("9900000000000000000")
	suite.Require().True(ok)

	//************************************************************************
	//				send from rollapp1 to rollapp2 through the hub
	//************************************************************************

	// build forward message
	coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	msg := ibctransfertypes.NewMsgTransfer(pathRollapp1toHub.EndpointA.ChannelConfig.PortID,
		pathRollapp1toHub.EndpointA.ChannelID,
		coinToSend,
		suite.rollapp1.SenderAccount.GetAddress().String(),
		suite.hubChain.SenderAccount.GetAddress().String(),
		timeoutHeight, 0)
	nextMetadata := &routertypes.PacketMetadata{
		Forward: &routertypes.ForwardMetadata{
			Receiver: suite.rollapp2.SenderAccount.GetAddress().String(),
			Port:     pathRollapp2toHub.EndpointB.ChannelConfig.PortID,
			Channel:  pathRollapp2toHub.EndpointB.ChannelID,
		},
	}
	memo, err := json.Marshal(nextMetadata)
	suite.Require().NoError(err)
	msg.Memo = string(memo)

	res, err := suite.rollapp1.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// update rollapp1 state before updating client
	suite.UpdateRollappState(rollapp1Endpoint.Chain.ChainID, rollapp1bds)
	// relay send from rollapp1 to the hub
	err = pathRollapp1toHub.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	res, err = RecvPacketWithResult(pathRollapp1toHub.EndpointB, packet)
	suite.Require().NoError(err)
	events := res.GetEvents()

	// parse the packet of the forward middleware
	forwardPacket, err := ibctesting.ParsePacketFromEvents(events)
	suite.Require().NoError(err)

	// check the balance of the module account escrow address on rollapp1 (holding the funds)
	escrowAddressRollapp1 := ibctransfertypes.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
	balance := GetRollappSimApp(suite.rollapp1).BankKeeper.GetBalance(suite.rollapp1.GetContext(), escrowAddressRollapp1, sdk.DefaultBondDenom)
	suite.Require().Equal(sdk.NewCoin(sdk.DefaultBondDenom, amount), balance)

	// check the balance of the account escrow on the hub (holding the new minted funds on the hub)
	fullDenomPathCahinB := ibctransfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom)
	coinSentFromAToB := sdk.NewCoin(ibctransfertypes.ParseDenomTrace(fullDenomPathCahinB).IBCDenom(), amount)
	hubEscrowAddress := ibctransfertypes.GetEscrowAddress(forwardPacket.GetSourcePort(), forwardPacket.GetSourceChannel())
	balance = GetHubSimApp(suite.hubChain).BankKeeper.GetBalance(suite.hubChain.GetContext(), hubEscrowAddress, coinSentFromAToB.Denom)
	suite.Require().Equal(coinSentFromAToB, balance)
	print("\n", fullDenomPathCahinB)

	// relay forward middleware packet from the hub to rollapp2
	err = pathRollapp2toHub.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	res, err = RecvPacketWithResult(pathRollapp2toHub.EndpointA, forwardPacket)
	suite.Require().NoError(err)

	// parse the ack from rollapp2
	events = res.GetEvents()
	ack, err := ibctesting.ParseAckFromEvents(events)
	suite.Require().NoError(err)

	// check that the balance is updated on rollapp2 (having the new minted funds on chain rollapp2)
	fullDenomPathCahinC := ibctransfertypes.GetPrefixedDenom(forwardPacket.GetDestPort(), forwardPacket.GetDestChannel(), fullDenomPathCahinB)
	coinSentFromBToC := sdk.NewCoin(ibctransfertypes.ParseDenomTrace(fullDenomPathCahinC).IBCDenom(), amount)
	balance = GetRollappSimApp(suite.rollapp2).BankKeeper.GetBalance(suite.rollapp2.GetContext(), suite.rollapp2.SenderAccount.GetAddress(), coinSentFromBToC.Denom)
	suite.Require().Equal(coinSentFromBToC, balance)
	print("\n", fullDenomPathCahinC)

	// update rollapp1 state before updating client
	suite.UpdateRollappState(rollapp2Endpoint.Chain.ChainID, rollapp2bds)
	// relay the ack to rollapp2
	err = pathRollapp2toHub.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	res, err = AcknowledgePacket(pathRollapp2toHub.EndpointB, forwardPacket, ack)
	suite.Require().NoError(err)

	// the forward middleware ack also generates the ack of the original send
	events = res.GetEvents()
	ack, err = ibctesting.ParseAckFromEvents(events)
	suite.Require().NoError(err)

	// relay Acknowledgement to rollapp1
	err = pathRollapp1toHub.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	_, err = AcknowledgePacket(pathRollapp1toHub.EndpointA, packet, ack)
	suite.Require().NoError(err)

	//************************************************************************
	//				send back from rollapp2 to rollapp1 through the hub
	//************************************************************************

	// build forward message
	coinToSend = sdk.NewCoin(coinSentFromBToC.Denom, amount)
	msg = ibctransfertypes.NewMsgTransfer(pathRollapp2toHub.EndpointA.ChannelConfig.PortID,
		pathRollapp2toHub.EndpointA.ChannelID,
		coinToSend,
		suite.rollapp2.SenderAccount.GetAddress().String(),
		suite.hubChain.SenderAccount.GetAddress().String(),
		timeoutHeight, 0)
	nextMetadata = &routertypes.PacketMetadata{
		Forward: &routertypes.ForwardMetadata{
			Receiver: suite.rollapp1.SenderAccount.GetAddress().String(),
			Port:     pathRollapp1toHub.EndpointB.ChannelConfig.PortID,
			Channel:  pathRollapp1toHub.EndpointB.ChannelID,
		},
	}
	memo, err = json.Marshal(nextMetadata)
	suite.Require().NoError(err)
	msg.Memo = string(memo)

	res, err = suite.rollapp2.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// update rollapp2 state before updating client
	suite.UpdateRollappState(rollapp2Endpoint.Chain.ChainID, rollapp2bds)
	// relay send from rollapp2 to the hub
	err = pathRollapp2toHub.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	res, err = RecvPacketWithResult(pathRollapp2toHub.EndpointB, packet)
	suite.Require().NoError(err)
	events = res.GetEvents()

	// parse the packet of the forward middleware
	forwardPacket, err = ibctesting.ParsePacketFromEvents(events)
	suite.Require().NoError(err)

	// check that the balance is updated on rollapp2 (burned to zero)
	balance = GetRollappSimApp(suite.rollapp2).BankKeeper.GetBalance(suite.rollapp2.GetContext(), suite.rollapp2.SenderAccount.GetAddress(), coinSentFromBToC.Denom)
	suite.Require().Zero(balance.Amount.Int64())

	// check the balance of the account escrow on the hub (not holding burned tokens)
	balance = GetHubSimApp(suite.hubChain).BankKeeper.GetBalance(suite.hubChain.GetContext(), hubEscrowAddress, coinSentFromAToB.Denom)
	suite.Require().Zero(balance.Amount.Int64())

	// relay forward middleware packet from the hub to rollapp1
	err = pathRollapp1toHub.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	res, err = RecvPacketWithResult(pathRollapp1toHub.EndpointA, forwardPacket)
	suite.Require().NoError(err)

	// parse the ack from rollapp1
	events = res.GetEvents()
	ack, err = ibctesting.ParseAckFromEvents(events)
	suite.Require().NoError(err)

	// check that the balance is updated on rollapp1 (having the new minted funds on chain rollapp2)
	balance = GetRollappSimApp(suite.rollapp2).BankKeeper.GetBalance(suite.rollapp2.GetContext(), suite.rollapp2.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	//suite.Require().Equal(coinSentFromBToC, balance)

	// check the balance of the module account escrow address on rollapp1 (not holding tokens)
	balance = GetRollappSimApp(suite.rollapp1).BankKeeper.GetBalance(suite.rollapp1.GetContext(), escrowAddressRollapp1, sdk.DefaultBondDenom)
	suite.Require().Zero(balance.Amount.Int64())

	// update rollapp1 state before updating client
	suite.UpdateRollappState(rollapp1Endpoint.Chain.ChainID, rollapp1bds)
	// relay the ack to the hub (from rollapp1)
	err = pathRollapp1toHub.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	res, err = AcknowledgePacket(pathRollapp1toHub.EndpointB, forwardPacket, ack)
	suite.Require().NoError(err)

	// the forward middleware ack also generates the ack of the source send (from rollapp2 to the hub)
	events = res.GetEvents()
	ack, err = ibctesting.ParseAckFromEvents(events)
	suite.Require().NoError(err)

	// relay Acknowledgement to rollapp2
	err = pathRollapp2toHub.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	_, err = AcknowledgePacket(pathRollapp2toHub.EndpointA, packet, ack)
	suite.Require().NoError(err)
}

// TestSuite initialize dymint rollapp and dymension hub chain
func TestSuite(t *testing.T) {
	suite.Run(t, &IrcTestSuite{})
}

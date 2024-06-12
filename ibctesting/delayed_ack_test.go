package ibctesting_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

const (
	disabledTimeoutTimestamp = uint64(0)
)

type DelayedAckTestSuite struct {
	IBCTestUtilSuite
}

func TestDelayedAckTestSuite(t *testing.T) {
	suite.Run(t, new(DelayedAckTestSuite))
}

func (suite *DelayedAckTestSuite) SetupTest() {
	suite.IBCTestUtilSuite.SetupTest()
	ConvertToApp(suite.hubChain).BankKeeper.SetDenomMetaData(suite.hubChain.GetContext(), banktypes.Metadata{
		Base: sdk.DefaultBondDenom,
	})
}

// Transfer from cosmos chain to the hub. No delay expected
func (suite *DelayedAckTestSuite) TestTransferCosmosToHub() {
	// setup between cosmosChain and hubChain
	path := suite.NewTransferPath(suite.hubChain, suite.cosmosChain)
	suite.coordinator.Setup(path)

	hubEndpoint := path.EndpointA
	cosmosEndpoint := path.EndpointB
	hubIBCKeeper := suite.hubChain.App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	// send from cosmosChain to hubChain
	msg := types.NewMsgTransfer(cosmosEndpoint.ChannelConfig.PortID, cosmosEndpoint.ChannelID, coinToSendToB, cosmosEndpoint.Chain.SenderAccount.GetAddress().String(), hubEndpoint.Chain.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := cosmosEndpoint.Chain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(hubEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	suite.Require().True(found)
}

func (suite *DelayedAckTestSuite) TestTransferHubToCosmos() {
	// setup between cosmosChain and hubChain
	path := suite.NewTransferPath(suite.hubChain, suite.cosmosChain)
	suite.coordinator.Setup(path)

	hubEndpoint := path.EndpointA
	cosmosEndpoint := path.EndpointB
	cosmosIBCKeeper := suite.cosmosChain.App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	// send from cosmosChain to hubChain
	msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, hubEndpoint.Chain.SenderAccount.GetAddress().String(), cosmosEndpoint.Chain.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := hubEndpoint.Chain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	found := cosmosIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(cosmosEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	suite.Require().True(found)
}

func (suite *DelayedAckTestSuite) TestTransferRollappToHubNotFinalized() {
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)

	hubEndpoint := path.EndpointA
	rollappEndpoint := path.EndpointB
	hubIBCKeeper := suite.hubChain.App.GetIBCKeeper()

	suite.CreateRollapp()
	suite.RegisterSequencer()
	suite.GenesisEvent(path.EndpointA.ChannelID)
	suite.UpdateRollappState(uint64(suite.rollappChain.GetContext().BlockHeight()))

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	msg := types.NewMsgTransfer(
		rollappEndpoint.ChannelConfig.PortID,
		rollappEndpoint.ChannelID,
		coinToSendToB,
		suite.rollappChain.SenderAccount.GetAddress().String(),
		suite.hubChain.SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		"",
	)
	res, err := suite.rollappChain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	// expecting error as no AcknowledgePacket expected
	suite.Require().Error(err) // relay committed
	found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(hubEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	suite.Require().False(found)
}

func (suite *DelayedAckTestSuite) TestTransferRollappToHubFinalization() {
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)

	hubEndpoint := path.EndpointA
	hubIBCKeeper := suite.hubChain.App.GetIBCKeeper()

	rollappEndpoint := path.EndpointB
	rollappIBCKeeper := suite.rollappChain.App.GetIBCKeeper()

	suite.CreateRollapp()
	suite.RegisterSequencer()
	suite.GenesisEvent(path.EndpointA.ChannelID)

	// Upate rollapp state
	currentRollappBlockHeight := uint64(suite.rollappChain.GetContext().BlockHeight())
	suite.UpdateRollappState(currentRollappBlockHeight)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	/* --------------------- initiating transfer on rollapp --------------------- */
	msg := types.NewMsgTransfer(rollappEndpoint.ChannelConfig.PortID, rollappEndpoint.ChannelID, coinToSendToB, suite.rollappChain.SenderAccount.GetAddress().String(), suite.hubChain.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := suite.rollappChain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	found := rollappIBCKeeper.ChannelKeeper.HasPacketCommitment(rollappEndpoint.Chain.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	suite.Require().True(found)

	// relay send
	err = path.RelayPacket(packet)
	// expecting error as no AcknowledgePacket expected to return
	suite.Require().Error(err) // relay committed

	found = hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(hubEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	suite.Require().False(found)

	// Finalize the rollapp state
	currentRollappBlockHeight = uint64(suite.rollappChain.GetContext().BlockHeight())
	_, err = suite.FinalizeRollappState(1, currentRollappBlockHeight)
	suite.Require().NoError(err)

	// Validate ack is found
	found = hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(hubEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	suite.Require().True(found)
}

// TestHubToRollappTimeout tests the scenario where a packet is sent from the hub to the rollapp and the rollapp times out the packet.
// The packet should actually get timed out and funds returned to the user only after the rollapp state is finalized.
func (suite *DelayedAckTestSuite) TestHubToRollappTimeout() {
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)
	// Setup endpoints
	hubEndpoint := path.EndpointA
	rollappEndpoint := path.EndpointB
	hubIBCKeeper := suite.hubChain.App.GetIBCKeeper()
	// Create rollapp and update its initial state
	suite.CreateRollapp()
	suite.RegisterSequencer()
	suite.GenesisEvent(path.EndpointA.ChannelID)
	suite.UpdateRollappState(uint64(suite.rollappChain.GetContext().BlockHeight()))
	// Set the timeout height
	timeoutHeight := clienttypes.GetSelfHeight(suite.rollappChain.GetContext())
	amount, ok := sdk.NewIntFromString("1000000000000000000") // 1DYM
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	// Setup accounts
	senderAccount := hubEndpoint.Chain.SenderAccount.GetAddress()
	receiverAccount := rollappEndpoint.Chain.SenderAccount.GetAddress()
	// Check balances
	bankKeeper := ConvertToApp(suite.hubChain).BankKeeper
	preSendBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), senderAccount, sdk.DefaultBondDenom)
	// send from hubChain to rollappChain
	msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, senderAccount.String(), receiverAccount.String(), timeoutHeight, disabledTimeoutTimestamp, "")
	res, err := hubEndpoint.Chain.SendMsgs(msg)
	suite.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(hubEndpoint.Chain.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	suite.Require().True(found)
	// Check balance decreased
	postSendBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), senderAccount, sdk.DefaultBondDenom)
	suite.Require().Equal(preSendBalance.Amount.Sub(coinToSendToB.Amount), postSendBalance.Amount)
	// Update the client to create timeout
	err = hubEndpoint.UpdateClient()
	suite.Require().NoError(err)
	// Timeout the packet. Shouldn't release funds until rollapp height is finalized
	err = path.EndpointA.TimeoutPacket(packet)
	suite.Require().NoError(err)
	// Validate funds are still not returned to the sender
	postTimeoutBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), senderAccount, sdk.DefaultBondDenom)
	suite.Require().Equal(postSendBalance.Amount, postTimeoutBalance.Amount)
	// Finalize the rollapp state
	currentRollappBlockHeight := uint64(suite.rollappChain.GetContext().BlockHeight())
	_, err = suite.FinalizeRollappState(1, currentRollappBlockHeight)
	suite.Require().NoError(err)
	// Validate funds are returned to the sender
	postFinalizeBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), senderAccount, sdk.DefaultBondDenom)
	suite.Require().Equal(preSendBalance.Amount, postFinalizeBalance.Amount)
}

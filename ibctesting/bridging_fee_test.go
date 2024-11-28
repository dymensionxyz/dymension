package ibctesting_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/osmosis-labs/osmosis/v15/x/txfees"
	"github.com/stretchr/testify/suite"
)

type bridgingFeeSuite struct {
	utilSuite
}

func TestBridgingFeeTestSuite(t *testing.T) {
	suite.Run(t, new(bridgingFeeSuite))
}

func (s *bridgingFeeSuite) SetupTest() {
	s.utilSuite.SetupTest()
	s.hubApp().LightClientKeeper.SetEnabled(false)
}

func (s *bridgingFeeSuite) TestNotRollappNoBridgingFee() {
	// setup between cosmosChain and hubChain
	path := s.newTransferPath(s.hubChain(), s.cosmosChain())
	s.coordinator.Setup(path)
	cosmosEndpoint := path.EndpointB

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
	err = path.RelayPacket(packet)
	s.NoError(err) // relay committed

	denom := s.getRollappToHubIBCDenomFromPacket(packet)
	finalBalance := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), denom)
	s.Equal(sdk.NewCoin(denom, coinToSendToB.Amount), finalBalance)
}

func (s *bridgingFeeSuite) TestBridgingFee() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.registerSequencer()
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)

	rollappEndpoint := path.EndpointB
	rollappIBCKeeper := s.rollappChain().App.GetIBCKeeper()

	// Update rollapp state
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight())
	s.updateRollappState(currentRollappBlockHeight)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	s.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

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
	s.Require().NoError(err) // message committed
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	found := rollappIBCKeeper.ChannelKeeper.HasPacketCommitment(s.rollappCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found)
	err = path.RelayPacket(packet)
	s.Require().Error(err) // expecting error as no AcknowledgePacket expected to return

	// check balance before finalization
	denom := s.getRollappToHubIBCDenomFromPacket(packet)
	transferredCoins := sdk.NewCoin(denom, coinToSendToB.Amount)
	recipient := s.hubChain().SenderAccount.GetAddress()
	initialBalance := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), recipient)
	s.Require().Equal(initialBalance.AmountOf(denom), sdk.ZeroInt())

	// Finalize the rollapp state
	currentRollappBlockHeight = uint64(s.rollappCtx().BlockHeight())
	_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
	s.Require().NoError(err)

	// manually finalize packets through x/delayedack
	s.finalizeRollappPacketsByAddress(s.hubChain().SenderAccount.GetAddress().String())

	// check balance after finalization
	expectedFee := s.hubApp().DelayedAckKeeper.BridgingFeeFromAmt(s.hubCtx(), transferredCoins.Amount)
	expectedBalance := initialBalance.Add(transferredCoins).Sub(sdk.NewCoin(denom, expectedFee))
	finalBalance := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), recipient)
	s.Equal(expectedBalance, finalBalance)

	// check fees are burned
	addr := s.hubApp().AccountKeeper.GetModuleAccount(s.hubCtx(), txfees.ModuleName)
	txFeesBalance := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), addr.GetAddress(), denom)
	s.True(txFeesBalance.IsZero())
}

func (s *bridgingFeeSuite) TestBridgingFeeReturnTokens() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	s.registerSequencer()
	s.setRollappLightClientID(s.rollappCtx().ChainID(), path.EndpointA.ClientID)

	hubEndpoint := path.EndpointA
	rollappEndpoint := path.EndpointB
	rollappIBCKeeper := s.rollappChain().App.GetIBCKeeper()

	// Update rollapp state
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight())
	s.updateRollappState(currentRollappBlockHeight)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	s.Require().True(ok)
	initialCoin := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	// First transfer: Hub -> Rollapp
	msg := types.NewMsgTransfer(
		hubEndpoint.ChannelConfig.PortID,
		hubEndpoint.ChannelID,
		initialCoin,
		s.hubChain().SenderAccount.GetAddress().String(),
		s.rollappChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		"",
	)
	res, err := s.hubChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = path.RelayPacket(packet)
	s.NoError(err)

	// Get the IBC denom on rollapp
	rollappIBCDenom := s.getRollappToHubIBCDenomFromPacket(packet)
	rollappReceivedCoin := sdk.NewCoin(rollappIBCDenom, initialCoin.Amount)
	rollappBalance := s.rollappApp().BankKeeper.GetBalance(s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), rollappIBCDenom)
	s.Equal(rollappReceivedCoin, rollappBalance)

	// Second transfer: Rollapp -> Hub (sending back the tokens)
	msg = types.NewMsgTransfer(
		rollappEndpoint.ChannelConfig.PortID,
		rollappEndpoint.ChannelID,
		rollappReceivedCoin,
		s.rollappChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		"",
	)
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	found := rollappIBCKeeper.ChannelKeeper.HasPacketCommitment(s.rollappCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	s.Require().True(found)
	err = path.RelayPacket(packet)
	s.Require().Error(err) // expecting error as no AcknowledgePacket expected to return

	// Get the final IBC denom on hub (should be the original denom)
	hubDenom := s.getRollappToHubIBCDenomFromPacket(packet)
	s.Equal(sdk.DefaultBondDenom, hubDenom) // Verify it's the original token returning

	// Check balance before finalization
	recipient := s.hubChain().SenderAccount.GetAddress()
	initialHubBalance := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), recipient)

	// Finalize the rollapp state
	currentRollappBlockHeight = uint64(s.rollappCtx().BlockHeight())
	_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
	s.Require().NoError(err)

	// Manually finalize packets through x/delayedack
	s.finalizeRollappPacketsByAddress(s.hubChain().SenderAccount.GetAddress().String())

	// Check balance after finalization
	expectedFee := s.hubApp().DelayedAckKeeper.BridgingFeeFromAmt(s.hubCtx(), rollappReceivedCoin.Amount)
	expectedBalance := initialHubBalance.Add(sdk.NewCoin(hubDenom, rollappReceivedCoin.Amount)).Sub(sdk.NewCoin(hubDenom, expectedFee))
	finalBalance := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), recipient)
	s.Equal(expectedBalance, finalBalance)

	// Check fees are burned
	addr := s.hubApp().AccountKeeper.GetModuleAccount(s.hubCtx(), txfees.ModuleName)
	txFeesBalance := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), addr.GetAddress(), hubDenom)
	s.True(txFeesBalance.IsZero())
}

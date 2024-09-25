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
	s.fundSenderAccount()
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
	s.finalizeRollappPacketsUntilHeight(currentRollappBlockHeight)

	// check balance after finalization
	expectedFee := s.hubApp().DelayedAckKeeper.BridgingFeeFromAmt(s.hubCtx(), transferredCoins.Amount)
	expectedBalance := initialBalance.Add(transferredCoins).Sub(sdk.NewCoin(denom, expectedFee))
	finalBalance := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), recipient)
	s.Equal(expectedBalance, finalBalance)

	// check fees
	addr := s.hubApp().AccountKeeper.GetModuleAccount(s.hubCtx(), txfees.ModuleName)
	txFeesBalance := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), addr.GetAddress(), denom)
	s.Equal(expectedFee, txFeesBalance.Amount)
}

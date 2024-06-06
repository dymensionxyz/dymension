package ibctesting_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/osmosis-labs/osmosis/v15/x/txfees"
	"github.com/stretchr/testify/suite"
)

type BridgingFeeTestSuite struct {
	IBCTestUtilSuite
}

func TestBridgingFeeTestSuite(t *testing.T) {
	suite.Run(t, new(BridgingFeeTestSuite))
}

func (suite *BridgingFeeTestSuite) SetupTest() {
	suite.IBCTestUtilSuite.SetupTest()
}

func (suite *BridgingFeeTestSuite) TestNotRollappNoBridgingFee() {
	// setup between cosmosChain and hubChain
	path := suite.NewTransferPath(suite.hubChain, suite.cosmosChain)
	suite.coordinator.Setup(path)
	hubEndpoint := path.EndpointA
	cosmosEndpoint := path.EndpointB

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
	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	denom := suite.GetRollappToHubIBCDenomFromPacket(packet)
	finalBalance := ConvertToApp(suite.hubChain).BankKeeper.GetBalance(suite.hubChain.GetContext(), suite.hubChain.SenderAccount.GetAddress(), denom)
	suite.Assert().Equal(sdk.NewCoin(denom, coinToSendToB.Amount), finalBalance)
}

func (suite *BridgingFeeTestSuite) TestBridgingFee() {
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)

	rollappEndpoint := path.EndpointB
	rollappIBCKeeper := suite.rollappChain.App.GetIBCKeeper()

	suite.CreateRollapp()
	suite.RegisterSequencer()
	suite.GenesisEvent(path.EndpointA.ChannelID)

	// Update rollapp state
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
	err = path.RelayPacket(packet)
	suite.Require().Error(err) // expecting error as no AcknowledgePacket expected to return

	// check balance before finalization
	denom := suite.GetRollappToHubIBCDenomFromPacket(packet)
	transferredCoins := sdk.NewCoin(denom, coinToSendToB.Amount)
	recipient := suite.hubChain.SenderAccount.GetAddress()
	initialBalance := ConvertToApp(suite.hubChain).BankKeeper.SpendableCoins(suite.hubChain.GetContext(), recipient)
	suite.Require().Equal(initialBalance.AmountOf(denom), sdk.ZeroInt())

	// Finalize the rollapp state
	currentRollappBlockHeight = uint64(suite.rollappChain.GetContext().BlockHeight())
	_, err = suite.FinalizeRollappState(1, currentRollappBlockHeight)
	suite.Require().NoError(err)

	// check balance after finalization
	expectedFee := ConvertToApp(suite.hubChain).DelayedAckKeeper.BridgingFeeFromAmt(suite.hubChain.GetContext(), transferredCoins.Amount)
	expectedBalance := initialBalance.Add(transferredCoins).Sub(sdk.NewCoin(denom, expectedFee))
	finalBalance := ConvertToApp(suite.hubChain).BankKeeper.SpendableCoins(suite.hubChain.GetContext(), recipient)
	suite.Assert().Equal(expectedBalance, finalBalance)

	// check fees
	addr := ConvertToApp(suite.hubChain).AccountKeeper.GetModuleAccount(suite.hubChain.GetContext(), txfees.ModuleName)
	txFeesBalance := ConvertToApp(suite.hubChain).BankKeeper.GetBalance(suite.hubChain.GetContext(), addr.GetAddress(), denom)
	suite.Assert().Equal(expectedFee, txFeesBalance.Amount)
}

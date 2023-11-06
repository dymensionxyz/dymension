package ibctesting_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

//transfer from cosmos -nothing
//transfer from rollapp w/o token registration - nothing
//transfer from rollapp w/ token registration - success

func (suite *KeeperTestSuite) TestDenomRegistation_RollappToHub() {
	path := suite.NewTransferPath(suite.rollappChain, suite.hubChain)
	suite.coordinator.Setup(path)

	suite.CreateRollappWithMetadata()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") //10DYM
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	msg := types.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.rollappChain.SenderAccount.GetAddress().String(), suite.hubChain.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := suite.rollappChain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	//expect error as no AcknowledgePacket expected
	suite.Require().Error(err) // relay committed

	metatdata, found := suite.hubChain.GetSimApp().BankKeeper.GetDenomMetaData(suite.hubChain.GetContext(), "utest")
	suite.Require().True(found)
	suite.Require().Equal("test", metatdata.Name)
}

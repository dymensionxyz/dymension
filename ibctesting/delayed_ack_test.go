package ibctesting_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

// Transfer from cosmos chain to the hub. No delay expected
func (suite *KeeperTestSuite) TestTransferCosmosToHub() {
	// setup between cosmosChain and hubChain
	path := suite.NewTransferPath(suite.cosmosChain, suite.hubChain)
	suite.coordinator.Setup(path)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") //10DYM
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	// send from cosmosChain to hubChain
	msg := types.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.cosmosChain.SenderAccount.GetAddress().String(), suite.hubChain.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := suite.cosmosChain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *KeeperTestSuite) TestTransferRollappToHub() {
	path := suite.NewTransferPath(suite.rollappChain, suite.hubChain)
	suite.coordinator.Setup(path)

	suite.CreateRollapp()

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
	//expeting error as no AcknowledgePacket expected
	suite.Require().Error(err) // relay committed

	stateInfo := &rollapptypes.StateInfo{
		StateInfoIndex: rollapptypes.StateInfoIndex{},
		Sequencer:      "",
		StartHeight:    0,
		NumBlocks:      0,
		DAPath:         "",
		Version:        0,
		CreationHeight: 0,
		Status:         0,
		BDs:            rollapptypes.BlockDescriptors{},
	}

	err = ConvertToApp(suite.hubChain).RollappKeeper.GetHooks().AfterStateFinalized(
		suite.hubChain.GetContext(),
		suite.rollappChain.ChainID,
		stateInfo,
	)
	suite.Require().NoError(err)
}

//TODO:
// transfer from rollapp to hub - check delay is finialized eventually
// timeout??

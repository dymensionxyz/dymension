package ibctesting_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

type DenomMetaDataTestSuite struct {
	IBCTestUtilSuite
	ctx sdk.Context
}

func TestDenomMetaDataTestSuite(t *testing.T) {
	suite.Run(t, new(DenomMetaDataTestSuite))
}

func (suite *DenomMetaDataTestSuite) SetupTest() {
	suite.IBCTestUtilSuite.SetupTest()
}

// TestDenomRegistationHubToRollapp tests the following scenario:
// 1. transfer from cosmos -nothing
// 2. transfer from rollapp w/o token registration - nothing
// 3. transfer from rollapp w/ token registration - success
// 4. same token not for the first time - no double registration
func (suite *DenomMetaDataTestSuite) TestDenomRegistationRollappToHub() {
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)

	//register rollapp with metadata for stake denom
	suite.CreateRollappWithMetadata(sdk.DefaultBondDenom)
	suite.RegisterSequencer()
	// Finalize the rollapp 100 blocks later so all packets are received immediately
	currentRollappBlockHeight := uint64(suite.rollappChain.GetContext().BlockHeight())
	suite.UpdateRollappState(1, currentRollappBlockHeight)
	suite.FinalizeRollappState(1, currentRollappBlockHeight+100)

	found := ConvertToApp(suite.hubChain).BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), sdk.DefaultBondDenom)
	suite.Require().False(found)
	found = ConvertToApp(suite.hubChain).BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), "udym")
	suite.Require().False(found)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") //10DYM
	suite.Require().True(ok)

	/* ------------------- move non-registered token from rollapp ------------------- */
	dymTokensToSend := sdk.NewCoin("udym", amount)
	apptesting.FundAccount(ConvertToApp(suite.rollappChain), suite.rollappChain.GetContext(), suite.rollappChain.SenderAccount.GetAddress(), sdk.Coins{dymTokensToSend})
	msg := types.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, dymTokensToSend, suite.rollappChain.SenderAccount.GetAddress().String(), suite.hubChain.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := suite.rollappChain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	suite.Require().NoError(err)

	udymVoucherDenom := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), "udym"))
	stakeVoucherDenom := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom))

	found = ConvertToApp(suite.hubChain).BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), sdk.DefaultBondDenom)
	suite.Require().False(found)
	found = ConvertToApp(suite.hubChain).BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), "udym")
	suite.Require().False(found)
	found = ConvertToApp(suite.hubChain).BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), udymVoucherDenom.IBCDenom())
	suite.Require().False(found)
	found = ConvertToApp(suite.hubChain).BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), stakeVoucherDenom.IBCDenom())
	suite.Require().False(found)

	/* --------------------- move native token from rollapp --------------------- */
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	msg = types.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, coinToSendToB, suite.rollappChain.SenderAccount.GetAddress().String(), suite.hubChain.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err = suite.rollappChain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	suite.Require().NoError(err)

	metadata, found := ConvertToApp(suite.hubChain).BankKeeper.GetDenomMetaData(suite.hubChain.GetContext(), stakeVoucherDenom.IBCDenom())
	suite.Require().True(found)
	suite.Equal("bigstake", metadata.Display)
	suite.Equal("BIGSTAKE", metadata.Symbol)
}

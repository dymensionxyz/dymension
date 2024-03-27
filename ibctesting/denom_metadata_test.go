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
}

func TestDenomMetaDataTestSuite(t *testing.T) {
	suite.Run(t, new(DenomMetaDataTestSuite))
}

func (suite *DenomMetaDataTestSuite) SetupTest() {
	suite.IBCTestUtilSuite.SetupTest()
}

func (suite *DenomMetaDataTestSuite) TestDenomRegistationRollappToHub() {
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)

	// register rollapp with metadata for stake denom
	suite.CreateRollappWithMetadata(sdk.DefaultBondDenom)
	suite.RegisterSequencer()

	app := ConvertToApp(suite.hubChain)

	// invoke genesis event, in order to register denoms
	suite.GenesisEvent(path.EndpointB.Chain.ChainID, path.EndpointA.ChannelID)

	// Finalize the rollapp 100 blocks later so all packets are received immediately
	currentRollappBlockHeight := uint64(suite.rollappChain.GetContext().BlockHeight())
	suite.UpdateRollappState(currentRollappBlockHeight)
	err := suite.FinalizeRollappState(1, currentRollappBlockHeight+100)
	suite.Require().NoError(err)

	found := app.BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), sdk.DefaultBondDenom)
	suite.Require().False(found)
	found = app.BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), "udym")
	suite.Require().False(found)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
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

	found = app.BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), sdk.DefaultBondDenom)
	suite.Require().False(found)
	found = app.BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), "udym")
	suite.Require().False(found)
	found = app.BankKeeper.HasDenomMetaData(suite.hubChain.GetContext(), udymVoucherDenom.IBCDenom())
	suite.Require().False(found)
	metadata, found := app.BankKeeper.GetDenomMetaData(suite.hubChain.GetContext(), stakeVoucherDenom.IBCDenom())
	suite.Require().True(found)
	suite.Equal("bigstake", metadata.Display)
	suite.Equal("BIGSTAKE", metadata.Symbol)
}

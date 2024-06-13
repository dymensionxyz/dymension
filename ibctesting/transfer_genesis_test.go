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

type transferGenesisSuite struct {
	utilSuite
}

func TestTransferGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(transferGenesisSuite))
}

func (s *transferGenesisSuite) SetupTest() {
	s.utilSuite.SetupTest()
}

// In the happy path, the new rollapp will send ibc transfers with a special
// memo immediately when the channel opens. This will cause  all the denoms to get registered, and tokens
// to go to the right addresses.
func (s *transferGenesisSuite) TestHappyPath() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	s.createRollapp(false, nil) // genesis protocol is not finished yet
	s.registerSequencer()

	// Finalize the rollapp 100 blocks later so all packets are received immediately
	currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight())
	s.updateRollappState(currentRollappBlockHeight)
	_, err := s.finalizeRollappState(1, currentRollappBlockHeight+100)
	s.Require().NoError(err)

	found := s.hubApp().BankKeeper.HasDenomMetaData(s.hubCtx(), sdk.DefaultBondDenom)
	s.Require().False(found)
	found = s.hubApp().BankKeeper.HasDenomMetaData(s.hubCtx(), "udym")
	s.Require().False(found)

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := sdk.NewIntFromString("10000000000000000000") // 10DYM
	s.Require().True(ok)

	/* ------------------- move non-registered token from rollapp ------------------- */
	dymTokensToSend := sdk.NewCoin("udym", amount)
	apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.Coins{dymTokensToSend})
	msg := types.NewMsgTransfer(
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		dymTokensToSend,
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

	// relay send
	err = path.RelayPacket(packet)
	s.Require().NoError(err)

	udymVoucherDenom := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), "udym"))
	stakeVoucherDenom := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom))

	found = s.hubApp().BankKeeper.HasDenomMetaData(s.hubCtx(), sdk.DefaultBondDenom)
	s.Require().False(found)
	found = s.hubApp().BankKeeper.HasDenomMetaData(s.hubCtx(), "udym")
	s.Require().False(found)
	found = s.hubApp().BankKeeper.HasDenomMetaData(s.hubCtx(), udymVoucherDenom.IBCDenom())
	s.Require().False(found)
	metadata, found := s.hubApp().BankKeeper.GetDenomMetaData(s.hubCtx(), stakeVoucherDenom.IBCDenom())
	s.Require().True(found, "missing denom metadata for rollapps taking token")
	s.Equal("bigstake", metadata.Display)
	s.Equal("BIGSTAKE", metadata.Symbol)
}

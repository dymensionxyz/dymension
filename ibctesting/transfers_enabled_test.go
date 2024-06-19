package ibctesting_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/cosmos/ibc-go/v6/testing/simapp"
)

type transfersEnabledSuite struct {
	utilSuite

	path *ibctesting.Path
}

func TestTransfersEnabledTestSuite(t *testing.T) {
	suite.Run(t, new(transfersEnabledSuite))
}

func (s *transfersEnabledSuite) SetupTest() {
	s.utilSuite.SetupTest()
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	s.createRollapp(false, nil)
	s.registerSequencer()
	s.path = path
	s.Require().False(s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled)
}

// Regular (non genesis) transfers RA->Hub (and Hub->RA) should both be blocked when the bridge is not open
// Note: we don't explicitly test the 'enabled' case, since this is tested in other tests in this package
func (s *transfersEnabledSuite) TestRollappToHubDisabled() {
	amt := math.NewIntFromUint64(10000000000000000000)
	denom := "foo"
	tokens := sdk.NewCoin(denom, amt)

	timeoutHeight := clienttypes.NewHeight(100, 110)

	msg := types.NewMsgTransfer(
		s.path.EndpointB.ChannelConfig.PortID,
		s.path.EndpointB.ChannelID,
		tokens,
		s.rollappChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		"",
	)

	receiverBalance := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), denom)
	s.Require().True(receiverBalance.Amount.IsZero())

	apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.Coins{msg.Token})
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// money is escrowed
	senderBalance := s.rollappApp().BankKeeper.GetBalance(s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), denom)
	s.Require().True(senderBalance.Amount.IsZero())

	err = s.path.RelayPacket(packet)
	s.Require().Error(err) // error means ack was delayed, which means the transfer was accepted

	// money is refunded
	senderBalance = s.rollappApp().BankKeeper.GetBalance(s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), denom)
	s.Require().True(senderBalance.Amount.IsPositive())

	// no double spend
	receiverBalance = s.hubApp().BankKeeper.GetBalance(s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), denom)
	s.Require().True(receiverBalance.Amount.IsZero())
}

// Regular (non genesis) transfers (RA->Hub) and Hub->RA should both be blocked when the bridge is not open
func (s *transfersEnabledSuite) TestHubToRollappDisabled() {
	amt := math.NewIntFromUint64(10000000000000000000)
	denom := "foo"
	tokens := sdk.NewCoin(denom, amt)

	timeoutHeight := clienttypes.NewHeight(100, 110)

	msg := types.NewMsgTransfer(
		s.path.EndpointA.ChannelConfig.PortID,
		s.path.EndpointA.ChannelID,
		tokens,
		s.hubChain().SenderAccount.GetAddress().String(),
		s.rollappChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		"",
	)

	shouldFail := true

	for range 2 {

		apptesting.FundAccount(s.hubApp(), s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), sdk.Coins{msg.Token})

		_, _, err := simapp.SignAndDeliver(
			s.hubChain().T,
			s.hubChain().TxConfig,
			s.hubApp().GetBaseApp(),
			s.hubCtx().BlockHeader(),
			[]sdk.Msg{msg},
			hubChainID(),
			[]uint64{s.hubChain().SenderAccount.GetAccountNumber()},
			[]uint64{s.hubChain().SenderAccount.GetSequence()},
			true,
			!shouldFail,
			s.hubChain().SenderPrivKey,
		)

		if shouldFail {
			shouldFail = false
			s.Require().True(errorsmod.IsOf(err, gerr.ErrFailedPrecondition))
			ra := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
			ra.ChannelId = s.path.EndpointA.ChannelID
			s.hubApp().RollappKeeper.SetRollapp(s.hubCtx(), ra)
			s.hubApp().RollappKeeper.EnableTransfers(s.hubCtx(), ra.RollappId)
		} else {
			s.Require().NoError(err)
		}
	}
}

package ibctesting

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/stretchr/testify/suite"
	"testing"
)

package ibctesting_test

import (
"testing"

"cosmossdk.io/math"

banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

sdk "github.com/cosmos/cosmos-sdk/types"
"github.com/dymensionxyz/dymension/v3/app/apptesting"
"github.com/stretchr/testify/suite"

"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
ibctesting "github.com/cosmos/ibc-go/v6/testing"
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


// Regular (non genesis) transfers RA->Hub and Hub->RA should both be blocked when the bridge is not open
func (s *transfersEnabledSuite) TestTransferRAtoHubIsDisabled() {

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
	s.Require().NoError(err)

	// money is refunded
	senderBalance = s.rollappApp().BankKeeper.GetBalance(s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), denom)
	s.Require().True(senderBalance.Amount.IsPositive())

	// no double spend
	receiverBalance = s.hubApp().BankKeeper.GetBalance(s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), denom)
	s.Require().True(receiverBalance.Amount.IsZero())
}


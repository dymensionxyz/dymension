package ibctesting_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
)

type transfersEnabledSuite struct {
	ibcTestingSuite
	path *ibctesting.Path
}

func TestTransfersEnabledTestSuite(t *testing.T) {
	suite.Run(t, new(transfersEnabledSuite))
}

func (s *transfersEnabledSuite) SetupTest() {
	s.ibcTestingSuite.SetupTest()
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	s.createRollapp(false, nil)
	s.registerSequencer()
	s.path = path

	// force set the canonical client for the rollapp
	s.setRollappLightClientID(rollappChainID(), s.path.EndpointA.ClientID)
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

		_, err := s.hubChain().SendMsgs([]sdk.Msg{msg}...)
		if shouldFail {
			shouldFail = false
			s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
			ra := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
			ra.ChannelId = s.path.EndpointA.ChannelID
			ra.GenesisState.TransferProofHeight = 1 // enable
			s.hubApp().RollappKeeper.SetRollapp(s.hubCtx(), ra)
		} else {
			s.Require().NoError(err)
		}
	}
}

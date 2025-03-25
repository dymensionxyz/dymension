package ibctesting_test

import (
	"testing"

	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/cosmos/gogoproto/proto"
	forwardtypes "github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/stretchr/testify/suite"
)

/*
Friday:
Got the basic hook structure up:
A memo can contain a hook name and data
It will end up in x/forward
which is connected to the warp route keeper, so it can initiate sends

Next sensible things:
- Need to ideally have a test that takes a transfer with the memo, and ensures the forwarder gets called
 (can use a dummy hook)
-
*/

type forwardSuite struct {
	eibcSuite
}

func TestForwardSuite(t *testing.T) {
	suite.Run(t, new(forwardSuite))
}

func (s *forwardSuite) SetupTest() {
	s.eibcSuite.SetupTest()
}

func (s *forwardSuite) TestForward() {
	s.T().Log("running test forward!")
	basicHook := forwardtypes.HookCalldata{
		HyperlaneTransfer: &warptypes.MsgRemoteTransfer{
			TokenId:           "1",
			DestinationDomain: "1",
			Sender:            "cosmos1",
		},
	}
	basicHookBytes, err := proto.Marshal(&basicHook)
	s.Require().NoError(err)
	s.eibcTransferFulfillment([]eibcTransferFulfillmentTC{
		{
			name:              "forwarding works",
			fulfillerStartBal: "300",
			eibcTransferFee:   "150",
			transferAmt:       "200",
			fulfillHook:       basicHookBytes,
		},
	})
}

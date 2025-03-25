package ibctesting_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
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

// TODO: use mock generation
type mockHook struct {
	called bool
}

func (h mockHook) ValidateData(hookData []byte) error {
	return nil
}

func (h mockHook) Run(ctx sdk.Context, order *types.DemandOrder,
	fundsSource sdk.AccAddress,
	newTransferRecipient sdk.AccAddress,
	fulfiller sdk.AccAddress,
	hookData []byte) error {
	h.called = true
	return nil
}

func (s *forwardSuite) TestForward() {
	dummy := "dummy"
	h := mockHook{}
	s.utilSuite.hubApp().EIBCKeeper.SetFulfillHooks(
		map[string]eibckeeper.FulfillHook{
			dummy: h,
		},
	)
	s.T().Log("running test forward!")

	hookData := eibctypes.FulfillHook{
		HookName: dummy,
		HookData: []byte{},
	}
	raw, err := proto.Marshal(&hookData)
	s.Require().NoError(err)
	s.eibcTransferFulfillment([]eibcTransferFulfillmentTC{
		{
			name:              "forwarding works",
			fulfillerStartBal: "300",
			eibcTransferFee:   "150",
			transferAmt:       "200",
			fulfillHook:       raw,
		},
	})
	s.Require().True(h.called)
}

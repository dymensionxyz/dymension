package ibctesting_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfer "github.com/dymensionxyz/dymension/v3/x/transfer"
	transfertypes "github.com/dymensionxyz/dymension/v3/x/transfer/types"
	"github.com/stretchr/testify/suite"
)

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
	*forwardSuite
	called bool
}

func (h *mockHook) ValidateData(hookData []byte) error {
	return nil
}

func (h *mockHook) Run(ctx sdk.Context, fundsSource sdk.AccAddress, budget sdk.Coin, hookData []byte) error {

	h.called = true
	return nil
}

func (s *forwardSuite) TestForward() {
	dummy := "dummy"
	h := mockHook{
		forwardSuite: s,
	}
	s.utilSuite.hubApp().TransferHooks.SetHooks(
		map[string]transfer.CompletionHookInstance{
			dummy: &h,
		},
	)
	s.T().Log("running test forward!")

	hookData := transfertypes.CompletionHookCall{
		Name: dummy,
		Data: []byte{},
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

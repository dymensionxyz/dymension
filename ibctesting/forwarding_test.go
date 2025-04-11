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

func (h *mockHook) Run(ctx sdk.Context, order *types.DemandOrder,
	fundsSource sdk.AccAddress,
	newTransferRecipient sdk.AccAddress,
	fulfiller sdk.AccAddress,
	hookData []byte) error {

	// mimic the regular transfer for this test
	h.utilSuite.hubApp().BankKeeper.SendCoins(ctx, fundsSource, order.GetRecipientBech32Address(), order.Price)

	h.called = true
	return nil
}

func (s *forwardSuite) TestForward() {
	dummy := "dummy"
	h := mockHook{
		forwardSuite: s,
	}
	s.utilSuite.hubApp().EIBCKeeper.SetFulfillHooks(
		map[string]eibckeeper.FulfillHook{
			dummy: &h,
		},
	)
	s.T().Log("running test forward!")

	hookData := eibctypes.OnFulfillHook{
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
